package dodod

import (
	"log"
	"os"
)

// Logger interface
type Logger interface {
	Errorf(f string, v ...interface{})
	Warningf(f string, v ...interface{})
	Infof(f string, v ...interface{})
	Debugf(f string, v ...interface{})
}

type defaultLogger struct {
	*log.Logger
}

func (l *defaultLogger) Errorf(f string, v ...interface{}) {
	l.Printf("ERROR: "+f, v...)
}

func (l *defaultLogger) Warningf(f string, v ...interface{}) {
	l.Printf("WARNING: "+f, v...)
}

func (l *defaultLogger) Infof(f string, v ...interface{}) {
	l.Printf("INFO: "+f, v...)
}

func (l *defaultLogger) Debugf(f string, v ...interface{}) {
	l.Printf("DEBUG: "+f, v...)
}

// DefaultLogger is the default logger for dodod
// Set different logger to modify the logging behavior
var DefaultLogger Logger = &defaultLogger{Logger: log.New(os.Stderr, "dodod ", log.LstdFlags)}
