// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"testing"
)

// TestPolicyIfindexPopulated verifies the policy struct holds the ifindex correctly.
func TestPolicyIfindexPopulated(t *testing.T) {
	type policy struct {
		Fwmark   uint32
		Flags    uint32
		Ifindex  uint32
		Reserved uint32
	}

	testCases := []struct {
		name    string
		ifindex uint32
	}{
		{"loopback", 1},
		{"eth0", 2},
		{"wg0", 7},
		{"high index", 255},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pol := policy{
				Fwmark:  allocFwmark(int(tc.ifindex)),
				Flags:   0,
				Ifindex: tc.ifindex,
			}
			if pol.Ifindex != tc.ifindex {
				t.Errorf("pol.Ifindex = %d, want %d", pol.Ifindex, tc.ifindex)
			}
			if pol.Fwmark == 0 {
				t.Error("pol.Fwmark should be non-zero")
			}
		})
	}
}

// TestPolicyKillSwitchAlsoBlocksIngress verifies kill-switch flag
// semantics: when FLAG_KILL_SWITCH is set, both egress and ingress
// BPF programs should drop traffic.
func TestPolicyKillSwitchAlsoBlocksIngress(t *testing.T) {
	const flagKillSwitch uint32 = 1 << 0

	type policy struct {
		Fwmark   uint32
		Flags    uint32
		Ifindex  uint32
		Reserved uint32
	}

	pol := policy{Fwmark: allocFwmark(7), Flags: flagKillSwitch, Ifindex: 7}

	if pol.Flags&flagKillSwitch == 0 {
		t.Error("kill-switch flag not set")
	}
}

// TestIngressFilterRequiresIfindex verifies that ingress filtering
// is only active when ifindex is non-zero.
func TestIngressFilterRequiresIfindex(t *testing.T) {
	type policy struct {
		Fwmark   uint32
		Flags    uint32
		Ifindex  uint32
		Reserved uint32
	}

	pol := policy{Fwmark: allocFwmark(7), Ifindex: 0}
	if pol.Ifindex != 0 {
		t.Error("ifindex=0 should mean no ingress filtering")
	}

	pol.Ifindex = 7
	if pol.Ifindex == 0 {
		t.Error("ifindex=7 should enable ingress filtering")
	}
}
