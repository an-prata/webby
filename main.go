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

	if _, err = os.Stat(serverPath); err != nil {
		log.LogErr("Server path " + serverPath + " does not exist. Exiting...")
		os.Exit(1)
	}

	if _, err = os.Stat(sitePath); err != nil {
		log.LogErr("Website path " + sitePath + " does not exist. Exiting...")
		os.Exit(1)
	}

	err = filepath.WalkDir(sitePath, func(path string, d fs.DirEntry, err error) error {
		filePath := strings.Clone(path)
		path = strings.ReplaceAll(path, sitePath, "")

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

	_, errCert := os.Stat(certPath)
	_, errKey := os.Stat(keyPath)

	if errCert != nil || errKey != nil {
		log.LogWarn("Could not find certificate or keys. HTTPS will not be supported.")
		http.ListenAndServe(":8080", nil)
		return
	}

	http.ListenAndServeTLS(":8443", certPath, keyPath, nil)
}
