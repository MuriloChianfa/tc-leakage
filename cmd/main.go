package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/vishvananda/netlink"
)

const (
	fwmark     = 0x4E4C // "NL"
	routeTable = 100
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: netleak <interface> <command> [args...]\n\n")
		fmt.Fprintf(os.Stderr, "Routes all traffic from <command> (and children) through <interface>.\n")
		fmt.Fprintf(os.Stderr, "If the interface goes down, traffic is dropped (kill-switch).\n\n")
		fmt.Fprintf(os.Stderr, "Example:\n")
		fmt.Fprintf(os.Stderr, "  sudo netleak ppp0 curl ifconfig.me\n")
		os.Exit(1)
	}

	targetIface := os.Args[1]
	cmdArgs := os.Args[2:]

	code, err := run(targetIface, cmdArgs)
	if err != nil {
		log.Fatalf("netleak: %v", err)
	}
	os.Exit(code)
}

func run(targetIface string, cmdArgs []string) (int, error) {
	// --- Verify target interface exists ---
	targetLink, err := netlink.LinkByName(targetIface)
	if err != nil {
		return 1, fmt.Errorf("interface %q: %w", targetIface, err)
	}
	log.Printf("Target interface: %s (index %d)", targetIface, targetLink.Attrs().Index)

	// --- Create cgroup v2 ---
	sessionID := strconv.Itoa(os.Getpid())
	cgPath, err := createCgroup(sessionID)
	if err != nil {
		return 1, err
	}
	defer cleanupCgroup(cgPath)
	log.Printf("Cgroup: %s", cgPath)

	// --- Load BPF programs ---
	objs, err := loadBPF()
	if err != nil {
		return 1, err
	}
	defer objs.Close()

	// --- Attach BPF programs to cgroup ---
	sockLink, egressLink, err := attachBPF(objs, cgPath)
	if err != nil {
		return 1, err
	}
	defer sockLink.Close()
	defer egressLink.Close()

	// --- Get cgroup ID ---
	cgID, err := getCgroupID(cgPath)
	if err != nil {
		return 1, fmt.Errorf("cgroup ID: %w", err)
	}
	log.Printf("Cgroup ID: %d", cgID)

	// --- Setup policy routing ---
	if err := setupRouting(targetLink, fwmark, routeTable); err != nil {
		return 1, fmt.Errorf("routing: %w", err)
	}
	defer cleanupRouting(targetLink, fwmark, routeTable)
	log.Printf("Policy routing: fwmark 0x%X -> table %d -> dev %s", fwmark, routeTable, targetIface)

	// --- Populate BPF map: cgroup_id -> policy ---
	pol := policy{Fwmark: uint32(fwmark), Flags: 0}
	if err := objs.CgroupPolicyMap.Put(cgID, pol); err != nil {
		return 1, fmt.Errorf("populate map: %w", err)
	}
	defer func() {
		if err := objs.CgroupPolicyMap.Delete(cgID); err != nil {
			log.Printf("Warning: map cleanup: %v", err)
		}
	}()

	// --- Move self into cgroup ---
	if err := joinCgroup(cgPath, os.Getpid()); err != nil {
		return 1, fmt.Errorf("join cgroup: %w", err)
	}

	// --- Start interface monitor ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go monitorInterface(ctx, targetIface, objs.CgroupPolicyMap, cgID)

	// --- Run target command ---
	log.Printf("Executing: %v", cmdArgs)
	return execAndWait(ctx, cancel, cmdArgs)
}

// execAndWait forks the target command, forwards signals, waits for
// it to exit, and returns its exit code.
func execAndWait(_ context.Context, cancel context.CancelFunc, cmdArgs []string) (int, error) {
	child := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	if err := child.Start(); err != nil {
		return 1, fmt.Errorf("exec %s: %w", cmdArgs[0], err)
	}

	// Forward signals to the child process
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			_ = child.Process.Signal(sig)
		}
	}()

	// Wait for child to exit
	waitErr := child.Wait()
	signal.Stop(sigCh)
	close(sigCh)
	cancel()

	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("wait: %w", waitErr)
	}
	return 0, nil
}
