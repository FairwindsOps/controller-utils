// Copyright 2020 FairwindsOps Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"log"
	"os"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

var logger logr.Logger = logr.Discard()

func init() {
	if os.Getenv("CONTROLLER_UTILS_LOG_LEVEL") != "" {
		logLevel, err := strconv.Atoi(os.Getenv("CONTROLLER_UTILS_LOG_LEVEL"))
		if err != nil {
			panic(err)
		}
		stdr.SetVerbosity(logLevel)
		SetLogger(stdr.New(log.New(os.Stdout, "", 0)))
	}
}

// SetLogger sets the logger to be used for this library.
func SetLogger(l logr.Logger) {
	logger = l
}

// GetLogger returns the logger object.
func GetLogger() logr.Logger {
	return logger
}
