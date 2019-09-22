package config

import (
	"os"

	"github.com/op/go-logging"
)

var (
	logger = logging.MustGetLogger("shoptrac")
	logFormat = logging.MustStringFormatter("[%{time:2006-01-02 15:04:05}] %{level} %{message}")
)

func SetupLogging(logfile string, loglevel string) {
	var backend logging.Backend
	fp, _ := os.OpenFile(logfile, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0644); if fp != nil {
		backend = logging.NewLogBackend(fp, "", 0)
	} else {
		backend = logging.NewLogBackend(os.Stdout, "", 0)
	}

	realBackend := logging.AddModuleLevel(logging.NewBackendFormatter(backend, logFormat))
	realBackend.SetLevel(StringToLogLevel(loglevel), "")
	logging.SetBackend(realBackend)
}

func Logger() *logging.Logger {
	return logger
}

func StringToLogLevel(s string) logging.Level {
	l, err := logging.LogLevel(s)
	if err == nil {
		return l
	} else {
		return logging.INFO
	}
}