// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/an-prata/webby/logger"
)

const DefaultSitePath = "/srv/webby/"

// Represents a command that may be given to a running server thread through a
// channel.
type ServerThreadCommand = uint8

const (
	// Shuts off the running thread and returns.
	Shutoff ServerThreadCommand = iota

	// Will close the current server and reinstantiate it from the same options and
	// log as provided during construction.
	Restart
)

type Server struct {
	Hndlr *Handler
	srv   *http.Server
	log   *logger.Log
	opts  ServerOptions
}

// Creates a new server given the specified options. Will return an error if any
// of the given paths could not be statted or if the program lacks read
// permissions. This function will map directories from the options given.
func NewServer(opts ServerOptions, log *logger.Log) (*Server, error) {
	var err error
	opts.checkForDefaults()

	if _, err = os.Stat(opts.Site); err != nil {
		return nil, errors.New("Could not stat '" + opts.Site + "'")
	}

	if opts.SupportsTLS() {
		if _, err = os.Stat(opts.Cert); err != nil {
			return nil, errors.New("Could not stat '" + opts.Cert + "'")
		}

		if _, err = os.Stat(opts.Key); err != nil {
			return nil, errors.New("Could not stat '" + opts.Key + "'")
		}
	}

	var port string

	if opts.Port > 0 {
		port = ":" + strconv.FormatInt(int64(opts.Port), 10)
	} else {
		port = ""
	}

	handler := NewHandler(log)
	handler.MapDir(opts.Site)
	handler.AddDeadResponses(opts.DeadPaths)

	httpSrv := http.Server{
		Addr:              port,
		Handler:           handler,
		ReadHeaderTimeout: time.Minute,
		WriteTimeout:      time.Minute,
	}

	return &Server{handler, &httpSrv, log, opts}, nil
}

// Starts the server, if TLS is supports then it is started in another thread
// and regular HTTP is started in the current thread. This function will only
// ever return on an error. If the server is started in this fashion then it may
// be stopped using the `Server.Stop()` method, in which case it will return an
// error indicated this.
func (s *Server) Start() error {
	if s.opts.SupportsTLS() {
		go s.srv.ListenAndServeTLS(s.opts.Cert, s.opts.Key)
	}

	return s.srv.ListenAndServe()

}

// Starts the server in a seperate thread and returns a channel for giving said
// thread commands. This method, unlike the more standard `Server.Start()`
// method, cannot be stopped using the `Server.Stop()` method and must instead
// be instructed to stop using the provided channel. This method also does not
// report errors except in logs.
func (s *Server) StartThreaded() chan ServerThreadCommand {
	commandChan := make(chan ServerThreadCommand, 1)

	go func() {
		for {
			go s.Start()
			command := <-commandChan
			s.Stop()

			if command == Shutoff {
				s.log.LogInfo("HTTP server shutting off...")
				return
			} else if command == Restart {
				s.log.LogInfo("HTTP server restarting...")
				srv, err := NewServer(s.opts, s.log)

				if err != nil {
					s.log.LogErr("Could not reinstantiate HTTP server")
					return
				}

				*s = *srv
			}
		}
	}()

	return commandChan
}

// Stops a server started by the `Server.Start()` method. This method will not
// stop servers started using the `Server.StartThreaded()` method.
func (s *Server) Stop() error {
	return s.srv.Close()
}
