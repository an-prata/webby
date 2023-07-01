// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"encoding/json"
	"errors"
	"os"
)

type ServerOptions struct {
	// Path to the root of the website to host. Use an empty string for default.
	// See `server.DefaultSitePath`
	Site string

	// Path to a TLS/SSL certificate. Use an empty string for no HTTPS.
	Cert string

	// Path to a TLS/SSL private key. Use an empty string for no HTTPS.
	Key string

	// The port to host on, negative numbers and zero will utilize a default (80
	// for HTTP and 443 for HTTPS).
	Port int32

	// Path to a file for logging. Use an empty string for no log file.
	Log string

	// Log level for printing to standard out. Can be "All", "None", "Error",
	// "Warning", or "Info".
	LogLevelPrint string

	// Log level for writing to file out. Can be "All", "None", "Error", "Warning",
	// or "Info".
	LogLevelRecord string

	// Paths that should be granted a dead response, can be used for fucking with
	// bot probing or the like. A dead response is just the name I gave to
	// redirecting a request back onto the client for the same path.
	DeadPaths []string
}

// Tries to parse JSON for a `ServerOptions` with the file at the given path.
// Returns an error and a default configuration on failure.
func LoadConfigFromPath(path string) (ServerOptions, error) {
	if _, err := os.Stat(path); err != nil {
		return DefaultOptions(), errors.New("Could not stat config at '" + path + "'")
	}

	var opts ServerOptions

	bytes, err := os.ReadFile(path)

	if err != nil {
		return DefaultOptions(), errors.New("Could not read config at '" + path + "'")
	}

	if json.Unmarshal(bytes, &opts) != nil {
		return DefaultOptions(), errors.New("Could not parse config JSON at '" + path + "'")
	}

	return opts, nil
}

func DefaultOptions() ServerOptions {
	return ServerOptions{
		Site:           "/srv/webby/website",
		Cert:           "",
		Key:            "",
		Port:           -1,
		Log:            "/srv/webby/webby.log",
		LogLevelPrint:  "all",
		LogLevelRecord: "all",
		DeadPaths:      []string{},
	}
}

// Replaces appropriate fields with default values.
func (opts *ServerOptions) checkForDefaults() {
	if opts.Site == "" {
		opts.Site = DefaultSitePath
	}

	if opts.Site[len(opts.Site)-1] != '/' {
		opts.Site += "/"
	}
}

func (opts *ServerOptions) supportsTLS() bool {
	return opts.Cert != "" && opts.Key != ""
}
