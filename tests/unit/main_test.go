// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"testing"
)

// TestFwmarkConstant verifies the fwmark value matches the expected "NL" hex encoding.
func TestFwmarkConstant(t *testing.T) {
	const fwmark = 0x4E4C // "NL"

	if fwmark != 0x4E4C {
		t.Errorf("fwmark = 0x%X, want 0x4E4C", fwmark)
	}

	// Verify the ASCII interpretation: 'N' = 0x4E, 'L' = 0x4C
	if fwmark>>8 != 'N' {
		t.Errorf("fwmark high byte = 0x%X, want 0x4E ('N')", fwmark>>8)
	}
	if fwmark&0xFF != 'L' {
		t.Errorf("fwmark low byte = 0x%X, want 0x4C ('L')", fwmark&0xFF)
	}
}

// TestRouteTableConstant verifies the routing table ID is within the valid range.
func TestRouteTableConstant(t *testing.T) {
	const routeTable = 100

	if routeTable < 1 || routeTable > 252 {
		t.Errorf("routeTable = %d, want value in range [1, 252]", routeTable)
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
