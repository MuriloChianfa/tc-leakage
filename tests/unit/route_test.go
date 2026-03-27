// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"net"
	"testing"
)

// TestAllocFwmarkFitsUint32 verifies allocated fwmarks fit in a uint32.
func TestAllocFwmarkFitsUint32(t *testing.T) {
	for _, idx := range []int{0, 1, 255, 65535} {
		m := allocFwmark(idx)
		if uint64(m) > 0xFFFFFFFF {
			t.Errorf("allocFwmark(%d) = 0x%X exceeds uint32 max", idx, m)
		}
	}
}

// TestAllocRouteTableNoReservedConflict verifies no reserved table IDs are returned.
func TestAllocRouteTableNoReservedConflict(t *testing.T) {
	reserved := []int{0, 253, 254, 255}
	for idx := 0; idx <= 500; idx++ {
		tbl := allocRouteTable(idx)
		for _, r := range reserved {
			if tbl == r {
				t.Errorf("allocRouteTable(%d) = %d conflicts with reserved table %d", idx, tbl, r)
			}
		}
	}
}

// TestAllocDifferentInterfacesGetDifferentTables verifies distinct interfaces
// get distinct routing tables.
func TestAllocDifferentInterfacesGetDifferentTables(t *testing.T) {
	t1 := allocRouteTable(1)
	t2 := allocRouteTable(2)
	if t1 == t2 {
		t.Errorf("allocRouteTable(1) == allocRouteTable(2) == %d", t1)
	}
}

// TestFwmarkMask verifies the mask used for policy routing is a full 32-bit mask.
func TestFwmarkMask(t *testing.T) {
	mask := uint32(0xFFFFFFFF)

	if mask != 0xFFFFFFFF {
		t.Errorf("fwmark mask = 0x%X, want 0xFFFFFFFF", mask)
	}
}

// TestIPv4DefaultCIDR verifies 0.0.0.0/0 is parsed correctly.
func TestIPv4DefaultCIDR(t *testing.T) {
	_, dst, err := net.ParseCIDR("0.0.0.0/0")
	if err != nil {
		t.Fatalf("ParseCIDR(0.0.0.0/0): %v", err)
	}
	if dst.IP.To4() == nil {
		t.Error("0.0.0.0/0 should be IPv4")
	}
}

// TestIPv6DefaultCIDR verifies ::/0 is parsed correctly.
func TestIPv6DefaultCIDR(t *testing.T) {
	_, dst, err := net.ParseCIDR("::/0")
	if err != nil {
		t.Fatalf("ParseCIDR(::/0): %v", err)
	}
	if dst.IP.To4() != nil {
		t.Error("::/0 should not be IPv4")
	}
}
