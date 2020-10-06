package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testing"
)

var log logr.Logger = testing.NullLogger{}

// SetLogger sets the logger to be used for this library.
func SetLogger(l logr.Logger) {
	log = l
}

// GetLogger returns the logger object.
func GetLogger() logr.Logger {
	return log
}
