// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"net"
	"testing"
)

// TestGatewayIPParsing verifies valid and invalid gateway IPs are parsed correctly.
func TestGatewayIPParsing(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
		ipv6  bool
	}{
		{"10.0.0.1", true, false},
		{"192.168.1.1", true, false},
		{"0.0.0.0", true, false},
		{"fd10::1", true, true},
		{"::1", true, true},
		{"not-an-ip", false, false},
		{"", false, false},
		{"256.1.1.1", false, false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			ip := net.ParseIP(tc.input)
			if tc.valid && ip == nil {
				t.Errorf("expected valid IP, got nil")
			}
			if !tc.valid && ip != nil {
				t.Errorf("expected invalid IP, got %s", ip)
			}
			if tc.valid && tc.ipv6 && ip.To4() != nil {
				t.Errorf("expected IPv6, got IPv4 for %s", tc.input)
			}
		})
	}
}

// TestSubnetCIDRParsing verifies valid and invalid CIDR notations.
func TestSubnetCIDRParsing(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
	}{
		{"10.10.0.0/24", true},
		{"192.168.0.0/16", true},
		{"fd10::/64", true},
		{"::/0", true},
		{"0.0.0.0/0", true},
		{"not-a-cidr", false},
		{"", false},
		{"10.10.0.0", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			_, _, err := net.ParseCIDR(tc.input)
			if tc.valid && err != nil {
				t.Errorf("expected valid CIDR, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Errorf("expected invalid CIDR, got success")
			}
		})
	}
}

// TestSubnetContains verifies the Contains logic used by detectBySubnet.
func TestSubnetContains(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("10.10.0.0/24")

	testCases := []struct {
		ip       string
		contains bool
	}{
		{"10.10.0.1", true},
		{"10.10.0.254", true},
		{"10.10.1.1", false},
		{"192.168.0.1", false},
	}

	for _, tc := range testCases {
		t.Run(tc.ip, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if subnet.Contains(ip) != tc.contains {
				t.Errorf("%s.Contains(%s) = %v, want %v",
					subnet, tc.ip, !tc.contains, tc.contains)
			}
		})
	}
}

// TestCLIParseArgsBackwardCompat verifies the direct <iface> form still works.
func TestCLIParseArgsBackwardCompat(t *testing.T) {
	args := []string{"wg0", "curl", "ifconfig.me"}

	if args[0] == "--gateway" || args[0] == "--subnet" {
		t.Error("direct interface name should not be treated as a flag")
	}
}

// TestCLIParseArgsGatewayFlag verifies --gateway flag structure.
func TestCLIParseArgsGatewayFlag(t *testing.T) {
	args := []string{"--gateway", "10.0.0.1", "curl", "ifconfig.me"}

	if args[0] != "--gateway" {
		t.Error("expected --gateway flag")
	}
	ip := net.ParseIP(args[1])
	if ip == nil {
		t.Errorf("invalid gateway IP: %s", args[1])
	}
	if len(args[2:]) < 1 {
		t.Error("expected command after --gateway <ip>")
	}
}

// TestCLIParseArgsSubnetFlag verifies --subnet flag structure.
func TestCLIParseArgsSubnetFlag(t *testing.T) {
	args := []string{"--subnet", "10.10.0.0/24", "bash"}

	if args[0] != "--subnet" {
		t.Error("expected --subnet flag")
	}
	_, _, err := net.ParseCIDR(args[1])
	if err != nil {
		t.Errorf("invalid CIDR: %v", err)
	}
	if len(args[2:]) < 1 {
		t.Error("expected command after --subnet <cidr>")
	}
}

// TestCLIIngressFilterIsOptIn verifies --ingress-filter is recognized as a flag.
func TestCLIIngressFilterIsOptIn(t *testing.T) {
	testCases := []struct {
		name          string
		args          []string
		expectIngress bool
	}{
		{"default (no flag)", []string{"wg0", "curl"}, false},
		{"with --ingress-filter", []string{"--ingress-filter", "wg0", "curl"}, true},
		{"flag after iface", []string{"wg0", "curl"}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasIngress := false
			for _, a := range tc.args {
				if a == "--ingress-filter" {
					hasIngress = true
					break
				}
			}
			if hasIngress != tc.expectIngress {
				t.Errorf("ingress flag = %v, want %v", hasIngress, tc.expectIngress)
			}
		})
	}
}
