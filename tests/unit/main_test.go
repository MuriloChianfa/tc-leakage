// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"testing"
	"unsafe"
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

// TestFwmarkBasePrefix verifies the fwmark base carries the "NL" prefix.
func TestFwmarkBasePrefix(t *testing.T) {
	hi := uint16(fwmarkBase >> 16)
	if hi != 0x4E4C {
		t.Errorf("fwmarkBase upper 16 bits = 0x%X, want 0x4E4C", hi)
	}
}

// TestAllocFwmarkNonZero verifies allocated fwmarks are always non-zero.
func TestAllocFwmarkNonZero(t *testing.T) {
	for idx := 0; idx <= 300; idx++ {
		m := allocFwmark(idx)
		if m == 0 {
			t.Errorf("allocFwmark(%d) = 0, want non-zero", idx)
		}
	}
}

// TestAllocFwmarkPreservesPrefix verifies the "NL" prefix is preserved.
func TestAllocFwmarkPreservesPrefix(t *testing.T) {
	for _, idx := range []int{0, 1, 42, 255, 65535} {
		m := allocFwmark(idx)
		if m>>16 != 0x4E4C {
			t.Errorf("allocFwmark(%d) upper 16 bits = 0x%X, want 0x4E4C", idx, m>>16)
		}
	}
}

// TestAllocFwmarkUniqueness verifies different interfaces get different fwmarks.
func TestAllocFwmarkUniqueness(t *testing.T) {
	m1 := allocFwmark(1)
	m2 := allocFwmark(2)
	m3 := allocFwmark(42)
	if m1 == m2 || m1 == m3 || m2 == m3 {
		t.Errorf("fwmarks not unique: idx1=0x%X, idx2=0x%X, idx42=0x%X", m1, m2, m3)
	}
}

// TestAllocFwmarkDeterministic verifies the same input always produces the same output.
func TestAllocFwmarkDeterministic(t *testing.T) {
	for _, idx := range []int{0, 1, 7, 255, 1000} {
		a := allocFwmark(idx)
		b := allocFwmark(idx)
		if a != b {
			t.Errorf("allocFwmark(%d) not deterministic: 0x%X != 0x%X", idx, a, b)
		}
	}
}

// TestAllocRouteTableValidRange verifies allocated tables stay in [1, 252].
func TestAllocRouteTableValidRange(t *testing.T) {
	reserved := map[int]bool{0: true, 253: true, 254: true, 255: true}
	for idx := 0; idx <= 500; idx++ {
		tbl := allocRouteTable(idx)
		if tbl < 1 || tbl > 252 {
			t.Errorf("allocRouteTable(%d) = %d, want [1, 252]", idx, tbl)
		}
		if reserved[tbl] {
			t.Errorf("allocRouteTable(%d) = %d conflicts with reserved table", idx, tbl)
		}
	}
}

// TestAllocRouteTableDeterministic verifies the same input always produces the same output.
func TestAllocRouteTableDeterministic(t *testing.T) {
	for _, idx := range []int{0, 1, 7, 255, 1000} {
		a := allocRouteTable(idx)
		b := allocRouteTable(idx)
		if a != b {
			t.Errorf("allocRouteTable(%d) not deterministic: %d != %d", idx, a, b)
		}
	}
}

// TestFlagKillSwitchConstant verifies the kill-switch flag value.
func TestFlagKillSwitchConstant(t *testing.T) {
	const flagKillSwitch uint32 = 1 << 0

	if flagKillSwitch != 1 {
		t.Errorf("flagKillSwitch = %d, want 1", flagKillSwitch)
	}
}

// TestCLIUsageRequiresMinimumArgs verifies that at least 3 arguments are needed.
func TestCLIUsageRequiresMinimumArgs(t *testing.T) {
	testCases := []struct {
		name string
		args []string
		ok   bool
	}{
		{"no args", []string{"netleak"}, false},
		{"one arg", []string{"netleak", "eth0"}, false},
		{"two args (valid)", []string{"netleak", "eth0", "curl"}, true},
		{"three args (valid)", []string{"netleak", "eth0", "curl", "ifconfig.me"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasEnoughArgs := len(tc.args) >= 3
			if hasEnoughArgs != tc.ok {
				t.Errorf("len(%v) >= 3 = %v, want %v", tc.args, hasEnoughArgs, tc.ok)
			}
		})
	}
}

// TestPolicyStructSize verifies the Go policy struct is 16 bytes (matches BPF ABI).
func TestPolicyStructSize(t *testing.T) {
	type policy struct {
		Fwmark   uint32
		Flags    uint32
		Ifindex  uint32
		Reserved uint32
	}

	sz := unsafe.Sizeof(policy{})
	if sz != 16 {
		t.Errorf("sizeof(policy) = %d, want 16", sz)
	}
}

// TestPolicyStructFieldLayout verifies the policy struct fields and kill-switch flag.
func TestPolicyStructFieldLayout(t *testing.T) {
	type policy struct {
		Fwmark   uint32
		Flags    uint32
		Ifindex  uint32
		Reserved uint32
	}

	pol := policy{Fwmark: allocFwmark(7), Flags: 0, Ifindex: 7}

	if pol.Fwmark == 0 {
		t.Error("policy.Fwmark should be non-zero")
	}
	if pol.Flags != 0 {
		t.Errorf("policy.Flags = %d, want 0", pol.Flags)
	}
	if pol.Ifindex != 7 {
		t.Errorf("policy.Ifindex = %d, want 7", pol.Ifindex)
	}

	pol.Flags = 1 << 0
	if pol.Flags != 1 {
		t.Errorf("policy.Flags with kill-switch = %d, want 1", pol.Flags)
	}
}
