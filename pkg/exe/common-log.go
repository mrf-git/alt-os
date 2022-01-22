package exe

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Fields represents a map of fields to log along with the next
// call to a log function.
type Fields map[string]interface{}

// Logger represents an object that can write logs at various
// log levels. Assumed to be concurrency-safe.
type Logger interface {
	WithFields(Fields) Logger
	Trace(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
}

// LoggerConf holds configuration parameters for a logger.
type LoggerConf struct {
	Enabled    bool
	Level      string
	FormatJson bool
	ExeTag     string
}

type _Logger struct {
	log     *logrus.Entry
	enabled bool
}

func (sl _Logger) WithFields(fields Fields) Logger {
	if sl.enabled {
		return _Logger{log: sl.log.WithFields(logrus.Fields(fields)), enabled: true}
	}
	return sl
}

func (sl _Logger) Trace(args ...interface{}) {
	if sl.enabled {
		sl.log.Trace(args)
	}
}

func (sl _Logger) Info(args ...interface{}) {
	if sl.enabled {
		sl.log.Info(args)
	}
}

func (sl _Logger) Warn(args ...interface{}) {
	if sl.enabled {
		sl.log.Warn(args)
	}
}

func (sl _Logger) Error(args ...interface{}) {
	if sl.enabled {
		sl.log.Error(args)
	}
}

func NewLogger(conf *LoggerConf) Logger {
	isEnabled := conf.Enabled
	var logLevel logrus.Level
	switch strings.ToLower(conf.Level) {
	case "trace":
		logLevel = logrus.TraceLevel
	case "info":
		logLevel = logrus.InfoLevel
	case "warn":
		logLevel = logrus.WarnLevel
	case "error":
		logLevel = logrus.ErrorLevel
	default:
		isEnabled = false
	}
	if !isEnabled {
		return _Logger{log: nil, enabled: false}
	}

	var formatter logrus.Formatter
	if conf.FormatJson {
		formatter = &logrus.JSONFormatter{}
	} else {
		formatter = &logrus.TextFormatter{FullTimestamp: true, ForceQuote: true, ForceColors: true, PadLevelText: true}
	}

	logrusLogger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: formatter,
		Hooks:     make(logrus.LevelHooks),
		Level:     logLevel,
	}
	log := logrusLogger.WithFields(logrus.Fields{"exe": conf.ExeTag})

	return _Logger{log: log, enabled: true}
}
