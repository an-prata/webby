// Copyright (c) 2024 Evan Overman (https://an-prata.it).
// Licensed under the MIT License.
// See LICENSE file in repository root for complete license text.

package client

const (
	// Runs the daemon proccess.
	Daemon = "daemon"

	// Starts the daemon process much like `Daemon` but forks the process into the
	// background.
	Start = "start"

	// Reads the server log file and outputs it to the console.
	ShowLog = "show-log"
)
