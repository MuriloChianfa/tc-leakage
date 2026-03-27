// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// hasGlobalIPv6 returns true if the interface has at least one non-link-local
// IPv6 address (i.e. a global or unique-local address that can actually route).
func hasGlobalIPv6(link netlink.Link) bool {
	addrs, err := netlink.AddrList(link, unix.AF_INET6)
	if err != nil {
		return false
	}
	for _, a := range addrs {
		if a.IP.IsGlobalUnicast() && !a.IP.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

// setupRouting creates policy routing rules and routes. The behaviour
// depends on the ipMode:
//
//   - auto:     dual-stack; blackhole IPv6 when the interface lacks a global v6 address
//   - v4only:   route only IPv4; always blackhole IPv6
//   - v6only:   route only IPv6; always blackhole IPv4
//   - fallback: dual-stack; if the interface lacks global v6, skip the v6
//     rule entirely (IPv6 falls back to the system default route)
//
// EEXIST is tolerated so multiple netleak sessions can coexist.
func setupRouting(targetLink netlink.Link, mark uint32, table int, mode ipMode) error {
	mask := uint32(0xFFFFFFFF)
	idx := targetLink.Attrs().Index
	ipv6ok := hasGlobalIPv6(targetLink)
	name := targetLink.Attrs().Name

	switch mode {
	case ipModeV4Only:
		log.Printf("Mode: only-v4 — routing IPv4 only, blocking IPv6")
		if err := addRuleAndRoute(unix.AF_INET, "0.0.0.0/0", mark, &mask, table, idx, false); err != nil {
			return err
		}
		return addRuleAndRoute(unix.AF_INET6, "::/0", mark, &mask, table, idx, true)

	case ipModeV6Only:
		log.Printf("Mode: only-v6 — routing IPv6 only, blocking IPv4")
		if err := addRuleAndRoute(unix.AF_INET6, "::/0", mark, &mask, table, idx, false); err != nil {
			return err
		}
		return addRuleAndRoute(unix.AF_INET, "0.0.0.0/0", mark, &mask, table, idx, true)

	case ipModeFallback:
		if err := addRuleAndRoute(unix.AF_INET, "0.0.0.0/0", mark, &mask, table, idx, false); err != nil {
			return err
		}
		if ipv6ok {
			return addRuleAndRoute(unix.AF_INET6, "::/0", mark, &mask, table, idx, false)
		}
		log.Printf("Warning: %s has no global IPv6 address — IPv6 will use default route (--fallback-to-v4)", name)
		return nil

	default: // ipModeAuto
		if err := addRuleAndRoute(unix.AF_INET, "0.0.0.0/0", mark, &mask, table, idx, false); err != nil {
			return err
		}
		if ipv6ok {
			return addRuleAndRoute(unix.AF_INET6, "::/0", mark, &mask, table, idx, false)
		}
		log.Printf("Warning: %s has no global IPv6 address — IPv6 traffic will be blocked", name)
		return addRuleAndRoute(unix.AF_INET6, "::/0", mark, &mask, table, idx, true)
	}
}

func addRuleAndRoute(family int, cidr string, mark uint32, mask *uint32, table int, ifindex int, blackhole bool) error {
	label := "IPv4"
	if family == unix.AF_INET6 {
		label = "IPv6"
	}

	rule := netlink.NewRule()
	rule.Family = family
	rule.Mark = mark
	rule.Mask = mask
	rule.Table = table
	if err := netlink.RuleAdd(rule); err != nil && !errors.Is(err, syscall.EEXIST) {
		return fmt.Errorf("ip rule add (%s): %w", label, err)
	}

	_, dst, _ := net.ParseCIDR(cidr)

	if blackhole {
		route := &netlink.Route{
			Dst:   dst,
			Table: table,
			Type:  unix.RTN_BLACKHOLE,
		}
		if err := netlink.RouteAdd(route); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("ip route add blackhole (%s): %w", label, err)
		}
	} else {
		route := &netlink.Route{
			LinkIndex: ifindex,
			Dst:       dst,
			Table:     table,
			Scope:     netlink.SCOPE_LINK,
		}
		if err := netlink.RouteAdd(route); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("ip route add (%s): %w", label, err)
		}
	}

	return nil
}

// cleanupRouting removes policy routing rules and routes for both families.
// Tries to delete both normal and blackhole routes (one will succeed,
// the other will silently fail). Errors are ignored because another
// session may have already cleaned up.
func cleanupRouting(targetLink netlink.Link, mark uint32, table int) {
	mask := uint32(0xFFFFFFFF)
	idx := targetLink.Attrs().Index

	families := []struct {
		af   int
		cidr string
	}{
		{unix.AF_INET, "0.0.0.0/0"},
		{unix.AF_INET6, "::/0"},
	}

	for _, f := range families {
		_, dst, _ := net.ParseCIDR(f.cidr)

		_ = netlink.RouteDel(&netlink.Route{
			LinkIndex: idx,
			Dst:       dst,
			Table:     table,
		})
		_ = netlink.RouteDel(&netlink.Route{
			Dst:   dst,
			Table: table,
			Type:  unix.RTN_BLACKHOLE,
		})

		rule := netlink.NewRule()
		rule.Family = f.af
		rule.Mark = mark
		rule.Mask = &mask
		rule.Table = table
		_ = netlink.RuleDel(rule)
	}
}
