// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"fmt"
	"net"

	"github.com/an-prata/webby/logger"
)

const SOCKET_PATH = "/run/webby.sock"

type DaemonCommand string

const (
	NONE       DaemonCommand = ""
	RESTART                  = "restart"
	LOG_RECORD               = "log-record"
	LOG_PRINT                = "log-print"
)

type DaemonCommandSuccess uint8

const (
	SUCCESS DaemonCommandSuccess = iota
	FAILURE
)

type DaemonListener struct {
	// The Unix socket by which to listen for incoming commands/requests.
	socket    net.Listener
	callbacks map[DaemonCommand]func(uint8) error
	log       logger.Log
}

// Creates a new Unix Domain Socket and returns a pointer to a listener for
// application commands and requests on that socket.
func NewDaemonListener(callbacks map[DaemonCommand]func(uint8) error, log logger.Log) (DaemonListener, error) {
	socket, err := net.Listen("unix", SOCKET_PATH)
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

	err = daemon.callbacks[DaemonCommand(buf[:n-1])](buf[n-1])

	if err != nil {
		daemon.log.LogErr((fmt.Sprintf("failed to respond to command: %s %d", string(buf[:n-1]), uint8(buf[n-1]))))
		connection.Write([]byte{byte(FAILURE)})
	} else {
		connection.Write([]byte{byte(SUCCESS)})
	}
}
