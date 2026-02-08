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

const (
	bpfObjPath = "bpf/netleak.o"
	bpfPinPath = "/sys/fs/bpf"
)

// policy matches struct policy in bpf/netleak.h.
type policy struct {
	Fwmark uint32
	Flags  uint32
}

const flagKillSwitch uint32 = 1 << 0

// bpfObjects maps to the ELF sections in bpf/netleak.o.
type bpfObjects struct {
	NetleakSockCreate *ebpf.Program `ebpf:"netleak_sock_create"`
	NetleakEgress     *ebpf.Program `ebpf:"netleak_egress"`
	CgroupPolicyMap   *ebpf.Map     `ebpf:"cgroup_policy_map"`
}

// Close releases all BPF resources.
func (o *bpfObjects) Close() {
	o.NetleakSockCreate.Close()
	o.NetleakEgress.Close()
	o.CgroupPolicyMap.Close()
}

// loadBPF reads the compiled BPF object from disk, loads it into the
// kernel, and returns the resulting programs and map. The map is pinned
// to bpffs so multiple netleak sessions share the same map.
func loadBPF() (*bpfObjects, error) {
	objData, err := os.ReadFile(bpfObjPath)
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

// attachBPF attaches both BPF programs to the given cgroup path:
//   - sock_create: sets sk->sk_mark on new sockets (steers routing)
//   - skb/egress:  kill-switch enforcement (drops packets when interface is down)
//
// The returned links must be closed to detach.
func attachBPF(objs *bpfObjects, cgPath string) (sockLink link.Link, egressLink link.Link, err error) {
	sockLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgPath,
		Attach:  ebpf.AttachCGroupInetSockCreate,
		Program: objs.NetleakSockCreate,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("attach sock_create: %w", err)
	}

	egressLink, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgPath,
		Attach:  ebpf.AttachCGroupInetEgress,
		Program: objs.NetleakEgress,
	})
	if err != nil {
		sockLink.Close()
		return nil, nil, fmt.Errorf("attach egress: %w", err)
	}

	return sockLink, egressLink, nil
}
