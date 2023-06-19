// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/an-prata/webby/logger"
)

const DefaultSitePath = "/srv/webby/"

type Server struct {
	srv  *http.Server
	log  *logger.Log
	opts ServerOptions
}

type ServerOptions struct {
	// Path to the root of the website to host. Use an empty string for default.
	// See `server.DefaultSitePath`
	Site string

	// Path to a TLS/SSL certificate. Use an empty string for no HTTPS.
	Cert string

	// Path to a TLS/SSL private key. Use an empty string for no HTTPS.
	Key string

	supportsTLS bool
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

	handler := NewHandler(log)
	handler.MapDir(opts.Site)

	httpSrv := http.Server{
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

func (opts *ServerOptions) checkForDefaults() {
	if opts.Site == "" {
		opts.Site = DefaultSitePath
	}

	opts.supportsTLS = opts.Cert != "" && opts.Key != ""

	if opts.Site[len(opts.Site)-1] != '/' {
		opts.Site += "/"
	}
}
