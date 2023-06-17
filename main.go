// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/an-prata/webby/logger"
)

const serverPath = "/srv/http"

func defaultResponse(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "Page not found...\n")
}

func main() {
	log, err := logger.NewLog(logger.All, logger.None, "")

	if err != nil {
		panic(err)
	}

	err = filepath.WalkDir(serverPath, func(path string, d fs.DirEntry, err error) error {
		filePath := path
		path = strings.ReplaceAll(path, serverPath, "")

		if d.IsDir() {
			filePath += "/index.html"
			path += "/"
		}

		log.LogInfo("Mapped " + filePath + " to " + path)

		http.HandleFunc(path, func(w http.ResponseWriter, req *http.Request) {
			if _, err := os.Stat(filePath); err != nil {
				log.LogErr(filePath + " no longer exists. Using default response.")
				defaultResponse(w, req)
				return
			}

			log.LogInfo("Got request from " + req.RemoteAddr + " for " + req.RequestURI)
			http.ServeFile(w, req, filePath)
		})

		return nil
	})

	if err != nil {
		panic(err)
	}

	http.ListenAndServe(":8080", nil)
}
