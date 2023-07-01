// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/an-prata/webby/logger"
)

type Handler struct {
	pathMap    map[string]string
	handlerMap map[string]http.Handler
	log        *logger.Log
}

type CustomHandler struct {
	Handler func(http.ResponseWriter, *http.Request)
}

func NewHandler(log *logger.Log) *Handler {
	return &Handler{
		map[string]string{},
		map[string]http.Handler{},
		log,
	}
}

// Maps the given request URI to a file path. Returns an error if a stat of the
// given file path fails.
func (h *Handler) MapFile(uriPath, filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		h.log.LogErr("Could not map '" + uriPath + "' to '" + filePath + "' due to failed stat")
		return errors.New("Could not stat '" + filePath + "'")
	}

	h.log.LogInfo("Mapped URI '" + uriPath + "' to file '" + filePath + "'")
	h.pathMap[uriPath] = filePath

	if strings.Contains(uriPath, "..") {
		h.log.LogWarn("Mapped file using '..', this may add security vulnerabilities")
	}

	return nil
}

// Map a directory and all subdirectories to paths on the server. All directory
// roots, when requested, will serve an "index.html" file from that directory.
func (h *Handler) MapDir(dirPath string) error {
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if _, err := os.Stat(path); err != nil {
			h.log.LogErr("Could not stat '" + path + "'")
			return nil
		}

		path = strings.ReplaceAll(path, dirPath, "")

		if d.IsDir() {
			h.pathMap["/"+path] = dirPath + path + "index.html"
			h.log.LogInfo("Mapped URI '/" + path + "index.html' to file '" + dirPath + path + "'")
		} else {
			h.pathMap["/"+path] = dirPath + path
			h.log.LogInfo("Mapped URI '/" + path + "' to file '" + dirPath + path + "'")
		}

		return nil
	})

	if err != nil {
		return errors.New("Could not walk directory '" + dirPath + "'")
	}

	return nil
}

// For each path given a response that redirects the client to the same path but
// on itself (e.g. "http://localhost/some/dead/path") will be given.
func (h *Handler) AddDeadResponses(paths []string) {
	for _, path := range paths {
		if len(path) > 0 || path[0] != '/' {
			path = "/" + path
		}

		h.log.LogInfo("Mapping URI '" + path + "' to a dead response.")
		h.handlerMap[path] = CustomHandler{
			Handler: func(w http.ResponseWriter, req *http.Request) {
				h.log.LogInfo("Dead responding to request from '" + req.RemoteAddr + "'")
				http.Redirect(w, req, "http://localhost/"+path, http.StatusMovedPermanently)
			},
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.log.LogInfo("Got request (" + req.Proto + ") from " + req.RemoteAddr + " for " + req.URL.Path)
	file, ok := h.pathMap[req.URL.Path]

	if ok {
		if _, err := os.Stat(file); err != nil {
			h.log.LogErr("A request was made for '" + file + "' but stat failed")
		}

		http.ServeFile(w, req, file)
		return
	}

	handler, ok := h.handlerMap[req.URL.Path]

	if ok {
		handler.ServeHTTP(w, req)
		return
	}

	// No file nor special handler for requested path.
	http.NotFound(w, req)
}

func (h CustomHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Handler(w, req)
}
