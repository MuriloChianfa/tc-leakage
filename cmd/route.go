// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
)

// setupRouting creates:
//
//	ip rule  add fwmark <mark> table <table>
//	ip route add default dev <iface> table <table>
//
// EEXIST is tolerated so multiple netleak sessions can coexist.
func setupRouting(targetLink netlink.Link, mark uint32, table int) error {
	mask := uint32(0xFFFFFFFF)
	rule := netlink.NewRule()
	rule.Mark = mark
	rule.Mask = &mask
	rule.Table = table
	if err := netlink.RuleAdd(rule); err != nil && !errors.Is(err, syscall.EEXIST) {
		return fmt.Errorf("ip rule add: %w", err)
	}

	_, defaultDst, _ := net.ParseCIDR("0.0.0.0/0")
	route := &netlink.Route{
		LinkIndex: targetLink.Attrs().Index,
		Dst:       defaultDst,
		Table:     table,
		Scope:     netlink.SCOPE_LINK,
	}
	if err := netlink.RouteAdd(route); err != nil && !errors.Is(err, syscall.EEXIST) {
		return fmt.Errorf("ip route add: %w", err)
	}

	return nil
}

// cleanupRouting removes the policy routing rule and route.
// Errors are ignored because another session may have already cleaned up,
// or may still need these entries.
func cleanupRouting(targetLink netlink.Link, mark uint32, table int) {
	_, defaultDst, _ := net.ParseCIDR("0.0.0.0/0")
	route := &netlink.Route{
		LinkIndex: targetLink.Attrs().Index,
		Dst:       defaultDst,
		Table:     table,
	}
	_ = netlink.RouteDel(route)

	mask := uint32(0xFFFFFFFF)
	rule := netlink.NewRule()
	rule.Mark = mark
	rule.Mask = &mask
	rule.Table = table
	_ = netlink.RuleDel(rule)
}
