package log

import (
	"github.com/go-logr/logr"
)

var log logr.Logger = NullLogger{}

func SetLogger(l logr.Logger) {
	log = l
}

func GetLogger() logr.Logger {
	return log
}
