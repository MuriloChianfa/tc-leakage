// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"context"
	"log"

	"github.com/cilium/ebpf"
	"github.com/vishvananda/netlink"
)

// monitorInterface watches the target interface via netlink and toggles
// the kill-switch flag in the BPF map when the interface goes up or down.
// It blocks until ctx is cancelled.
func monitorInterface(ctx context.Context, ifaceName string, m *ebpf.Map, cgID uint64) {
	updates := make(chan netlink.LinkUpdate)
	done := make(chan struct{})
	if err := netlink.LinkSubscribe(updates, done); err != nil {
		log.Printf("Warning: interface monitor unavailable: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			close(done)
			return
		case update := <-updates:
			if update.Attrs().Name != ifaceName {
				continue
			}

			var flags uint32
			if update.Attrs().OperState != netlink.OperUp {
				flags = flagKillSwitch
				log.Printf("Interface %s went down — kill-switch ON", ifaceName)
			} else {
				log.Printf("Interface %s came up — kill-switch OFF", ifaceName)
			}

			pol := policy{Fwmark: uint32(fwmark), Flags: flags}
			if err := m.Put(cgID, pol); err != nil {
				log.Printf("Warning: kill-switch update failed: %v", err)
			}
		}
	}
}
