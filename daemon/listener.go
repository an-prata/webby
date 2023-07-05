// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"fmt"
	"net"

	"github.com/an-prata/webby/logger"
)

const SocketPath = "/run/webby.sock"

type DaemonCommand string

const (
	// The `None` variant here shouldn't really be used.
	None      DaemonCommand = ""
	Restart                 = "restart"
	LogRecord               = "log-record"
	LogPrint                = "log-print"
)

type DaemonCommandArg uint8

type DaemonCommandSuccess uint8

const (
	Success DaemonCommandSuccess = iota
	Failure
)

type DaemonListener struct {
	// The Unix socket by which to listen for incoming commands/requests.
	socket    net.Listener
	callbacks map[DaemonCommand]func(DaemonCommandArg) error
	log       logger.Log
}

// Creates a new Unix Domain Socket and returns a pointer to a listener for
// application commands and requests on that socket. When the listener is
// started all commands will be executed according to the given callbacks.
func NewDaemonListener(callbacks map[DaemonCommand]func(DaemonCommandArg) error, log logger.Log) (DaemonListener, error) {
	socket, err := net.Listen("unix", SocketPath)
	return DaemonListener{socket, callbacks, log}, err
}

func (daemon *DaemonListener) Listen() error {
	for {
		connection, err := daemon.socket.Accept()

		if err != nil {
			daemon.log.LogErr("failed to accept daemon connection")
			return err
		}

		go daemon.handleConnection(connection)
	}
}

func (daemon *DaemonListener) handleConnection(connection net.Conn) {
	defer connection.Close()

	var buf [526]byte
	n, err := connection.Read(buf[:])

	if err != nil {
		daemon.log.LogErr("could not read from daemon connection")
		return
	}

	fn, ok := daemon.callbacks[DaemonCommand(buf[:n-1])]

	if !ok {
		daemon.log.LogErr("No callback for requested daemon command " + string(buf[:n-1]))
	}

	err = fn(DaemonCommandArg(buf[n-1]))

	if err != nil {
		daemon.log.LogErr((fmt.Sprintf("failed to respond to command: %s %d", string(buf[:n-1]), uint8(buf[n-1]))))
		connection.Write([]byte{byte(Failure)})
	} else {
		connection.Write([]byte{byte(Success)})
	}
}
