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

func DaemonMain() {
	log, err := logger.NewLog(logger.All, logger.All, "")

	if err != nil {
		panic(err)
	}

	defer log.Close()
	opts, err := server.LoadConfigFromPath("/etc/webby/config.json")

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

	sigtermChannel := make(chan os.Signal, 1)
	signal.Notify(sigtermChannel, syscall.SIGTERM, syscall.SIGINT)

	reload := false
	serverCommandChan := srv.StartThreaded()

	commandListener, err := NewDaemonListener(map[DaemonCommand]DaemonCommandCallback{
		Restart: GetRestartCallback(serverCommandChan),

		Reload: func(_ DaemonCommandArg) error {
			sigtermChannel <- syscall.SIGINT
			return nil
		},

		LogRecord: GetLogRecordCallback(&log),
		LogPrint:  GetLogPrintCallback(&log),
	}, log)

	usingDaemonSocket := true

	if err != nil {
		log.LogErr(err.Error())
		log.LogErr("Could not open Unix Domain Socket")
		usingDaemonSocket = false
	} else {
		go commandListener.Listen()
	}

	sig := <-sigtermChannel
	serverCommandChan <- server.Shutoff

	if !reload {
		log.LogWarn("Received signal: " + sig.String())
	} else {
		log.LogWarn("Received reload command")
	}

	if usingDaemonSocket {
		log.LogInfo("Closing Unix Domain Socket...")
		commandListener.Close()
	}

	log.LogInfo("Stopping server...")
	srv.Stop()

	log.LogInfo("Closing log...")
	log.Close()

	if reload {
		DaemonMain()
	}
}
