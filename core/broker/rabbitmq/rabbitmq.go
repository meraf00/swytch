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
	m                  *sync.Mutex
	logger             logger.Log
	connection         *amqp.Connection
	notifyConnClose    chan *amqp.Error
	done               chan bool
	channels           map[string]*amqp.Channel
	notifyChanClose    map[string]chan *amqp.Error
	aggNotifyChanClose chan struct {
		*amqp.Error
		queue string
	}
	notifyConfirm map[string]chan amqp.Confirmation
	isReady       bool
	IsReady       chan struct{}
}

const (
	reconnectDelay = 5 * time.Second

	reInitDelay = 2 * time.Second

	resendDelay = 5 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to a server")
	errAlreadyClosed = errors.New("already closed: not connected to the server")
	errShutdown      = errors.New("client is shutting down")
)

// New creates a new consumer state instance, and automatically
// attempts to connect to the server.
func New(addr string, logger logger.Log) *Client {
	client := Client{
		m:               &sync.Mutex{},
		logger:          logger,
		done:            make(chan bool),
		IsReady:         make(chan struct{}, 1),
		channels:        make(map[string]*amqp.Channel),
		notifyChanClose: make(map[string]chan *amqp.Error),
		notifyConfirm:   make(map[string]chan amqp.Confirmation),
	}
	go client.handleReconnect(addr)
	return &client
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (client *Client) handleReconnect(addr string) {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		client.logger.Info("Attempting to connect")

		conn, err := client.connect(addr)

		if err != nil {
			if (err == amqp.ErrCredentials) || (err == amqp.ErrSASL) {
				client.logger.Fatalf("Invalid credentials")
				close(client.done)
				return
			}

			client.logger.Info("Failed to connect. Retrying...")

			select {
			case <-client.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := client.handleReInit(conn); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (client *Client) connect(addr string) (*amqp.Connection, error) {
	brokerAddr, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, err
	}

	client.changeConnection(conn)
	client.logger.Infof("Connected to %s", brokerAddr.Redacted())
	return conn, nil
}

// handleReInit will wait for a channel error
// and then continuously attempt to re-initialize both channels
func (client *Client) handleReInit(conn *amqp.Connection) bool {
	for {
		client.m.Lock()
		client.isReady = false
		client.m.Unlock()

		err := client.init(conn)

		if err != nil {
			client.logger.Info("Failed to initialize channel. Retrying...")

			select {
			case <-client.done:
				return true
			case <-client.notifyConnClose:
				client.logger.Info("Connection closed. Reconnecting...")
				return false
			case <-time.After(reInitDelay):
			}
			continue
		}

		for queue := range client.channels {
			go func() {
				err := <-client.notifyChanClose[queue]
				if err != nil {
					client.logger.Infof("Channel for queue %s closed. Re-running init...", queue)
					client.aggNotifyChanClose <- struct {
						*amqp.Error
						queue string
					}{err, queue}

				}
				close(client.aggNotifyChanClose)
			}()
		}

		select {
		case <-client.done:
			return true
		case <-client.notifyConnClose:
			client.logger.Info("Connection closed. Reconnecting...")
			return false
		case <-client.aggNotifyChanClose:
			client.logger.Info("Channel closed. Re-running init...")
		}
	}
}

// init will initialize channel & declare queue
func (client *Client) init(conn *amqp.Connection) error {
	channels := make(map[string]*amqp.Channel, len(client.channels))

	for queue := range client.channels {
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
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}
		channels[queue] = ch
	}

	client.changeChannels(channels)
	client.m.Lock()
	client.isReady = true
	select {
	case client.IsReady <- struct{}{}:
	default:
	}
	client.m.Unlock()

	return nil
}

func (client *Client) AddQueue(queue string) error {
	client.m.Lock()
	defer client.m.Unlock()

	if _, exists := client.channels[queue]; exists {
		return nil
	}

	ch, err := client.connection.Channel()
	if err != nil {
		return err
	}

	err = ch.Confirm(false)
	if err != nil {
		return err
	}

	client.channels[queue] = ch

	_, err = ch.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	client.notifyChanClose[queue] = make(chan *amqp.Error, 1)
	client.notifyConfirm[queue] = make(chan amqp.Confirmation, 1)
	ch.NotifyClose(client.notifyChanClose[queue])
	ch.NotifyPublish(client.notifyConfirm[queue])

	go func() {
		err := <-client.notifyChanClose[queue]
		if err != nil {
			client.logger.Infof("Channel for queue %s closed. Re-running init...", queue)
			client.aggNotifyChanClose <- struct {
				*amqp.Error
				queue string
			}{err, queue}

		}
		close(client.aggNotifyChanClose)
	}()

	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (client *Client) changeConnection(connection *amqp.Connection) {
	client.connection = connection
	client.notifyConnClose = make(chan *amqp.Error, 1)
	client.connection.NotifyClose(client.notifyConnClose)
}

// changeChannel takes a new channels to the queues,
// and updates the channel listeners to reflect this.
func (client *Client) changeChannels(channels map[string]*amqp.Channel) {
	client.channels = channels
	for queue := range client.channels {
		client.notifyChanClose[queue] = make(chan *amqp.Error, 1)
		client.notifyConfirm[queue] = make(chan amqp.Confirmation, 1)
		client.channels[queue].NotifyClose(client.notifyChanClose[queue])
		client.channels[queue].NotifyPublish(client.notifyConfirm[queue])
	}
	client.aggNotifyChanClose = make(chan struct {
		*amqp.Error
		queue string
	})
}

// Publish will push data onto the queue, and wait for a confirmation.
// This will block until the server sends a confirmation. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (client *Client) Publish(exchange, queue string, data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errors.New("failed to push: not connected")
	}
	client.m.Unlock()
	for {
		err := client.PublishNoConfirm(exchange, queue, data)
		if err != nil {
			client.logger.Info("Publish failed. Retrying...")
			select {
			case <-client.done:
				return errShutdown
			case <-time.After(resendDelay):
			}
			continue
		}
		confirm := <-client.notifyConfirm[queue]
		if confirm.Ack {
			return nil
		}
	}
}

// PublishNoConfirm will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// receive the message.
func (client *Client) PublishNoConfirm(exchange, queue string, data []byte) error {
	client.m.Lock()
	if !client.isReady {
		client.m.Unlock()
		return errNotConnected
	}
	client.m.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.channels[queue].PublishWithContext(
		ctx,
		exchange,
		queue,
		false,
		false,
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
	client.m.Unlock()

	ch, ok := client.channels[queue]
	if !ok {
		return nil, fmt.Errorf("queue %s not found", queue)
	}

	if err := ch.Qos(
		1,
		0,
		false,
	); err != nil {
		return nil, err
	}

	return ch.Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

// Close will cleanly shut down the channels and connection.
func (client *Client) Close() error {
	client.m.Lock()

	defer client.m.Unlock()

	if !client.isReady {
		return errAlreadyClosed
	}
	close(client.done)

	for _, ch := range client.channels {
		err := ch.Close()
		if err != nil {
			return err
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

	ch, ok := client.channels[queue]
	if !ok {
		return nil
	}

	return ch.NotifyClose(c)
}
