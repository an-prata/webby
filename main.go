// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"io"
	"net/http"
	"os"

	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

const serverPath = "/srv/webby"
const sitePath = serverPath + "/website"
const certPath = serverPath + "/cert.pem"
const keyPath = serverPath + "/key.pem"

func defaultResponse(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Page not found...\n")
}

func main() {
	log, err := logger.NewLog(logger.All, logger.None, "")

	if err != nil {
		panic(err)
	}

	opts := server.ServerOptions{
		Site: sitePath, Cert: "", Key: "",
	}

	server, err := server.NewServer(opts, &log)

	if err != nil {
		log.LogErr(err.Error())
		os.Exit(1)
	}

	log.LogErr(server.Start().Error())
	os.Exit(1)
}
