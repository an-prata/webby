// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"net"

	"github.com/an-prata/webby/logger"
)

// Sends a command, using the given command line argument, to the daemon using
// the provided socket.
//
// This function is intended as the end of execution for the command it
// represents and will therefore perform I/O operations, output to the user, and
// indicate errors only though these means.
func CmdSetLogRecordLevel(socket net.Conn, log *logger.Log, arg string) {
	if arg == "" {
		return
	}

	logLevel, err := logger.LevelFromString(arg)

	if err != nil {
		log.LogErr("Could not identify log level from given argument (" + arg + ")")
		log.LogInfo("try using 'error', 'warning', 'info', or 'all'")
		return
	}

	var buf [1]byte
	socket.Write(append([]byte(LogRecord), byte(logLevel)))
	socket.Read(buf[:])

	if DaemonCommandSuccess(buf[0]) != Success {
		log.LogErr("Could not change log level for recording")
	} else {
		log.LogInfo("Log level for recording changed to '" + arg + "'")
	}
}

// Sends the set print log level command to the daemon, using the given command
// line argument, through the provided socket.
//
// This function is intended as the end of execution for the command it
// represents and will therefore perform I/O operations, output to the user, and
// indicate errors only though these means.
func CmdSetLogPrintLevel(socket net.Conn, log *logger.Log, arg string) {
	if arg == "" {
		return
	}

	logLevel, err := logger.LevelFromString(arg)

	if err != nil {
		log.LogErr("Could not identify log level from given argument (" + arg + ")")
		log.LogInfo("try using 'error', 'warning', 'info', or 'all'")
		return
	}

	var buf [1]byte
	socket.Write(append([]byte(LogPrint), byte(logLevel)))
	socket.Read(buf[:])

	if DaemonCommandSuccess(buf[0]) != Success {
		log.LogErr("Could not change log level for printing")
	} else {
		log.LogInfo("Log level for printing changed to '" + arg + "'")
	}
}

// Sends the reload command to the daemon through the provided socket.
//
// This function is intended as the end of execution for the command it
// represents and will therefore perform I/O operations, output to the user, and
// indicate errors only though these means.
func CmdReload(socket net.Conn, log *logger.Log, arg bool) {
	if !arg {
		return
	}

	log.LogInfo("Reloading config and restarting webby...")

	var buf [1]byte
	socket.Write(append([]byte(Reload), 0))
	socket.Read(buf[:])

	if DaemonCommandSuccess(buf[0]) != Success {
		log.LogErr("Could not reload config or restart")
	} else {
		log.LogInfo("Reloaded and restarted!")
	}
}

// Sends the restart command to the daemon through the provided socket.
//
// This function is intended as the end of execution for the command it
// represents and will therefore perform I/O operations, output to the user, and
// indicate errors only though these means.
func CmdRestart(socket net.Conn, log *logger.Log, arg bool) {
	if !arg {
		return
	}

	log.LogInfo("Restarting webby...")

	var buf [1]byte
	socket.Write(append([]byte(Restart), 0))
	socket.Read(buf[:])

	if DaemonCommandSuccess(buf[0]) != Success {
		log.LogErr("Could not restart webby correctly")
	} else {
		log.LogInfo("Restarted!")
	}
}
