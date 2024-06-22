// Copyright (c) 2024 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package server

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/an-prata/webby/logger"
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

	// Whether or not to check for changes in the config or site files and reload
	// automatically.
	AutoReload bool

	// Paths that should be granted a dead response, can be used for fucking with
	// bot probing or the like. A dead response is just the name I gave to
	// redirecting a request back onto the client for the same path.
	DeadPaths []string

	// Redirect automatically from HTTP to HTTPS.
	RedirectHttp bool

	// Response write timeout in seconds.
	WriteTimeout int64

	// Request read timeout in seconds.
	ReadTimeout int64
}

// Tries to parse JSON for a `ServerOptions` with the file at the given path.
// Returns an error and a default configuration on parse failure, individual
// options are replaced by defaults for incorrect types and absences.
func LoadConfigFromPath(path string) (ServerOptions, error) {
	if _, err := os.Stat(path); err != nil {
		return DefaultOptions(), errors.New("Could not stat config at '" + path + "'")
	}

	var optsMap map[string]interface{}
	opts := DefaultOptions()

	bytes, err := os.ReadFile(path)

	if err != nil {
		return DefaultOptions(), errors.New("Could not read config at '" + path + "'")
	}

	if json.Unmarshal(bytes, &optsMap) != nil {
		return DefaultOptions(), errors.New("Could not parse config JSON at '" + path + "'")
	}

	for k, v := range optsMap {
		switch k {
		case "Site":
			if value, ok := v.(string); ok {
				opts.Site = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'Site' field in config to be a string.")
			}
		case "Cert":
			if value, ok := v.(string); ok {
				opts.Cert = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'Cert' field in config to be a string.")
			}
		case "Key":
			if value, ok := v.(string); ok {
				opts.Key = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'Key' field in config to be a string.")
			}
		case "Port":
			if value, ok := v.(float64); ok {
				opts.Port = int32(value)
			} else {
				logger.GlobalLog.LogWarn("Expected 'Port' field in config to be a number.")
			}
		case "Log":
			if value, ok := v.(string); ok {
				opts.Log = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'Log' field in config to be a string.")
			}
		case "LogLevelPrint":
			if value, ok := v.(string); ok {
				opts.LogLevelPrint = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'LogLevelPrint' field in config to be a string.")
			}
		case "LogLevelRecord":
			if value, ok := v.(string); ok {
				opts.LogLevelRecord = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'LogLevelRecord' field in config to be a string.")
			}
		case "AutoReload":
			if value, ok := v.(bool); ok {
				opts.AutoReload = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'AutoReload' field in config to be a bool.")
			}
		case "DeadPaths":
			if value, ok := v.([]interface{}); ok {
				for _, path := range value {
					if p, ok := path.(string); ok {
						opts.DeadPaths = append(opts.DeadPaths, p)
					} else {
						logger.GlobalLog.LogWarn("Expected all elements of 'DeadPaths' to be strings")
					}
				}
			} else {
				logger.GlobalLog.LogWarn("Expected 'DeadPaths' field in config to be a list of strings.")
			}
		case "RedirectHttp":
			if value, ok := v.(bool); ok {
				opts.RedirectHttp = value
			} else {
				logger.GlobalLog.LogWarn("Expected 'RedirectHttp' field in config to be a bool.")
			}
		case "WriteTimeout":
			if value, ok := v.(float64); ok {
				opts.WriteTimeout = int64(value)
			} else {
				logger.GlobalLog.LogWarn("Expected 'WriteTimeout' field in config to be a number.")
			}
		case "ReadTimeout":
			if value, ok := v.(float64); ok {
				opts.ReadTimeout = int64(value)
			} else {
				logger.GlobalLog.LogWarn("Expected 'ReadTimout' field in config to be a number.")
			}
		}
	}

	return opts, nil
}

// Prints log options to the info log.
func (opts *ServerOptions) Show() {
	logger.GlobalLog.LogInfo("Config: Site: " + opts.Site)
	logger.GlobalLog.LogInfo("Config: Cert: " + opts.Cert)
	logger.GlobalLog.LogInfo("Config: Key: " + opts.Key)
	logger.GlobalLog.LogInfo("Config: Port: " + strconv.FormatInt(int64(opts.Port), 10))
	logger.GlobalLog.LogInfo("Config: Log: " + opts.Log)
	logger.GlobalLog.LogInfo("Config: LogLevelPrint: " + opts.LogLevelPrint)
	logger.GlobalLog.LogInfo("Config: LogLevelRecord: " + opts.LogLevelRecord)
	logger.GlobalLog.LogInfo("Config: AutoReload: " + strconv.FormatBool(opts.AutoReload))
	logger.GlobalLog.LogInfo("Config: RedirectHttp: " + strconv.FormatBool(opts.RedirectHttp))
	logger.GlobalLog.LogInfo("Config: WriteTimeout: " + strconv.FormatInt(int64(opts.WriteTimeout), 10))
	logger.GlobalLog.LogInfo("Config: ReadTimeout: " + strconv.FormatInt(int64(opts.ReadTimeout), 10))
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
		WriteTimeout:   60,
		ReadTimeout:    60,
	}
}

func (opts *ServerOptions) WriteToFile(path string) error {
	json_string, err := json.MarshalIndent(opts, "", "    ")

	if err != nil {
		return errors.New("Failed to parse ServerOptions into JSON: " + err.Error())
	}

	file, err := os.Create(path)

	if err != nil {
		return errors.New("Could not create file '" + path + "': " + err.Error())
	}

	_, err = file.Write(json_string)

	if err != nil {
		return errors.New("Could not write to file '" + path + "': " + err.Error())
	}

	if err = file.Close(); err != nil {
		return errors.New("Could not close file '" + path + "': " + err.Error())
	}

	return nil
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
