// Copyright (c) 2023 Evan Overman (https://an-prata.it).
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
	log, err := logger.NewLog(logger.All, logger.All, "")

	if err != nil {
		panic(err)
	}

	defer log.Close()
	opts, err := server.LoadConfigFromPath(CONFIG_PATH)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using default configuration due to errors")
	}

	err = log.OpenFile(opts.Log)

	if err != nil {
		log.LogErr("Could not open '" + opts.Log + "' for logging")
	}

	err = log.SetRecordLevelFromString(opts.LogLevelPrint)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' for printing due to errors")
	}

	err = log.SetPrintLevelFromString(opts.LogLevelRecord)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' for recording due to errors")
	}

	srv, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		return
	}

	serverCommandChan := srv.StartThreaded()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	commandListener, err := NewDaemonListener(map[DaemonCommand]DaemonCommandCallback{
		Restart:   GetRestartCallback(serverCommandChan),
		Reload:    GetReloadCallback(signalChan),
		Stop:      GetStopCallback(signalChan),
		Status:    GetStatusCallback(srv.Hndlr, &log),
		LogRecord: GetLogRecordCallback(&log),
		LogPrint:  GetLogPrintCallback(&log),
	}, log)

	if err != nil {
		log.LogErr(err.Error())
		log.LogErr("Could not open Unix Domain Socket")
		os.Exit(1)
	}

	go commandListener.Listen()

	if opts.AutoReload {
		server.CallOnChange(func(signal server.FileChangeSignal) bool {
			if signal == server.TimeModifiedChange || signal == server.SizeChange {
				log.LogInfo("Config file change detected, reloading...")
				signalChan <- ReloadSignal{}
				return true
			} else if signal == server.InitialReadError || signal == server.ReadError {
				log.LogErr("Failed to read config while checking for change (auto reload is on)")
			}

			return false
		}, CONFIG_PATH)
	}

	sig := <-signalChan
	serverCommandChan <- server.Shutoff
	log.LogInfo("Received signal: " + sig.String())

	log.LogInfo("Closing Unix Domain Socket...")
	commandListener.Close()

	log.LogInfo("Stopping server...")
	srv.Stop()

	log.LogInfo("Closing log...")
	log.Close()

	_, ok := sig.(ReloadSignal)

	if ok {
		goto Start
	}
}
