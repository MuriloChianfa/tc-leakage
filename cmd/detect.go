// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// detectByGateway finds the network interface whose routing table contains
// a route with the given gateway IP. Returns the interface name.
func detectByGateway(gateway string) (string, error) {
	gw := net.ParseIP(gateway)
	if gw == nil {
		return "", fmt.Errorf("invalid gateway IP: %q", gateway)
	}

	family := unix.AF_INET
	if gw.To4() == nil {
		family = unix.AF_INET6
	}

	routes, err := netlink.RouteList(nil, family)
	if err != nil {
		return "", fmt.Errorf("list routes: %w", err)
	}

	for _, r := range routes {
		if r.Gw != nil && r.Gw.Equal(gw) {
			link, err := netlink.LinkByIndex(r.LinkIndex)
			if err != nil {
				continue
			}
			return link.Attrs().Name, nil
		}
	}

	return "", fmt.Errorf("no route with gateway %s", gateway)
}

// detectBySubnet finds the network interface that has an address within
// the given CIDR range. Returns the interface name.
func detectBySubnet(cidr string) (string, error) {
	_, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}

	links, err := netlink.LinkList()
	if err != nil {
		return "", fmt.Errorf("list links: %w", err)
	}

	family := unix.AF_INET
	if subnet.IP.To4() == nil {
		family = unix.AF_INET6
	}

	for _, link := range links {
		addrs, err := netlink.AddrList(link, family)
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if subnet.Contains(addr.IP) {
				return link.Attrs().Name, nil
			}
		}
	}

	return "", fmt.Errorf("no interface with address in %s", cidr)
}
