// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"net/http"
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

const (
	// The daemon command completed successfuly.
	Success DaemonCommandSuccess = iota

	// The daemon command failed.
	Failure
)

// Represents the status returned by the status callback
type WebbyStatus uint8

const (
	// We make sure that the first bit can be compared the same way it could on the
	// original success constants.

	Ok              WebbyStatus = WebbyStatus(Success)                     // All gets gave 200
	HttpNon2xx      WebbyStatus = WebbyStatus(Failure) | ((iota + 1) << 1) // Not every get gave 200
	HttpPartialFail                                                        // Some gets gave code >= 400
	HttpFail                                                               // All gets gave code >= 400
)

// Type alias for the function signature of a daemon command callback.
type DaemonCommandCallback func(DaemonCommandArg) DaemonCommandSuccess

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
	return func(_ DaemonCommandArg) DaemonCommandSuccess {
		serverCommandChan <- server.Restart
		return Success
	}
}

// Returns a function that will send a `ReloadSignal` though the given channel
// when called.
func GetReloadCallback(signalChan chan os.Signal) DaemonCommandCallback {
	return func(_ DaemonCommandArg) DaemonCommandSuccess {
		signalChan <- ReloadSignal{}
		return Success
	}
}

// Returns a function that will send a `StopSignal` through the given channel
// when called.
func GetStopCallback(signalChan chan os.Signal) DaemonCommandCallback {
	return func(_ DaemonCommandArg) DaemonCommandSuccess {
		signalChan <- StopSignal{}
		return Success
	}
}

// Returns a function that simply returns `Success` when called. If callbacks
// are being called and the daemon can give the success message to a connection
// then we consider this to be "ok" on webby's side.
func GetStatusCallback(handler *server.Handler) DaemonCommandCallback {
	return func(_ DaemonCommandArg) DaemonCommandSuccess {
		getsFailed := 0
		getsNot200 := 0

		for _, path := range handler.ValidPaths {
			response, err := http.Get("http://localhost" + path)

			if err != nil {
				logger.GlobalLog.LogErr(err.Error())
				logger.GlobalLog.LogErr("Could not make GET request to path '" + path + "'")
				getsFailed++
				continue
			}

			if response.StatusCode >= 400 {
				getsFailed++
			}

			if response.StatusCode != 200 {
				getsNot200++
			}
		}

		if getsFailed >= len(handler.ValidPaths) {
			logger.GlobalLog.LogErr("All HTTP requests made for status check failed")
			logger.GlobalLog.LogInfo("Status requested, giving 'HttpFail'")
			return DaemonCommandSuccess(HttpFail)
		}

		if getsFailed > 1 {
			logger.GlobalLog.LogErr("Some HTTP requests made for status check failed")
			logger.GlobalLog.LogInfo("Status requested, giving 'HttpPartialFail'")
			return DaemonCommandSuccess(HttpPartialFail)
		}

		if getsNot200 > 1 {
			logger.GlobalLog.LogWarn("Some HTTP requests made for status check gave code other that '200'")
			logger.GlobalLog.LogInfo("Status requests, giving 'HttpNon2xx'")
			return DaemonCommandSuccess(HttpNon2xx)
		}

		logger.GlobalLog.LogInfo("Status requested, giving 'OK'")
		return DaemonCommandSuccess(Ok)
	}
}

// Returns a function, that when called, will modify the given log's recording
// log level to match its parameters.
func GetLogPrintCallback() DaemonCommandCallback {
	return func(arg DaemonCommandArg) DaemonCommandSuccess {
		logLevel := logger.LogLevel(arg)
		logLevel, err := logger.CheckLogLevel(uint8(logLevel))

		if err != nil {
			logger.GlobalLog.LogWarn("Invalid log level given, using 'All'")
		}

		logger.GlobalLog.Printing = logLevel
		return Success
	}
}

// Returns a function, that when called, will modify the given log's printing
// log level to match its parameters.
func GetLogRecordCallback() DaemonCommandCallback {
	return func(arg DaemonCommandArg) DaemonCommandSuccess {
		logLevel := logger.LogLevel(arg)
		logLevel, err := logger.CheckLogLevel(uint8(logLevel))

		if err != nil {
			logger.GlobalLog.LogWarn("Invalid log level given, using 'All'")
		}

		logger.GlobalLog.Recording = logLevel
		return Success
	}
}
