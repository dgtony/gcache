package utils

import (
	"github.com/op/go-logging"
	"os"
)

var defaultBackend logging.LeveledBackend

func SetupLoggers(c *Config) {
	var format logging.Formatter
	switch c.General.LogFormat {
	case "long":
		format = logging.MustStringFormatter(
			`%{color}%{time:15:04:05.000} %{module:8s} %{shortfunc} ▶ %{level:.6s} %{id:03x}%{color:reset} %{message}`)
	default:
		format = logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{module:8s} ▶ %{level:.6s}%{color:reset} %{message}`)
	}

	var backend *logging.LogBackend
	switch c.General.LogOut {
	case "stderr":
		backend = logging.NewLogBackend(os.Stderr, "", 0)
	default:
		backend = logging.NewLogBackend(os.Stdout, "", 0)
	}
	backendFormatter := logging.NewBackendFormatter(backend, format)

	logLevelS := c.General.LogLevel
	var logLevel logging.Level
	switch logLevelS {
	case "debug":
		logLevel = logging.DEBUG
	case "notice":
		logLevel = logging.NOTICE
	case "warning":
		logLevel = logging.WARNING
	case "error":
		logLevel = logging.ERROR
	case "critical":
		logLevel = logging.CRITICAL
	default:
		logLevel = logging.INFO
	}

	backendLeveled := logging.AddModuleLevel(backendFormatter)
	backendLeveled.SetLevel(logLevel, "")
	defaultBackend = backendLeveled
}

func GetLogger(name string) *logging.Logger {
	logger := logging.MustGetLogger(name)
	logger.SetBackend(defaultBackend)
	return logger
}
