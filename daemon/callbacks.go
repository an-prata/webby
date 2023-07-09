// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"os"

	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

// The only argument that will be given to the callbacks for deamon commands.
// Each callback may interperet this differently, for example, the restart
// command ignores its argument, but log level commands will interperet this to
// be a log level.
type DaemonCommandArg uint8

// The success/failure of a daemon command. This will appear as a single byte
// response to any client commands indicating the success or failure of a
// command.
type DaemonCommandSuccess uint8

// Type alias for the function signature of a daemon command callback.
type DaemonCommandCallback func(DaemonCommandArg) error

const (
	// The daemon command completed successfuly.
	Success DaemonCommandSuccess = iota

	// The daemon command failed.
	Failure
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
