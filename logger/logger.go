// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package logger

import (
	"fmt"
	"os"
)

// Holds bit fields representing the different items to show in a log.
type LogLevel byte

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

// Creates a new log, passing an empty string will create a log with no file and
// will only print messages.
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

// Log a message at the error level.
func (log Log) LogErr(msg string) error {
	if log.Printing&Err == Err {
		fmt.Printf("[%s%sERR%s]:  %s\n", bold, red, normal, msg)
	}

	if log.Saving&Err == Err && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[ERR]:  %s\n", msg)
		return err
	}

	return nil
}

// Log a message at the warning level.
func (log Log) LogWarn(msg string) error {
	if log.Printing&Warn == Warn {
		fmt.Printf("[%s%sWARN%s]: %s\n", bold, yellow, normal, msg)
	}

	if log.Saving&Warn == Warn && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[WARN]: %s\n", msg)
		return err
	}

	return nil
}

// Log a message at the info level.
func (log Log) LogInfo(msg string) error {
	if log.Printing&Info == Info {
		fmt.Printf("[%s%sINFO%s]: %s\n", bold, blue, normal, msg)
	}

	if log.Saving&Info == Info && log.file != nil {
		_, err := fmt.Fprintf(log.file, "[INFO]: %s\n", msg)
		return err
	}

	return nil
}

// Closes the log file, if no file was opened when creating the log then this
// function will simply return no error.
func (log Log) Close() error {
	if log.file == nil {
		return nil
	}

	return log.file.Close()
}
