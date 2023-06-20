// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"io"
	"net/http"

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

	if opts.LogLevelPrint > logger.All || opts.LogLevelPrint < 0 {
		log.LogErr("Unexpected log level for printing, using 'All'")
		opts.LogLevelPrint = logger.All
	} else {
		log.Printing = opts.LogLevelPrint
	}

	if opts.LogLevelRecord > logger.All || opts.LogLevelRecord < 0 {
		log.LogErr("Unexpected log level for recording, using 'All'")
		opts.LogLevelRecord = logger.All
	} else {
		log.Saving = opts.LogLevelRecord
	}

	server, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		return
	}

	log.LogErr(server.Start().Error())
	return
}
