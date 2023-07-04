// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"io"
	"net/http"

	"github.com/an-prata/webby/daemon"
	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

func defaultResponse(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Page not found...\n")
}

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

	server, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		return
	}

	log.LogErr(server.Start().Error())
	return
}
