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

// Responsible for handling HTTP requests with one of a custom response from a
// custom handler, or a static file, prioritized in that order.
type Handler struct {
	// List of all valid web paths that this handler will respond to.
	ValidPaths []string

	// A map of URL paths to their corosponding file path.
	PathMap map[string]string

	handlerMap map[string]http.Handler
}

// A custom handler that may respond with special or dynamic data rather than a
// static file.
type CustomHandler struct {
	Handler func(http.ResponseWriter, *http.Request)
}

// Creates a new handler that will log messages to the given log.
func NewHandler() *Handler {
	return &Handler{
		[]string{},
		map[string]string{},
		map[string]http.Handler{},
	}
}

// Maps the given request URI to a file path. Returns an error if a stat of the
// given file path fails.
func (h *Handler) MapFile(uriPath, filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		logger.GlobalLog.LogErr("Could not map '" + uriPath + "' to '" + filePath + "' due to failed stat")
		return errors.New("Could not stat '" + filePath + "'")
	}

	logger.GlobalLog.LogInfo("Mapped URI '" + uriPath + "' to file '" + filePath + "'")
	h.PathMap[uriPath] = filePath
	h.ValidPaths = append(h.ValidPaths, uriPath)

	if strings.Contains(uriPath, "..") {
		logger.GlobalLog.LogWarn("Mapped file using '..', this may add security vulnerabilities")
	}

	return nil
}

// Map a directory and all subdirectories to paths on the server. All directory
// roots, when requested, will serve an "index.html" file from that directory.
func (h *Handler) MapDir(dirPath string) error {
	err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if _, err := os.Stat(path); err != nil {
			logger.GlobalLog.LogErr("Could not stat '" + path + "'")
			return nil
		}

		path = strings.ReplaceAll(path, dirPath, "")

		if d.IsDir() {
			h.PathMap["/"+path] = dirPath + path + "index.html"
			logger.GlobalLog.LogInfo("Mapped URI '/" + path + "index.html' to file '" + dirPath + path + "'")
		} else {
			h.PathMap["/"+path] = dirPath + path
			logger.GlobalLog.LogInfo("Mapped URI '/" + path + "' to file '" + dirPath + path + "'")
		}

		h.ValidPaths = append(h.ValidPaths, "/"+path)
		return nil
	})

	if err != nil {
		return errors.New("Could not walk directory '" + dirPath + "'")
	}

	return nil
}

// For each path given a response that redirects the client to the same path but
// on itself (e.g. "http://localhost/some/dead/path") will be given. This
// creates a custom handler, adding another custom handler will override this
// dead response. If a file is mapped to the same path as this dead response
// then, like other custom handlers, the dead response takes priority.
func (h *Handler) AddDeadResponses(paths []string) {
	for _, path := range paths {
		if len(path) > 0 && path[0] != '/' {
			path = "/" + path
		}

		logger.GlobalLog.LogInfo("Mapped URI '" + path + "' to a dead response.")
		h.handlerMap[path] = CustomHandler{
			Handler: func(w http.ResponseWriter, req *http.Request) {
				logger.GlobalLog.LogInfo("Dead responding to request from '" + req.RemoteAddr + "'")
				http.Redirect(w, req, "http://localhost/"+path, http.StatusMovedPermanently)
			},
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	logger.GlobalLog.LogInfo("Got request (" + req.Proto + ") from " + req.RemoteAddr + " for " + req.URL.Path)

	if strings.Contains(req.URL.Path, "..") {
		logger.GlobalLog.LogWarn("Request was made to a path containing '..' by " + req.RemoteAddr)
	}

	handler, ok := h.handlerMap[req.URL.Path]

	if ok {
		handler.ServeHTTP(w, req)
		return
	}

	file, ok := h.PathMap[req.URL.Path]

	if ok {
		if _, err := os.Stat(file); err != nil {
			logger.GlobalLog.LogErr("A request was made for '" + file + "' but stat failed")
		}

		http.ServeFile(w, req, file)
		return
	}

	// No file nor special handler for requested path.
	http.NotFound(w, req)
}

func (h CustomHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.Handler(w, req)
}
