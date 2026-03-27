// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"testing"
)

// TestMultiSessionFwmarkIsolation verifies two interfaces get distinct fwmarks.
func TestMultiSessionFwmarkIsolation(t *testing.T) {
	sessions := []struct {
		iface   string
		ifindex int
	}{
		{"wg0", 7},
		{"tun0", 12},
		{"ppp0", 23},
	}

	seen := make(map[uint32]string)
	for _, s := range sessions {
		m := allocFwmark(s.ifindex)
		if prev, dup := seen[m]; dup {
			t.Errorf("fwmark collision: %s (idx %d) and %s both got 0x%X",
				s.iface, s.ifindex, prev, m)
		}
		seen[m] = s.iface
	}
}

// TestMultiSessionRouteTableIsolation verifies two interfaces get distinct tables.
func TestMultiSessionRouteTableIsolation(t *testing.T) {
	sessions := []struct {
		iface   string
		ifindex int
	}{
		{"wg0", 7},
		{"tun0", 12},
		{"ppp0", 23},
	}

	seen := make(map[int]string)
	for _, s := range sessions {
		tbl := allocRouteTable(s.ifindex)
		if prev, dup := seen[tbl]; dup {
			t.Errorf("table collision: %s (idx %d) and %s both got %d",
				s.iface, s.ifindex, prev, tbl)
		}
		seen[tbl] = s.iface
	}
}

// TestSameInterfaceSameMark verifies concurrent sessions on the same
// interface get the same fwmark (sharing the same routing table is correct).
func TestSameInterfaceSameMark(t *testing.T) {
	m1 := allocFwmark(7)
	m2 := allocFwmark(7)
	if m1 != m2 {
		t.Errorf("same ifindex=7 produced different fwmarks: 0x%X vs 0x%X", m1, m2)
	}

	t1 := allocRouteTable(7)
	t2 := allocRouteTable(7)
	if t1 != t2 {
		t.Errorf("same ifindex=7 produced different tables: %d vs %d", t1, t2)
	}
}
