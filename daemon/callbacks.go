// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"os"

	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

// Represents a signal originating at a daemon command and sent through a
// channel by the reload callback.
type ReloadSignal struct{}

func (r ReloadSignal) String() string {
	return "Reload"
}

func (r ReloadSignal) Signal() {}

// Represent a signal originating at a daemmon command and sent through a
// channel by the stop callback
type StopSignal struct{}

func (r StopSignal) String() string {
	return "Stop"
}

func (r StopSignal) Signal() {}

// Returns a function that will sent the `server.Restart` constant through the
// given channel when called.
func GetRestartCallback(serverCommandChan chan server.ServerThreadCommand) DaemonCommandCallback {
	return func(_ DaemonCommandArg) error {
		serverCommandChan <- server.Restart
		return nil
	}
}

// Returns a function that will send a `ReloadSignal` though the given channel
// when called.
func GetReloadCallback(signalChan chan os.Signal) DaemonCommandCallback {
	return func(_ DaemonCommandArg) error {
		signalChan <- ReloadSignal{}
		return nil
	}
}

// Returns a function that will send a `StopSignal` through the given channel
// when called.
func GetStopCallback(signalChan chan os.Signal) DaemonCommandCallback {
	return func(_ DaemonCommandArg) error {
		signalChan <- StopSignal{}
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
