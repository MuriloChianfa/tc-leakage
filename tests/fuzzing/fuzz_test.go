// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package fuzzing

import (
	"net"
	"strings"
	"testing"
	"unicode/utf8"
)

const (
	fwmarkBase     = 0x4E4C0000
	routeTableBase = 100
	routeTableMax  = 252
)

func allocFwmark(ifindex int) uint32 {
	idx := uint16(ifindex)
	return uint32(fwmarkBase) | uint32(idx)
}

func allocRouteTable(ifindex int) int {
	span := routeTableMax - routeTableBase + 1
	idx := int(uint16(ifindex))
	return routeTableBase + (idx % span)
}

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

// FuzzAllocFwmark tests that fwmark allocation never panics and always
// produces valid (non-zero, fits uint32, preserves prefix) results.
func FuzzAllocFwmark(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(42)
	f.Add(255)
	f.Add(65535)
	f.Add(-1)
	f.Add(1 << 20)

	f.Fuzz(func(t *testing.T, ifindex int) {
		m := allocFwmark(ifindex)
		if m == 0 {
			t.Errorf("allocFwmark(%d) = 0", ifindex)
		}
		if m>>16 != 0x4E4C {
			t.Errorf("allocFwmark(%d) prefix = 0x%X, want 0x4E4C", ifindex, m>>16)
		}
	})
}

// FuzzAllocRouteTable tests that table allocation never panics and
// always returns values in the valid [1, 252] range.
func FuzzAllocRouteTable(f *testing.F) {
	f.Add(0)
	f.Add(1)
	f.Add(42)
	f.Add(255)
	f.Add(65535)
	f.Add(-1)
	f.Add(1 << 20)

	f.Fuzz(func(t *testing.T, ifindex int) {
		tbl := allocRouteTable(ifindex)
		if tbl < 1 || tbl > 252 {
			t.Errorf("allocRouteTable(%d) = %d, want [1, 252]", ifindex, tbl)
		}
	})
}

// FuzzGatewayIPParsing tests that gateway IP parsing never panics.
func FuzzGatewayIPParsing(f *testing.F) {
	f.Add("10.0.0.1")
	f.Add("192.168.1.1")
	f.Add("fd10::1")
	f.Add("::1")
	f.Add("")
	f.Add("not-an-ip")
	f.Add("256.0.0.1")
	f.Add(strings.Repeat("1", 100))

	f.Fuzz(func(t *testing.T, input string) {
		ip := net.ParseIP(input)
		// Just verify no panic; ip may be nil for invalid inputs
		_ = ip
	})
}

// FuzzSubnetCIDRParsing tests that CIDR parsing never panics.
func FuzzSubnetCIDRParsing(f *testing.F) {
	f.Add("10.10.0.0/24")
	f.Add("192.168.0.0/16")
	f.Add("fd10::/64")
	f.Add("::/0")
	f.Add("0.0.0.0/0")
	f.Add("")
	f.Add("not-a-cidr")
	f.Add("10.10.0.0")
	f.Add("10.10.0.0/33")

	f.Fuzz(func(t *testing.T, input string) {
		_, subnet, err := net.ParseCIDR(input)
		if err != nil {
			return
		}
		// If parsing succeeded, verify Contains doesn't panic
		_ = subnet.Contains(net.IPv4(10, 10, 0, 1))
		_ = subnet.Contains(net.ParseIP("fd10::1"))
	})
}
