// Copyright (c) 2024 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

const CONFIG_PATH = "/etc/webby/config.json"

// Main function of daemon execution.
func DaemonMain() {
Start:
	opts, err := server.LoadConfigFromPath(CONFIG_PATH)

	if err != nil {
		logger.GlobalLog.LogErr(err.Error())
		logger.GlobalLog.LogWarn("Using default configuration due to errors")
	}

	opts.Show()

	err = logger.GlobalLog.OpenFile(opts.Log)

	if err != nil {
		logger.GlobalLog.LogErr("Could not open '" + opts.Log + "' for logging")
	}

	err = logger.GlobalLog.SetRecordLevelFromString(opts.LogLevelPrint)

	if err != nil {
		logger.GlobalLog.LogErr(err.Error())
		logger.GlobalLog.LogWarn("Using log level 'All' for printing due to errors")
	}

	err = logger.GlobalLog.SetPrintLevelFromString(opts.LogLevelRecord)

	if err != nil {
		logger.GlobalLog.LogErr(err.Error())
		logger.GlobalLog.LogWarn("Using log level 'All' for recording due to errors")
	}

	srv, err := server.NewServer(opts)

	if err != nil {
		logger.GlobalLog.LogErr(err.Error())
		return
	}

	serverCommandChan := srv.StartThreaded()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	commandListener, err := NewDaemonListener(map[DaemonCommand]DaemonCommandCallback{
		Restart:   GetRestartCallback(serverCommandChan),
		Reload:    GetReloadCallback(signalChan),
		Stop:      GetStopCallback(signalChan),
		Status:    GetStatusCallback(srv.ReqHandler),
		LogRecord: GetLogRecordCallback(),
		LogPrint:  GetLogPrintCallback(),
	})

	if err != nil {
		logger.GlobalLog.LogErr(err.Error())
		logger.GlobalLog.LogErr("Could not open Unix Domain Socket")
		os.Exit(1)
	}

	go commandListener.Listen()

	if opts.AutoReload {
		server.CallOnChange(func(signal server.FileChangeSignal) bool {
			if signal == server.TimeModifiedChange || signal == server.SizeChange {
				logger.GlobalLog.LogInfo("Config file change detected, reloading...")
				signalChan <- ReloadSignal{}
				return true
			} else if signal == server.InitialReadError || signal == server.ReadError {
				logger.GlobalLog.LogErr("Failed to read config while checking for change (auto reload is on)")
			}

			return false
		}, CONFIG_PATH)

		for _, filePath := range srv.ReqHandler.PathMap {
			server.CallOnChange(func(signal server.FileChangeSignal) bool {
				if signal == server.TimeModifiedChange || signal == server.SizeChange {
					logger.GlobalLog.LogInfo("Site file change detected, reloading...")
					signalChan <- ReloadSignal{}
					return true
				} else if signal == server.InitialReadError || signal == server.ReadError {
					logger.GlobalLog.LogErr("Failed to read site file while checking for change (auto reload is on): " + filePath)
				}

				return false
			}, filePath)
		}
	}

	sig := <-signalChan
	serverCommandChan <- server.Shutoff
	logger.GlobalLog.LogInfo("Received signal: " + sig.String())

	logger.GlobalLog.LogInfo("Closing Unix Domain Socket...")
	commandListener.Close()

	logger.GlobalLog.LogInfo("Stopping server...")
	srv.Stop()

	logger.GlobalLog.LogInfo("Closing log...")
	logger.GlobalLog.Close()

	_, ok := sig.(ReloadSignal)

	if ok {
		goto Start
	}
}
