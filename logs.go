package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type LogConfig struct {
	level  string
	format string
	pretty bool
	colors bool
}

func configureLogs(config LogConfig) error {

	level, err := log.ParseLevel(config.level)
	if err != nil {
		return fmt.Errorf(`error while configuring logger: %v`, err)
	}

	log.SetLevel(level)
	log.SetOutput(os.Stdout)

	if strings.ToLower(config.format) == "json" {
		log.SetFormatter(&log.JSONFormatter{
			PrettyPrint: config.pretty,
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			DisableColors: !config.colors,
			FullTimestamp: true,
		})
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debugf(`set log format to %s`, config.format)
		log.Debugf(`set log level to %s`, config.level)
	}

	return nil
}

type LogrusWriter interface {
	io.Writer
	WithLevel(log.Level) LogrusWriter
	WithLogger(*log.Logger) LogrusWriter
	WithFields(log.Fields) LogrusWriter
}

type logrusWriter struct {
	level  log.Level
	logger *log.Logger
	fields log.Fields
}

func NewLogrusWriter(level log.Level) LogrusWriter {
	return &logrusWriter{
		level:  level,
		logger: log.StandardLogger(),
	}
}

func (l *logrusWriter) WithLevel(level log.Level) LogrusWriter {
	l.level = level
	return l
}

func (l *logrusWriter) WithLogger(logger *log.Logger) LogrusWriter {
	l.logger = logger
	return l
}

func (l *logrusWriter) WithFields(fields log.Fields) LogrusWriter {
	l.fields = fields
	return l
}

func (l *logrusWriter) Write(p []byte) (int, error) {
	logger := l.logger
	if logger.IsLevelEnabled(l.level) {
		fields := l.fields
		text := string(p)
		if text[0] == 10 || text[0] == 13 {
			text = text[1:]
		}
		if text[len(text)-1] == 10 || text[len(text)-1] == 13 {
			text = text[:len(text)-1]
		}
		logger.WithFields(fields).Log(l.level, text)
		return len(p), nil
	}
	return 0, nil
}
