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

type Server struct {
	srv  *http.Server
	log  *logger.Log
	opts ServerOptions
}

// Creates a new server given the specified options. Will return an error if any
// of the given paths could not be statted or if the program lacks read
// permissions.
func NewServer(opts ServerOptions, log *logger.Log) (*Server, error) {
	var err error
	opts.checkForDefaults()

	if _, err = os.Stat(opts.Site); err != nil {
		return nil, errors.New("Could not stat '" + opts.Site + "'")
	}

	if opts.supportsTLS {
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

	return &Server{&httpSrv, log, opts}, nil
}

func (s *Server) Start() error {
	if s.opts.supportsTLS {
		go s.srv.ListenAndServeTLS(s.opts.Cert, s.opts.Key)
	}

	return s.srv.ListenAndServe()

}
