package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testing"
)

var log logr.Logger = testing.NullLogger{}

func SetLogger(l logr.Logger) {
	log = l
}

func GetLogger() logr.Logger {
	return log
}
