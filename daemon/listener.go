// Copyright (c) 2023 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package daemon

import (
	"bytes"
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

type RescanCallback interface {
	Rescan() error
}

type LogRecordCallback interface {
	LogRecord(logger.LogLevel) error
}

type LogPrintCallback interface {
	LogPrint(logger.LogLevel) error
}

type DaemonListener struct {
	// The Unix socket by which to listen for incoming commands/requests.
	socket net.Listener

	rescanCallback    RescanCallback
	logRecordCallback LogRecordCallback
	logPrintCallback  LogPrintCallback
}

// Creates a new Unix Domain Socket and returns a pointer to a listener for
// application commands and requests on that socket.
func NewDaemonListener(rescan RescanCallback, logRecord LogRecordCallback, logPrint LogPrintCallback) (*DaemonListener, error) {
	socket, err := net.Listen("unix", SOCKET_PATH)

	if err != nil {
		return nil, err
	}

	return &DaemonListener{socket, rescan, logRecord, logPrint}, nil
}

func (daemon *DaemonListener) Listen(log logger.Log) error {
	for {
		connection, err := daemon.socket.Accept()

		if err != nil {
			log.LogErr("failed to accept daemon connection")
			return err
		}

		go func(connection net.Conn) {
			defer connection.Close()

			var buf [526]byte
			n, err := connection.Read(buf[:])

			if err != nil {
				log.LogErr("could not read from daemon connection")
				return
			}

			if bytes.Equal(buf[:n], []byte(RESTART)) {
				err = daemon.rescanCallback.Rescan()

				if err != nil {
					log.LogErr("failed to restart")
					_, err = connection.Write([]byte{byte(FAILURE)})
				} else {
					_, err = connection.Write([]byte{byte(SUCCESS)})
				}

				if err != nil {
					log.LogErr("failed to reply to restart command")
				}
			}

			if bytes.Equal(buf[:n-1], []byte(LOG_RECORD)[:]) {
				// Use the last byte as log level
				err = daemon.logRecordCallback.LogRecord(logger.LogLevel(buf[n-1]))

				if err != nil {
					log.LogErr("failed to set log record level")
					_, err = connection.Write([]byte{byte(FAILURE)})
				} else {
					_, err = connection.Write([]byte{byte(SUCCESS)})
				}

				if err != nil {
					log.LogErr("failed to reply to log recording level set")
				}
			}

			if bytes.Equal(buf[:n-1], []byte(LOG_PRINT)) {
				// Use the last byte as log level
				err = daemon.logPrintCallback.LogPrint(logger.LogLevel(buf[n-1]))

				if err != nil {
					log.LogErr("failed to set log printing level")
					_, err = connection.Write([]byte{byte(FAILURE)})
				} else {
					_, err = connection.Write([]byte{byte(SUCCESS)})
				}

				if err != nil {
					log.LogErr("failed to reply to log printing level set")
				}
			}
		}(connection)
	}
}
