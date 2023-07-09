// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/an-prata/webby/logger"
)

// The path of the Unix Domain Socket created by webby for accepting commands.
const SocketPath = "/run/webby.sock"

type DaemonListener struct {
	// The Unix socket by which to listen for incoming commands/requests.
	socket net.Listener

	// A map of daemon commands to their callbacks. The passed in argument will
	// always be the last byte read from the Unix Domain Socket and the command
	// should be everything up to that.
	callbacks map[DaemonCommand]DaemonCommandCallback

	shuttingOff bool

	// Channel for blocking the `Close()` function to prevent bad memory access.
	shuttoffChannel chan bool

	// Webby log.
	log logger.Log
}

// Creates a new Unix Domain Socket and returns a pointer to a listener for
// application commands and requests on that socket. When the listener is
// started all commands will be executed according to the given callbacks.
func NewDaemonListener(callbacks map[DaemonCommand]DaemonCommandCallback, log logger.Log) (DaemonListener, error) {
	os.Remove(SocketPath)
	socket, err := net.Listen("unix", SocketPath)
	shutoffChannel := make(chan bool, 1)
	return DaemonListener{socket, callbacks, false, shutoffChannel, log}, err
}

// Starts listening for connections on the Unix Domain Socket. Each connection
// will be able to run one command and will be responded to with a
// `DaemonCommandSuccess` value.
func (daemon *DaemonListener) Listen() error {
	var wg sync.WaitGroup

	for {
		connection, err := daemon.socket.Accept()

		if err != nil && !daemon.shuttingOff {
			daemon.log.LogErr("Failed to accept daemon connection")
			return err
		} else if daemon.shuttingOff {
			break
		}

		wg.Add(1)
		go daemon.handleConnection(connection, &wg)
	}

	daemon.log.LogInfo("Waiting for connections to close...")
	wg.Wait()
	daemon.shuttoffChannel <- true
	return nil
}

// Closes the backing Unix Domain Socket. After calling this function no other
// calls should be made to this struct's functions.
func (daemon *DaemonListener) Close() error {
	daemon.shuttingOff = true
	// _ = <-daemon.shuttoffChannel
	if daemon.socket == nil {
		daemon.log.LogErr("socket was nil")
	}
	return daemon.socket.Close()
}

// Handles an individual connection from the Unix Domain Socket.
func (daemon *DaemonListener) handleConnection(connection net.Conn, wg *sync.WaitGroup) {
	defer connection.Close()
	defer wg.Done()

	var buf [526]byte
	n, err := connection.Read(buf[:])

	if err != nil {
		daemon.log.LogErr("Could not read from daemon connection")
		return
	}

	fn, ok := daemon.callbacks[DaemonCommand(buf[:n-1])]

	if !ok {
		daemon.log.LogErr("No callback for requested daemon command " + string(buf[:n-1]))
	}

	err = fn(DaemonCommandArg(buf[n-1]))

	if err != nil {
		daemon.log.LogErr((fmt.Sprintf("Failed to respond to command: %s %d", string(buf[:n-1]), uint8(buf[n-1]))))
		connection.Write([]byte{byte(Failure)})
	} else {
		connection.Write([]byte{byte(Success)})
	}
}
