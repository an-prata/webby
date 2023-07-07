// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
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

	// Whether or not to check for changes in the config and reload automatically.
	AutoReload bool

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

// Get the default configuration.
func DefaultOptions() ServerOptions {
	return ServerOptions{
		Site:           "/srv/webby/website",
		Cert:           "",
		Key:            "",
		Port:           -1,
		Log:            "/srv/webby/webby.log",
		LogLevelPrint:  "all",
		LogLevelRecord: "all",
		AutoReload:     true,
		DeadPaths:      []string{},
	}
}

// Returns true if both configurations are equal.
func (lhs *ServerOptions) Equals(rhs *ServerOptions) bool {
	// The lambda function is here because if possible we would like to avoid
	// making string comparrisons in a loop. By moving it to the end in a lambda
	// it has a chance of the and operator short circuting, saving us a potentially
	// arbatrary number of string comparisons.
	return strings.EqualFold(lhs.Site, rhs.Site) &&
		strings.EqualFold(lhs.Cert, rhs.Cert) &&
		strings.EqualFold(lhs.Key, rhs.Key) &&
		lhs.Port == rhs.Port &&
		strings.EqualFold(lhs.Log, rhs.Log) &&
		strings.EqualFold(lhs.LogLevelPrint, rhs.LogLevelRecord) &&
		lhs.AutoReload == rhs.AutoReload &&
		func() bool {
			sameDeadPaths := len(lhs.DeadPaths) == len(rhs.DeadPaths)

			for i := 0; sameDeadPaths && i < len(lhs.DeadPaths); i++ {
				sameDeadPaths = strings.EqualFold(lhs.DeadPaths[i], rhs.DeadPaths[i])
			}

			return sameDeadPaths
		}()
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
