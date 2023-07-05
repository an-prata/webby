// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Holds bit fields representing the different items to show in a log.
type LogLevel uint8

const (
	// Show no logs.
	None LogLevel = 0

	// Show errors.
	Err LogLevel = 1 << (iota - 1)

	// Show warnings.
	Warn

	// Show info.
	Info

	// Shortcut for showing all messages.
	All = Err | Warn | Info
)

const (
	red    string = "\033[31m"
	yellow        = "\033[33m"
	blue          = "\033[34m"
	bold          = "\033[1m"
	normal        = "\033[0m"
)

// Represents a single log that will print to stdout and save to a file.
type Log struct {
	// The log items that will be printed to the console.
	Printing LogLevel

	// Log items that will be saved to the log file.
	Saving LogLevel

	// Pointer to a file for saving log messages, may be nil.
	file *os.File
}

// Produces a log level from a string. The string is not cap-sensitive and must
// be one of "error", "warning", or "info". Some alternative strings will also
// be accepted, such as "err", "war", and "inf" as well as the first character
// of each, "e", "w", and "i". "all" and "none" are accepted as a special case.
// Returns the log level `All` on error.
func LevelFromString(str string) (LogLevel, error) {
	switch strings.ToLower(str) {
	case "none", "n":
		return None, nil
	case "errors", "error", "err", "e":
		return Err, nil
	case "warnings", "warning", "warn", "war", "w":
		return Err | Warn, nil
	case "information", "info", "inf", "i", "all", "a":
		return Err | Warn | Info, nil
	}

	return All, errors.New("Could not parse log level string")
}

// Checks the given uint8 for validity as a log level. If it is invalid an error
// is returned with the All log level.
func CheckLogLevel(level uint8) (LogLevel, error) {
	if level > uint8(All) {
		return All, errors.New("Invalid log level")
	}

	return LogLevel(level), nil
}

// Creates a new log, passing an empty string will create a log with no file and
// will only print messages. This function will never error if the given file
// path is empty.
func NewLog(print LogLevel, save LogLevel, file string) (Log, error) {
	log := Log{print, save, nil}

	if file == "" {
		return log, nil
	}

	f, err := os.Create(file)

	if err == nil {
		log.file = f
	}

	return log, err
}

// Creates a new file or truncates it at the given path.
func (log *Log) OpenFile(path string) error {
	file, err := os.Create(path)

	if err != nil {
		return errors.New("Could not open new log file")
	}

	log.file = file
	return nil
}

// Log a message at the error level.
func (log *Log) LogErr(msg string) error {
	now := time.Now().Format(time.UnixDate)

	if log.Printing&Err == Err {
		fmt.Printf("[%s%sERR%s]  (%s): %s\n", bold, red, normal, now, msg)
	}

	if log.Saving&Err == Err && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[ERR]  (%s): %s\n", now, msg)
		return err
	}

	return nil
}

// Log a message at the warning level.
func (log *Log) LogWarn(msg string) error {
	now := time.Now().Format(time.UnixDate)

	if log.Printing&Warn == Warn {
		fmt.Printf("[%s%sWARN%s] (%s): %s\n", bold, yellow, normal, now, msg)
	}

	if log.Saving&Warn == Warn && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[WARN] (%s): %s\n", now, msg)
		return err
	}

	return nil
}

// Log a message at the info level.
func (log *Log) LogInfo(msg string) error {
	now := time.Now().Format(time.UnixDate)

	if log.Printing&Info == Info {
		fmt.Printf("[%s%sINFO%s] (%s): %s\n", bold, blue, normal, now, msg)
	}

	if log.Saving&Info == Info && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[INFO] (%s): %s\n", now, msg)
		return err
	}

	return nil
}

// Closes the log file, if no file was opened when creating the log then this
// function will simply return no error.
func (log *Log) Close() error {
	if log.file == nil {
		return nil
	}

	if log.file.Sync() != nil {
		return errors.New("Failed to sync log file")
	}

	if log.file.Close() != nil {
		return errors.New("Failed to close log file")
	}

	return nil
}
