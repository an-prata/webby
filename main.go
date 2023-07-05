// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"github.com/an-prata/webby/daemon"
	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

func main() {
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

	if opts.Log != "" {
		log.LogInfo("Opening '" + opts.Log + "' for recording logs")
		err = log.OpenFile(opts.Log)

		if err != nil {
			log.LogErr("Could not open '" + opts.Log + "' for logging")
		}
	}

	printing, err := logger.LevelFromString(opts.LogLevelPrint)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' due to errors for printing")
	}

	recording, err := logger.LevelFromString(opts.LogLevelRecord)

	if err != nil {
		log.LogErr(err.Error())
		log.LogWarn("Using log level 'All' due to errors for recording")
	}

	log.Printing = printing
	log.Saving = recording

	srv, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		return
	}

	commandListener, err := daemon.NewDaemonListener(map[daemon.DaemonCommand]func(daemon.DaemonCommandArg) error{
		daemon.Restart: func(_ daemon.DaemonCommandArg) error {
			// When the `Server.Start()` function returns it is automatically called
			// again in a loop.
			return srv.Stop()
		},

		daemon.LogRecord: func(arg daemon.DaemonCommandArg) error {
			logLevel := logger.LogLevel(arg)
			logLevel, err := logger.CheckLogLevel(uint8(logLevel))

			if err != nil {
				log.LogWarn("invalid log level given, using 'All'")
			}

			log.Saving = logLevel
			return nil
		},

		daemon.LogPrint: func(arg daemon.DaemonCommandArg) error {
			logLevel := logger.LogLevel(arg)
			logLevel, err := logger.CheckLogLevel(uint8(logLevel))

			if err != nil {
				log.LogWarn("invalid log level given, using 'All'")
			}

			log.Printing = logLevel
			return nil
		},
	}, log)

	if err != nil {
		log.LogErr("Could not open Unix Domain Socket")
	} else {
		go commandListener.Listen()
	}

	for {
		// Will restart the server on close.
		log.LogErr(srv.Start().Error())
		srv, err = server.NewServer(opts, &log)

		if err != nil {
			log.LogErr(err.Error())
			return
		}
	}
}
