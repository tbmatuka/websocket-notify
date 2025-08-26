package websocket_notify

import (
	"log"
	"sync"
)

type Logger struct {
	debug bool
}

var (
	loggerLock = &sync.Mutex{} //nolint:gochecknoglobals
	logger     *Logger         //nolint:gochecknoglobals
)

func getLogger() *Logger {
	if logger == nil {
		// only lock on initialization
		loggerLock.Lock()

		// check again after lock
		if logger == nil {
			logger = new(Logger)
		}

		loggerLock.Unlock()
	}

	return logger
}

func (logger *Logger) SetDebug(newDebug bool) {
	logger.debug = newDebug
}

func (logger *Logger) IsDebug() bool {
	return logger.debug
}

func (logger *Logger) Debug(message string) {
	if logger.debug {
		log.Println(message)
	}
}
