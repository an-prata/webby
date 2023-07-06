// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

func GetRestartCallback(serverCommandChan chan server.ServerThreadCommand) DaemonCommandCallback {
	return func(arg DaemonCommandArg) error {
		serverCommandChan <- server.Restart
		return nil
	}
}

// Returns a function, that when called, will modify the given log's recording
// log level to match its parameters.
func GetLogPrintCallback(log *logger.Log) DaemonCommandCallback {
	return func(arg DaemonCommandArg) error {
		logLevel := logger.LogLevel(arg)
		logLevel, err := logger.CheckLogLevel(uint8(logLevel))

		if err != nil {
			log.LogWarn("Invalid log level given, using 'All'")
		}

		log.Printing = logLevel
		return nil
	}
}

// Returns a function, that when called, will modify the given log's printing
// log level to match its parameters.
func GetLogRecordCallback(log *logger.Log) DaemonCommandCallback {
	return func(arg DaemonCommandArg) error {
		logLevel := logger.LogLevel(arg)
		logLevel, err := logger.CheckLogLevel(uint8(logLevel))

		if err != nil {
			log.LogWarn("Invalid log level given, using 'All'")
		}

		log.Recording = logLevel
		return nil
	}
}
