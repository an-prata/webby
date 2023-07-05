// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"flag"
	"net"

	"github.com/an-prata/webby/daemon"
	"github.com/an-prata/webby/logger"
)

func main() {
	var daemonProc bool
	var restart bool
	var logRecord string
	var logPrint string

	flag.BoolVar(&daemonProc, "daemon", false, "runs the webby server daemon process rather than behaving like a control application")
	flag.BoolVar(&restart, "restart", false, "restarts the webby HTTP server, rescanning directories")
	flag.StringVar(&logRecord, "log-record", "", "the log level to record to file, defaults to 'All'")
	flag.StringVar(&logPrint, "log-print", "", "the log level to print to standard out, defaults to 'All'")

	flag.Parse()

	if daemonProc {
		daemon.DaemonMain()
		return
	}

	log, _ := logger.NewLog(logger.All, logger.None, "")
	socket, err := net.Dial("unix", daemon.SocketPath)

	if err != nil {
		log.LogErr("Could not open Unix Domain Socket, you may need elevated privileges")
		return
	}

	defer socket.Close()

	if logRecord != "" {
		logLevel, err := logger.LevelFromString(logRecord)

		if err != nil {
			log.LogErr("Could not identify log level from given argument (" + logRecord + ")")
			log.LogInfo("try using 'error', 'warning', 'info', or 'all'")
			return
		}

		var buf [1]byte
		socket.Write(append([]byte(daemon.LogRecord), byte(logLevel)))
		socket.Read(buf[:])

		if daemon.DaemonCommandSuccess(buf[0]) != daemon.Success {
			log.LogErr("Could not change log level for recording")
		} else {
			log.LogInfo("Log level for recording changed to '" + logRecord + "'")
		}
	}

	if logPrint != "" {
		logLevel, err := logger.LevelFromString(logPrint)

		if err != nil {
			log.LogErr("Could not identify log level from given argument (" + logPrint + ")")
			log.LogInfo("try using 'error', 'warning', 'info', or 'all'")
			return
		}

		var buf [1]byte
		socket.Write(append([]byte(daemon.LogPrint), byte(logLevel)))
		socket.Read(buf[:])

		if daemon.DaemonCommandSuccess(buf[0]) != daemon.Success {
			log.LogErr("Could not change log level for printing")
		} else {
			log.LogInfo("Log level for printing changed to '" + logPrint + "'")
		}
	}

	if restart {
		log.LogInfo("Restarting webby...")

		var buf [1]byte
		socket.Write(append([]byte(daemon.Restart), 0))
		socket.Read(buf[:])

		if daemon.DaemonCommandSuccess(buf[0]) != daemon.Success {
			log.LogErr("Could not restart webby correctly")
		} else {
			log.LogInfo("Restarted!")
		}
	}
}
