package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/meraf00/swytch/core/lib/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client is the base struct for handling connection recovery, consumption and
// publishing.
type Client struct {
	m               *sync.Mutex
	log             logger.Log
	connection      *amqp.Connection
	notifyConnClose chan *amqp.Error
	queues          map[string]bool
	channels        map[string]*amqp.Channel
	channelReady    map[string]chan struct{}
	notifyConfirm   map[string]chan amqp.Confirmation
	notifyChanClose map[string]chan *amqp.Error
	done            chan bool
	isReady         bool
	IsReady         chan struct{}
}

const (
	maxReconnectAttempts = 15
	reconnectDelay       = 5 * time.Second
	reInitDelay          = 2 * time.Second
	resendDelay          = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("client is shutting down")
)

// New creates a new consumer state instance, and automatically
// attempts to connect to the server.
func New(addr string, log logger.Log) *Client {
	client := Client{
		m:               &sync.Mutex{},
		log:             log,
		queues:          make(map[string]bool),
		channels:        make(map[string]*amqp.Channel),
		channelReady:    make(map[string]chan struct{}),
		notifyConfirm:   make(map[string]chan amqp.Confirmation),
		notifyChanClose: make(map[string]chan *amqp.Error),
		done:            make(chan bool),
		IsReady:         make(chan struct{}, 1),
	}
	go client.handleReconnect(addr)
	return &client
}

// EnsureQueue ensures a queue exists and is ready for use.
func (client *Client) EnsureQueue(queue string) error {
	client.m.Lock()
	if open, exists := client.queues[queue]; open && exists {
		client.m.Unlock()
		return nil
	}

	client.queues[queue] = false
	readyChan := make(chan struct{}, 1)
	client.channelReady[queue] = readyChan
	client.m.Unlock()

	conn := client.Connection()
	if conn == nil {
		return errNotConnected
	}

	client.handleReInit(conn, queue)

	select {
	case <-readyChan:
		return nil
	case <-client.done:
		return errShutdown
	}
}

func (client *Client) Queues() map[string]bool {
	client.m.Lock()
	defer client.m.Unlock()
	return client.queues
}

// Connection returns the current AMQP connection.
func (client *Client) Connection() *amqp.Connection {
	client.m.Lock()
	defer client.m.Unlock()
	return client.connection
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *Client) handleReconnect(addr string) {
	reconnectAttempts := 0

	for {
		select {
		case <-client.done:
			return
		default:
		}

		if reconnectAttempts >= maxReconnectAttempts {
			client.log.Error("Max reconnect attempts reached. Giving up.")
			return
		}

		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		client.log.Info("Attempting to connect")

		_, err := client.connect(addr)

		if err != nil {
			if (err == amqp.ErrCredentials) || (err == amqp.ErrSASL) {
				client.log.Error("Invalid credentials.")
				close(client.done)
				return
			}

			client.log.Errorf("Failed to connect: %v", err)
			client.log.Warn("Failed to connect. Retrying...")

			reconnectAttempts++
			time.Sleep(reconnectDelay)

			continue
		}

		reconnectAttempts = 0

		queues := client.Queues()
		fmt.Printf("%v\n", queues)

		for queue := range queues {
			err := client.EnsureQueue(queue)
			if err != nil {
				client.log.Errorf("Failed to ensure queue %s: %v", queue, err)
			}
		}

		select {
		case <-client.done:
			return
		case err := <-client.notifyConnClose:
			client.resetChannels()
			client.log.Infof("Connection closed: %v", err)
		}
	}
}

// connect will create a new AMQP connection
func (client *Client) connect(addr string) (*amqp.Connection, error) {
	addrURL, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	client.changeConnection(conn)

	client.m.Lock()
	client.isReady = true
	client.m.Unlock()

	select {
	case client.IsReady <- struct{}{}:
	default:
	}

	client.log.Info("Connected to RabbitMQ at ", addrURL.Redacted())
	return conn, nil
}

func (client *Client) resetChannels() {
	client.m.Lock()
	defer client.m.Unlock()
	for queue := range client.channels {
		client.queues[queue] = false
	}
}

// handleReInit will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (client *Client) handleReInit(conn *amqp.Connection, queue string) bool {
	for {
		client.m.Lock()
		client.queues[queue] = false
		client.m.Unlock()

		err := client.init(conn, queue)

		if err != nil {
			client.log.Info("Failed to initialize channel. Retrying...")

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				client.resetChannels()
				client.log.Info("Connection closed. Reconnecting...")
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		client.m.Lock()
		notifyChanClose := client.notifyChanClose[queue]
		client.m.Unlock()

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			client.resetChannels()
			client.log.Info("Connection closed. Reconnecting...")

			return false
		case <-notifyChanClose:
			client.m.Lock()
			client.queues[queue] = false
			client.m.Unlock()
			client.log.Info("Channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (client *Client) init(conn *amqp.Connection, queue string) error {
	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}
	_, err = ch.QueueDeclare(
		queue,
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return err
	}

	client.changeChannel(ch, queue)

	client.m.Lock()
	client.queues[queue] = true
	readyChan := client.channelReady[queue]
	client.m.Unlock()

	select {
	case readyChan <- struct{}{}:
	default:
	}

	client.log.Info("Setup queue: ", queue)
	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (client *Client) changeConnection(connection *amqp.Connection) {
	client.m.Lock()
	defer client.m.Unlock()
	client.connection = connection
	client.notifyConnClose = make(chan *amqp.Error, 1)
	client.connection.NotifyClose(client.notifyConnClose)
}

// changeChannel takes a new channel to the queue,
// and updates the channel listeners to reflect this.
func (client *Client) changeChannel(channel *amqp.Channel, queue string) {
	client.m.Lock()
	defer client.m.Unlock()
	client.channels[queue] = channel
	client.notifyChanClose[queue] = make(chan *amqp.Error, 1)
	client.notifyConfirm[queue] = make(chan amqp.Confirmation, 1)
	channel.NotifyClose(client.notifyChanClose[queue])
	channel.NotifyPublish(client.notifyConfirm[queue])
}

// Push will push data onto the queue, and wait for a confirmation.
// This will block until the server sends a confirmation. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (client *Client) Push(queue string, data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errors.New("failed to push: not connected")
	}
	client.m.Unlock()

	confirmChannel, ok := client.notifyConfirm[queue]
	if !ok {
		return errors.New("confirmation channel not found for queue: " + queue)
	}

	for {
		err := client.UnsafePush(queue, data)
		if err != nil {
			client.log.Info("Push failed. Retrying...")
			select {
			case <-client.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}

		confirm := <-confirmChannel
		if confirm.Ack {
			client.log.Infof("Push confirmed [%d]!", confirm.DeliveryTag)
			return nil
		}
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (client *Client) UnsafePush(queue string, data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errNotConnected
	}
	client.m.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	channel, ok := client.channels[queue]
	if !ok {
		return errors.New("channel not found for queue: " + queue)
	}

	return channel.PublishWithContext(
		ctx,
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}

// Consume will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (client *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return nil, errNotConnected
	}

	channel, ok := client.channels[queue]
	if !ok {
		client.m.Unlock()
		return nil, errors.New("channel not found for queue: " + queue)
	}
	client.m.Unlock()

	if err := channel.Qos(
		1,
		0,
		false,
	); err != nil {
		return nil, err
	}

	return channel.Consume(
		queue,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
}

// Close will cleanly shut down the channel and connection.
func (client *Client) Close() error {
	client.m.Lock()
	defer client.m.Unlock()

	if !client.isReady {
		return errAlreadyClosed
	}
	close(client.done)

	for queue, channel := range client.channels {
		err := channel.Close()
		if err != nil {
			client.log.Errorf("Failed to close channel for queue %s: %v", queue, err)
		}
	}
	err := client.connection.Close()
	if err != nil {
		return err
	}

	client.isReady = false
	return nil
}

func (client *Client) NotifyClose(queue string, c chan *amqp.Error) chan *amqp.Error {
	client.m.Lock()
	defer client.m.Unlock()
	channel, ok := client.channels[queue]
	if !ok {
		return nil
	}
	return channel.NotifyClose(c)
}
