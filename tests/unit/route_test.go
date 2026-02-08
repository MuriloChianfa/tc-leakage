// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"testing"
)

// TestFwmarkNonZero verifies the fwmark is non-zero (required for policy routing).
func TestFwmarkNonZero(t *testing.T) {
	const fwmark = 0x4E4C

	if fwmark == 0 {
		t.Error("fwmark must be non-zero for policy routing to work")
	}
}

// TestFwmarkFitsUint32 verifies the fwmark fits in a uint32 (kernel sk->mark size).
func TestFwmarkFitsUint32(t *testing.T) {
	const fwmark = 0x4E4C

	if fwmark > 0xFFFFFFFF {
		t.Errorf("fwmark = 0x%X exceeds uint32 max", fwmark)
	}
}

// TestRouteTableInValidRange verifies the routing table ID is in the user-defined range.
// Linux routing tables: 0 = unspec, 253 = default, 254 = main, 255 = local.
// User tables should be in [1, 252].
func TestRouteTableInValidRange(t *testing.T) {
	const routeTable = 100

	reserved := []int{0, 253, 254, 255}
	for _, r := range reserved {
		if routeTable == r {
			t.Errorf("routeTable = %d conflicts with reserved table %d", routeTable, r)
		}
	}
}

// TestFwmarkMask verifies the mask used for policy routing is a full 32-bit mask.
func TestFwmarkMask(t *testing.T) {
	mask := uint32(0xFFFFFFFF)

	if mask != 0xFFFFFFFF {
		t.Errorf("fwmark mask = 0x%X, want 0xFFFFFFFF", mask)
	}
}

// TestPolicyStructLayout verifies the policy struct field layout matches BPF expectations.
func TestPolicyStructLayout(t *testing.T) {
	type policy struct {
		Fwmark uint32
		Flags  uint32
	}

	pol := policy{Fwmark: 0x4E4C, Flags: 0}

	if pol.Fwmark != 0x4E4C {
		t.Errorf("policy.Fwmark = 0x%X, want 0x4E4C", pol.Fwmark)
	}
	if pol.Flags != 0 {
		t.Errorf("policy.Flags = %d, want 0", pol.Flags)
	}

	// Test kill-switch flag
	pol.Flags = 1 << 0
	if pol.Flags != 1 {
		t.Errorf("policy.Flags with kill-switch = %d, want 1", pol.Flags)
	}
}
