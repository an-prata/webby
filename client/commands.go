// Copyright (c) 2025 Evan Overman.
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package client

import (
	"os"

	"github.com/an-prata/webby/daemon"
	"github.com/an-prata/webby/server"
)

const (
	// Runs the daemon proccess.
	Daemon = "daemon"

	// Starts the daemon process much like `Daemon` but forks the process into the
	// background.
	Start = "start"

	// Reads the server log file and outputs it to the console.
	ShowLog = "show-log"
)

func ShowLogFile() error {
	opts, err := server.LoadConfigFromPath(daemon.CONFIG_PATH)

	if err != nil {
		return err
	}

	buf, err := os.ReadFile(opts.Log)

	if err != nil {
		return err
	}

	print(string(buf))

	return nil
}
