// Copyright (c) 2025 Evan Overman.
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package main

import (
	"flag"
	"net"
	"strconv"

	"github.com/an-prata/webby/client"
	"github.com/an-prata/webby/daemon"
	"github.com/an-prata/webby/logger"
	"github.com/an-prata/webby/server"
)

func main() {
	const VERSION_MAJOR = 4
	const VERSION_MINOR = 3
	const VERSION_PATCH = 0

	var version bool
	var daemonProc bool
	var start bool
	var reload bool
	var restart bool
	var stop bool
	var status bool
	var genConfig bool
	var logRecord string
	var logPrint string
	var showLog bool

	flag.BoolVar(&version, "version", false, "prints version and exits")
	flag.BoolVar(&daemonProc, client.Daemon, false, "runs the webby server daemon process rather than behaving like a control application")
	flag.BoolVar(&start, client.Start, false, "starts the daemon in a new process and forks it into the background")
	flag.BoolVar(&showLog, client.ShowLog, false, "shows the server log")
	flag.BoolVar(&reload, daemon.Reload, false, "reloads the configuration file and then restarts, this will reset log levels")
	flag.BoolVar(&restart, daemon.Restart, false, "restarts the webby HTTP server, rescanning directories")
	flag.BoolVar(&stop, daemon.Stop, false, "stops the running daemon")
	flag.BoolVar(&status, daemon.Status, false, "gets webby's status by requesting that webby make HTTP get requests to all hosted paths")
	flag.BoolVar(&genConfig, daemon.GenConfig, false, "generated a new default config at '"+daemon.CONFIG_PATH+"'")
	flag.StringVar(&logRecord, daemon.LogRecord, "", "sets the log level to record to file, defaults to 'All'")
	flag.StringVar(&logPrint, daemon.LogPrint, "", "sets the log level to print to standard out, defaults to 'All'")

	flag.Parse()

	if version {
		println("webby version " + strconv.Itoa(VERSION_MAJOR) + "." + strconv.Itoa(VERSION_MINOR) + "." + strconv.Itoa(VERSION_PATCH))
		println("Copyright (c) Evan Overman 2023 - 2025")
		return
	}

	if daemonProc {
		daemon.DaemonMain()
		return
	}

	log, _ := logger.NewLog(logger.All, logger.None, "")

	if genConfig {
		log.LogInfo("Writing default config to '" + daemon.CONFIG_PATH + "'...")

		config := server.DefaultOptions()
		err := config.WriteToFile(daemon.CONFIG_PATH)

		if err != nil {
			log.LogErr(err.Error())
		}

		log.LogInfo("Done!")
		return
	}

	if showLog {
		err := client.ShowLogFile()

		if err != nil {
			log.LogErr("Could not read server log file: " + err.Error())
		}

		return
	}

	if start {
		daemon.StartForkedDaemon(&log)
		return
	}

	socket, err := net.Dial("unix", daemon.SocketPath)

	if err != nil {
		log.LogErr("Could not open Unix Domain Socket, webby may not be running or you may need elevated privileges")

		if status {
			log.LogInfo("webby's daemon uses a Unix Domain Socket for control")
			log.LogInfo("being unable to open the socket likely means webby is not running")
		}

		return
	}

	defer socket.Close()

	daemon.CmdSetLogRecordLevel(socket, &log, logRecord)
	daemon.CmdSetLogPrintLevel(socket, &log, logPrint)
	daemon.CmdRestart(socket, &log, restart)
	daemon.CmdReload(socket, &log, reload)
	daemon.CmdStop(socket, &log, stop)
	daemon.CmdStatus(socket, &log, status)
}
