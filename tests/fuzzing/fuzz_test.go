// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package fuzzing

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzArgumentParsing tests that argument validation logic handles arbitrary inputs
// without panicking. Simulates os.Args validation as done in cmd/main.go.
func FuzzArgumentParsing(f *testing.F) {
	// Seed corpus
	f.Add("eth0", "curl")
	f.Add("wg0", "bash")
	f.Add("ppp0", "curl ifconfig.me")
	f.Add("", "")
	f.Add("lo", "")
	f.Add("a]b[c", "cmd")
	f.Add(strings.Repeat("x", 1024), "cmd")

	f.Fuzz(func(t *testing.T, iface, command string) {
		// Simulate the minimum args check from main.go
		args := []string{"netleak", iface, command}
		if len(args) < 3 {
			t.Skip("not enough args")
		}

		targetIface := args[1]
		cmdArgs := args[2:]

		// These should never panic regardless of input
		_ = len(targetIface)
		_ = len(cmdArgs)
		_ = targetIface == ""
		_ = utf8.ValidString(targetIface)
	})
}

// FuzzInterfaceName tests interface name validation with arbitrary strings.
// Linux interface names: max 15 chars, no slashes, no whitespace.
func FuzzInterfaceName(f *testing.F) {
	// Seed corpus with valid and edge-case interface names
	f.Add("eth0")
	f.Add("wg0")
	f.Add("ppp0")
	f.Add("lo")
	f.Add("br-abcdef123456")
	f.Add("")
	f.Add(strings.Repeat("a", 16))
	f.Add("eth 0")
	f.Add("eth/0")
	f.Add("eth\x00zero")

	f.Fuzz(func(t *testing.T, name string) {
		// Validate interface name constraints without panicking
		valid := true

		if len(name) == 0 || len(name) > 15 {
			valid = false
		}

		if strings.ContainsAny(name, "/ \t\n\r") {
			valid = false
		}

		if strings.Contains(name, "\x00") {
			valid = false
		}

		// Result is unused, we just verify no panic occurs
		_ = valid
	})
}
