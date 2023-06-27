package logger

import (
	"strings"

	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	DebugLevel = "debug"
	InfoLevel  = "info"
	WarnLevel  = "warn"
	ErrorLevel = "error"
	FatalLevel = "fatal"
)

type (
	// logger is logger backed up by zap.Logger.
	logger struct {
		zap *zap.SugaredLogger
	}

	// Config contains the config items for logger
	Config struct {
		// Level is the desired log level
		Level string
	}
)

var (
	_ log.Logger     = (*logger)(nil)
	_ log.WithLogger = &logger{}
)

// NewDefaultLogger returns a logger at debug level and log into STDOUT
func NewDefaultLogger(logLevel string) logger {
	return NewZapLogger(BuildZapLogger(Config{Level: logLevel}))
}

// NewZapLogger returns a new zap based logger from zap.Logger
func NewZapLogger(zl *zap.SugaredLogger) logger {
	return logger{zap: zl}
}

// BuildZapLogger builds and returns a new zap.Logger for this logging configuration
func BuildZapLogger(cfg Config) *zap.SugaredLogger {
	return buildZapLogger(cfg, true)
}

func (l logger) Debug(msg string, keyvals ...interface{}) {
	l.zap.With(keyvals...).Debug(msg)
}

func (l logger) Info(msg string, keyvals ...interface{}) {
	l.zap.With(keyvals...).Info(msg)
}

func (l logger) Warn(msg string, keyvals ...interface{}) {
	l.zap.With(keyvals...).Warn(msg)
}

func (l logger) Error(msg string, keyvals ...interface{}) {
	l.zap.With(keyvals...).Error(msg)
}

func (l logger) Fatal(msg string, keyvals ...interface{}) {
	l.zap.With(keyvals...).Fatal(msg)
}

func (l logger) With(keyvals ...interface{}) log.Logger {
	zsl := l.zap.With(keyvals...)
	return NewZapLogger(zsl)
}

func buildZapLogger(cfg Config, disableCaller bool) *zap.SugaredLogger {
	encodeConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      zapcore.OmitKey, // we use our own caller
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	if disableCaller {
		encodeConfig.CallerKey = zapcore.OmitKey
		encodeConfig.EncodeCaller = nil
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(parseZapLevel(cfg.Level)),
		Development:      false,
		Sampling:         nil,
		Encoding:         "json",
		EncoderConfig:    encodeConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    disableCaller,
	}
	logger, _ := config.Build()
	return logger.Sugar()
}

func parseZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case DebugLevel:
		return zap.DebugLevel
	case InfoLevel:
		return zap.InfoLevel
	case WarnLevel:
		return zap.WarnLevel
	case ErrorLevel:
		return zap.ErrorLevel
	case FatalLevel:
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
