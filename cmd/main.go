package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// bpfObjects is a struct that matches the sections in tc_leakage.o:
//   - Program: tc_leakage
//   - Map:     PID_IF_MAP
type bpfObjects struct {
	XdpPidRedirect  *ebpf.Program `ebpf:"tc_leakage"`
	PIDInterfaceMap *ebpf.Map     `ebpf:"PID_IF_MAP"`
}

const mapPinPath = "/sys/fs/bpf/tc/globals/PID_IF_MAP"
const progPinPath = "/sys/fs/bpf/tc_leakage"
const bpfObjPath = "bpf/leakage.o"

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  tc-leakage load --iface <iface>
      Loads & attaches the eBPF XDP program to <iface> and pins the map.

  tc-leakage set --pid <pid> --redir <iface>
      Updates the pinned map with (PID -> ifindex) so packets from <pid> redirect to <iface>.

  tc-leakage show
      Dumps the contents of the pinned map.

  tc-leakage help
      Shows this help.

Examples:
  sudo ./tc-leakage load --iface enp2s0
  sudo ./tc-leakage set --pid 29799 --redir ppp0
  sudo ./tc-leakage show
`)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	cmd := os.Args[1]
	switch cmd {
	case "load":
		loadFlags := flag.NewFlagSet("load", flag.ExitOnError)
		iface := loadFlags.String("iface", "", "Network interface (e.g. enp2s0) to attach XDP")
		_ = loadFlags.Parse(os.Args[2:])
		if *iface == "" {
			loadFlags.Usage()
			os.Exit(1)
		}

		if err := loadAndAttach(*iface); err != nil {
			log.Fatalf("Load error: %v", err)
		}
		fmt.Printf("Loaded & attached XDP to %s successfully\n", *iface)

	case "set":
		setFlags := flag.NewFlagSet("set", flag.ExitOnError)
		pid := setFlags.Uint("pid", 0, "PID to match")
		redir := setFlags.String("redir", "", "Interface name to redirect packets (e.g. eth1)")
		_ = setFlags.Parse(os.Args[2:])
		if *pid == 0 || *redir == "" {
			setFlags.Usage()
			os.Exit(1)
		}

		if err := setPidRedirect(uint32(*pid), *redir); err != nil {
			log.Fatalf("Set error: %v", err)
		}

	case "show":
		if err := showMap(); err != nil {
			log.Fatalf("Show error: %v", err)
		}

	case "help":
		usage()

	default:
		usage()
	}
}

func loadAndAttach(iface string) error {
	objData, err := os.ReadFile(bpfObjPath)
	if err != nil {
		return fmt.Errorf("reading bpf object file: %w", err)
	}

	rdr := bytes.NewReader(objData)

	spec, err := ebpf.LoadCollectionSpecFromReader(rdr)
	if err != nil {
		return fmt.Errorf("load collection spec: %w", err)
	}

	var objs bpfObjects
	if err := spec.LoadAndAssign(&objs, &ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: "/sys/fs/bpf",
		},
	}); err != nil {
		return fmt.Errorf("load and assign: %w", err)
	}

	ifIndex, err := getIfIndex(iface)
	if err != nil {
		return fmt.Errorf("getting ifindex: %w", err)
	}
	l, err := link.AttachXDP(link.XDPOptions{
		Program:   objs.XdpPidRedirect,
		Interface: ifIndex,
		Flags:     link.XDPGenericMode, // link.XDPDriverMode
	})
	if err != nil {
		return fmt.Errorf("attach XDP: %w", err)
	}

	err = objs.XdpPidRedirect.Pin(progPinPath)
	if err != nil {
		return fmt.Errorf("pin program: %w", err)
	}

	err = objs.PIDInterfaceMap.Pin(mapPinPath)
	if err != nil {
		// If you see "File exists," ensure you remove or unpin it first
		return fmt.Errorf("pin map: %w", err)
	}

	_ = l

	return nil
}

func setPidRedirect(pid uint32, redirIface string) error {
	m, err := ebpf.LoadPinnedMap(mapPinPath, nil)
	if err != nil {
		return fmt.Errorf("open pinned map: %w", err)
	}
	defer m.Close()

	ifIndex, err := getIfIndex(redirIface)
	if err != nil {
		return fmt.Errorf("getting ifindex: %w", err)
	}

	err = m.Put(pid, uint32(ifIndex))
	if err != nil {
		return fmt.Errorf("map put: %w", err)
	}

	fmt.Printf("Set PID %d => ifindex %d\n", pid, ifIndex)
	return nil
}

func showMap() error {
	m, err := ebpf.LoadPinnedMap(mapPinPath, nil)
	if err != nil {
		return fmt.Errorf("open pinned map: %w", err)
	}
	defer m.Close()

	var (
		key   uint32
		value uint32
	)

	iter := m.Iterate()
	for iter.Next(&key, &value) {
		fmt.Printf("PID %d => ifindex %d\n", key, value)
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("iterate map: %w", err)
	}
	return nil
}

func getIfIndex(ifaceName string) (int, error) {
	path := fmt.Sprintf("/sys/class/net/%s/ifindex", ifaceName)
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read ifindex: %w", err)
	}
	ifIndexStr := strings.TrimSpace(string(data))
	idx, err := strconv.Atoi(ifIndexStr)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ifindex: %w", err)
	}
	return idx, nil
}
