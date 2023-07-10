// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type FileChangeSignal = uint8

const (
	ReadError FileChangeSignal = iota
	InitialReadError
	SizeChange
	TimeModifiedChange
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

// Watches for changes in the given file, intended for configs but anything
// should work. This function will report all errors through the given callback.
//
// This function will not call the given callback more than once per detected
// file change and because of this file modification date changes take
// precedence over size changes.
//
// Callback should return true to terminate the goroutine checking for changes
// and false to continue.
func CallOnChange(callback func(FileChangeSignal) bool, filePath string) {
	go callOnChange(callback, filePath)
}

func callOnChange(callback func(FileChangeSignal) bool, filePath string) {
	previousStat, err := os.Stat(filePath)
	shouldReturn := false

	if err != nil {
		shouldReturn = callback(InitialReadError)
	}

	for {
		currentStat, err := os.Stat(filePath)

		if err != nil {
			shouldReturn = callback(ReadError)
			goto Sleep
		}

		if currentStat.ModTime() != previousStat.ModTime() {
			shouldReturn = callback(TimeModifiedChange)
			goto Sleep
		}

		if currentStat.Size() != previousStat.Size() {
			shouldReturn = callback(SizeChange)
			goto Sleep
		}

	Sleep:
		if shouldReturn {
			return
		}

		previousStat = currentStat
		time.Sleep(1 * time.Second)
	}
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

// Returns true if the config has the needed fields populated to support TLS and
// HTTPS connections.
func (opts *ServerOptions) SupportsTLS() bool {
	return opts.Cert != "" && opts.Key != ""
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
