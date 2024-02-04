// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"

	"github.com/an-prata/webby/logger"
)

// Represents possible commands from client connections.
type DaemonCommand string

const (
	// Restarts the HTTP server and rescans directories. Useful when edits have
	// been made to the website contents. Should ignore the passed in argument.
	Restart = "restart"

	// Reloads the configuration file and then restarts.
	Reload = "reload"

	// Stops the current daemon.
	Stop = "stop"

	// Gets webby's current status, status needs to be representable in the single
	// byte returned to the client after a daemon command.
	Status = "status"

	// Like the Daemon variant this variant should not have a callback, and is
	// intended to generate a default configuration file.
	GenConfig = "gen-config"

	// Sets the log level for recording logs to file. Should interperet its
	// argument to be the desired log level.
	LogRecord = "log-record"

	// Sets the log level for printing to standard out. As a daemonized program
	// this will be what shows up when checking the output of `# systemctl status
	// webby`. Should interperet its argument to be the desired log level.
	LogPrint = "log-print"
)

const maximumSocketChecks = 10

// Starts a daemon process and forks it.
func StartForkedDaemon(log *logger.Log) {
	user, err := user.Current()

	if err != nil {
		log.LogErr("Could not get information on the current user")
		return
	}

	// Base-10 and 32 bit.
	uid, err := strconv.ParseUint(user.Uid, 10, 32)

	if err != nil {
		log.LogErr("Could not parse UID from '" + user.Uid + "'")
		return
	}

	gid, err := strconv.ParseInt(user.Gid, 10, 32)

	if err != nil {
		log.LogErr("Could not parse GID from '" + user.Gid + "'")
		return
	}

	cred := syscall.Credential{
		Uid:         uint32(uid),
		Gid:         uint32(gid),
		Groups:      []uint32{},
		NoSetGroups: true,
	}

	sysproc := syscall.SysProcAttr{
		Credential: &cred,
		Noctty:     true,
	}

	attr := os.ProcAttr{
		Dir: ".",
		Env: os.Environ(),
		Files: []*os.File{
			os.Stdin,
			nil,
			nil,
		},
		Sys: &sysproc,
	}

	bin, err := exec.LookPath(os.Args[0])

	if err != nil {
		log.LogErr("Could not find webby binary")
	}

	log.LogInfo("Found webby binary (" + bin + ")...")
	log.LogInfo("Starting process...")

	proc, err := os.StartProcess(
		bin,
		[]string{os.Args[0], "-daemon"},
		&attr,
	)

	if err != nil {
		log.LogErr("Failed to start system process")
		log.LogErr(err.Error())
		return
	}

	// Detatch new process from parent.
	log.LogInfo("Detatching process from parent...")
	err = proc.Release()

	if err != nil {
		log.LogErr("Could not detatch process")
		return
	}

	log.LogInfo("Waiting for webby daemon process to respond...")

	for i := 0; i < maximumSocketChecks; i++ {
		socket, err := net.Dial("unix", SocketPath)

		if err == nil {
			socket.Close()
			log.LogInfo("Started webby!")
			return
		}

		time.Sleep(1 * time.Second)
	}

	log.LogErr("Could create a process but webby failed to start, you may need elevated permissions")
}

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

// Sends the stop command to the daemon through the provided socket.
//
// This function is intended as the end of execution for the command it
// represents and will therefore perform I/O operations, output to the user, and
// indicate errors only though these means.
func CmdStop(socket net.Conn, log *logger.Log, arg bool) {
	if !arg {
		return
	}

	log.LogInfo("Stopping webby...")

	var buf [1]byte
	socket.Write(append([]byte(Stop), 0))
	socket.Read(buf[:])

	if DaemonCommandSuccess(buf[0]) != Success {
		log.LogErr("Could not stop webby")
	} else {
		log.LogInfo("Stopped!")
	}
}

func CmdStatus(socket net.Conn, log *logger.Log, arg bool) {
	if !arg {
		return
	}

	log.LogInfo("Requesting status from webby..")

	var buf [1]byte
	socket.Write(append([]byte(Status), 0))
	socket.Read(buf[:])

	status := WebbyStatus(buf[0])

	log.LogInfo("Got status!")

	print("\nstatus: ")

	if status == Ok {
		println("OK\n")
		println("webby made HTTP GET requests to all hosted paths and got 200 for each.\n")
		return
	}

	if status == HttpNon2xx {
		println("Non 200\n")
		println("webby made HTTP GET requests to all hosted paths, all responded but some did not give 200.\n")
		return
	}

	if status == HttpPartialFail {
		println("Partial Fail\n")
		println("webby made HTTP GET requests to all hosted paths but some responded with a failure code, e.g. 400.\n")
		return
	}

	if status == HttpFail {
		println("Fail\n")
		println("webby made HTTP GET requests to all hosted paths and all responded with a failure code, e.g. 400.\n")
		return
	}
}
