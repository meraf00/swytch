package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	WithFields(fields map[string]any) Log
	Named(string) Log
	HTTP() Log // Returns a logger configured for HTTP logging
}

type log struct {
	entry      *zap.SugaredLogger
	mainLogger *zap.SugaredLogger
	httpLogger *zap.SugaredLogger
}

func NewLogger() Log {
	consoleCfg := zap.NewProductionEncoderConfig()
	consoleCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)

	// HTTP logger (JSON)
	jsonCfg := zap.NewProductionEncoderConfig()
	jsonCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	jsonEncoder := zapcore.NewJSONEncoder(jsonCfg)

	writeSyncer := zapcore.AddSync(os.Stdout)

	mainCore := zapcore.NewCore(consoleEncoder, writeSyncer, zap.InfoLevel)
	httpCore := zapcore.NewCore(jsonEncoder, writeSyncer, zap.InfoLevel)

	mainLogger := zap.New(mainCore)
	httpLogger := zap.New(httpCore)

	sugaredMainLogger := mainLogger.Sugar()
	sugaredHTTPLogger := httpLogger.Sugar()

	return &log{
		entry:      sugaredMainLogger,
		mainLogger: sugaredMainLogger,
		httpLogger: sugaredHTTPLogger,
	}
}

func (l *log) Console() Log {
	return &log{
		entry:      l.mainLogger,
		mainLogger: l.mainLogger,
		httpLogger: l.httpLogger,
	}
}

func (l *log) HTTP() Log {
	return &log{
		entry:      l.httpLogger,
		mainLogger: l.mainLogger,
		httpLogger: l.httpLogger,
	}
}

func (l *log) Debug(args ...any) { l.entry.Debug(args...) }
func (l *log) Info(args ...any)  { l.entry.Info(args...) }
func (l *log) Warn(args ...any)  { l.entry.Warn(args...) }
func (l *log) Error(args ...any) { l.entry.Error(args...) }
func (l *log) Fatal(args ...any) { l.entry.Fatal(args...) }

func (l *log) Debugf(format string, args ...any) { l.entry.Debugf(format, args...) }
func (l *log) Infof(format string, args ...any)  { l.entry.Infof(format, args...) }
func (l *log) Warnf(format string, args ...any)  { l.entry.Warnf(format, args...) }
func (l *log) Errorf(format string, args ...any) { l.entry.Errorf(format, args...) }
func (l *log) Fatalf(format string, args ...any) { l.entry.Fatalf(format, args...) }

func (l *log) WithFields(fields map[string]any) Log {
	return &log{
		entry:      l.entry.With(MapToArgs(fields)...),
		mainLogger: l.mainLogger.With(MapToArgs(fields)...),
		httpLogger: l.httpLogger.With(MapToArgs(fields)...),
	}
}

func (l *log) Named(name string) Log {
	return &log{
		entry:      l.entry.Named(name),
		mainLogger: l.mainLogger.Named(name),
		httpLogger: l.httpLogger.Named(name),
	}
}

func MapToArgs(fields map[string]any) []any {
	args := make([]any, 0, len(fields)*2)

	for k, v := range fields {
		args = append(args, k, v)
	}

	return args
}

func MapToZapFields(fields map[string]any) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			zapFields = append(zapFields, zap.String(k, val))
		case int:
			zapFields = append(zapFields, zap.Int(k, val))
		case int8:
			zapFields = append(zapFields, zap.Int8(k, val))
		case int16:
			zapFields = append(zapFields, zap.Int16(k, val))
		case int32:
			zapFields = append(zapFields, zap.Int32(k, val))
		case int64:
			zapFields = append(zapFields, zap.Int64(k, val))
		case uint:
			zapFields = append(zapFields, zap.Uint(k, val))
		case uint8:
			zapFields = append(zapFields, zap.Uint8(k, val))
		case uint16:
			zapFields = append(zapFields, zap.Uint16(k, val))
		case uint32:
			zapFields = append(zapFields, zap.Uint32(k, val))
		case uint64:
			zapFields = append(zapFields, zap.Uint64(k, val))
		case float32:
			zapFields = append(zapFields, zap.Float32(k, val))
		case float64:
			zapFields = append(zapFields, zap.Float64(k, val))
		case bool:
			zapFields = append(zapFields, zap.Bool(k, val))
		default:
			zapFields = append(zapFields, zap.Any(k, val))
		}
	}
	return zapFields
}
