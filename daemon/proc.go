// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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
		LogRecord: GetLogRecordCallback(&log),
		LogPrint:  GetLogPrintCallback(&log),
	}, log)

	if opts.AutoReload {
		go func() {
			previousOpts := opts
			for {
				currentOpts, err := server.LoadConfigFromPath(CONFIG_PATH)

				if err != nil {
					log.LogErr("Could not auto reload config due to read errors")
				} else if currentOpts.Equals(&previousOpts) {
					log.LogInfo("Detected config change, sending reload signal...")
					signalChan <- ReloadSignal{}
				}

				// Thats five seconds, I really dont like Go's durations.
				time.Sleep(time.Duration(5_000_000_000))
			}
		}()
	}

	usingDaemonSocket := true

	if err != nil {
		log.LogErr(err.Error())
		log.LogErr("Could not open Unix Domain Socket")
		usingDaemonSocket = false
	} else {
		go commandListener.Listen()
	}

	sig := <-signalChan
	serverCommandChan <- server.Shutoff
	log.LogInfo("Received signal: " + sig.String())

	if usingDaemonSocket {
		log.LogInfo("Closing Unix Domain Socket...")
		commandListener.Close()
	}

	log.LogInfo("Stopping server...")
	srv.Stop()

	log.LogInfo("Closing log...")
	log.Close()

	_, ok := sig.(ReloadSignal)

	if ok {
		goto Start
	}
}
