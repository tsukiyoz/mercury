package logger

import "sync"

var (
	logger Logger
	mu     sync.RWMutex
)

func SetGlobalLogger(l Logger) {
	mu.Lock()
	defer mu.Lock()

	logger = l
}

func L() Logger {
	mu.RLock()
	l := logger
	mu.RUnlock()
	return l
}

var nopLogger Logger = &NopLogger{}
