// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const sessionDir = "/run/netleak"

// acquireSession creates a session file under /run/netleak/<pid> that
// records the interface name, fwmark, and route table for this process.
// This prevents collisions and lets concurrent sessions be inspected.
func acquireSession(pid int, iface string, mark uint32, table int) error {
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("session dir: %w", err)
	}
	path := filepath.Join(sessionDir, strconv.Itoa(pid))
	data := fmt.Sprintf("iface=%s\nfwmark=0x%X\ntable=%d\n", iface, mark, table)
	return os.WriteFile(path, []byte(data), 0644)
}

// releaseSession removes the session file for the given PID.
func releaseSession(pid int) {
	path := filepath.Join(sessionDir, strconv.Itoa(pid))
	_ = os.Remove(path)
}
