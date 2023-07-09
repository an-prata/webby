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
	var start bool
	var reload bool
	var restart bool
	var stop bool
	var logRecord string
	var logPrint string

	flag.BoolVar(&daemonProc, daemon.Daemon, false, "runs the webby server daemon process rather than behaving like a control application")
	flag.BoolVar(&start, daemon.Start, false, "starts the daemon in a new process and forks it into the background")
	flag.BoolVar(&reload, daemon.Reload, false, "reloads the configuration file and then restarts, this will reset log levels")
	flag.BoolVar(&restart, daemon.Restart, false, "restarts the webby HTTP server, rescanning directories")
	flag.BoolVar(&stop, daemon.Stop, false, "stops the running daemon")
	flag.StringVar(&logRecord, daemon.LogRecord, "", "sets the log level to record to file, defaults to 'All'")
	flag.StringVar(&logPrint, daemon.LogPrint, "", "sets the log level to print to standard out, defaults to 'All'")

	flag.Parse()

	if daemonProc {
		daemon.DaemonMain()
		return
	}

	log, _ := logger.NewLog(logger.All, logger.None, "")

	if start {
		daemon.StartForkedDaemon(&log)
		return
	}

	socket, err := net.Dial("unix", daemon.SocketPath)

	if err != nil {
		log.LogErr("Could not open Unix Domain Socket, you may need elevated privileges")
		return
	}

	defer socket.Close()

	daemon.CmdSetLogRecordLevel(socket, &log, logRecord)
	daemon.CmdSetLogPrintLevel(socket, &log, logPrint)
	daemon.CmdRestart(socket, &log, restart)
	daemon.CmdReload(socket, &log, reload)
	daemon.CmdStop(socket, &log, stop)
}
