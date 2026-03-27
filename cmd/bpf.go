// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

const bpfPinPath = "/sys/fs/bpf"

// bpfObjSearchPaths lists where to look for the compiled BPF object,
// in priority order. The first path is the installed system location
// used by .deb/.rpm packages; the second is the relative path used
// during development.
var bpfObjSearchPaths = []string{
	"/usr/lib/netleak/netleak.o",
	"bpf/netleak.o",
}

// policy matches struct policy in bpf/netleak.h (16 bytes, stable ABI).
type policy struct {
	Fwmark   uint32
	Flags    uint32
	Ifindex  uint32
	Reserved uint32
}

const flagKillSwitch uint32 = 1 << 0

// bpfObjects maps to the ELF sections in bpf/netleak.o.
type bpfObjects struct {
	NetleakSockCreate *ebpf.Program `ebpf:"netleak_sock_create"`
	NetleakEgress     *ebpf.Program `ebpf:"netleak_egress"`
	NetleakIngress    *ebpf.Program `ebpf:"netleak_ingress"`
	CgroupPolicyMap   *ebpf.Map     `ebpf:"cgroup_policy_map"`
}

// Close releases all BPF resources.
func (o *bpfObjects) Close() {
	o.NetleakSockCreate.Close()
	o.NetleakEgress.Close()
	o.NetleakIngress.Close()
	o.CgroupPolicyMap.Close()
}

// findBPFObj returns the first path from bpfObjSearchPaths that exists
// on disk. It checks the installed system path before the development path.
func findBPFObj() (string, error) {
	for _, p := range bpfObjSearchPaths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("BPF object not found in any of %v", bpfObjSearchPaths)
}

// loadBPF reads the compiled BPF object from disk, loads it into the
// kernel, and returns the resulting programs and map. The map is pinned
// to bpffs so multiple netleak sessions share the same map.
func loadBPF() (*bpfObjects, error) {
	objPath, err := findBPFObj()
	if err != nil {
		return nil, err
	}

	objData, err := os.ReadFile(objPath)
	if err != nil {
		return nil, fmt.Errorf("read BPF object: %w", err)
	}

	spec, err := ebpf.LoadCollectionSpecFromReader(bytes.NewReader(objData))
	if err != nil {
		return nil, fmt.Errorf("parse BPF object: %w", err)
	}

	var objs bpfObjects
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{PinPath: bpfPinPath},
	}); err != nil {
		return nil, fmt.Errorf("load BPF: %w", err)
	}

	return &objs, nil
}

// attachBPF attaches BPF programs to the given cgroup path:
//   - sock_create: sets sk->sk_mark on new sockets (steers routing)
//   - skb/egress:  kill-switch enforcement (drops packets when interface is down)
//   - skb/ingress: ingress filtering (only when ingressFilter is true)
//
// The returned links must be closed to detach.
func attachBPF(objs *bpfObjects, cgPath string, ingressFilter bool) (sockLink link.Link, egressLink link.Link, ingressLink link.Link, err error) {
	sockLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgPath,
		Attach:  ebpf.AttachCGroupInetSockCreate,
		Program: objs.NetleakSockCreate,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("attach sock_create: %w", err)
	}

	egressLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgPath,
		Attach:  ebpf.AttachCGroupInetEgress,
		Program: objs.NetleakEgress,
	})
	if err != nil {
		sockLink.Close()
		return nil, nil, nil, fmt.Errorf("attach egress: %w", err)
	}

	if ingressFilter {
		ingressLink, err = link.AttachCgroup(link.CgroupOptions{
			Path:    cgPath,
			Attach:  ebpf.AttachCGroupInetIngress,
			Program: objs.NetleakIngress,
		})
		if err != nil {
			sockLink.Close()
			egressLink.Close()
			return nil, nil, nil, fmt.Errorf("attach ingress: %w", err)
		}
	}

	return sockLink, egressLink, ingressLink, nil
}
