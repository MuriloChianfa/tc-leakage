package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	cgroupBase = "/sys/fs/cgroup"
	cgroupNS   = "netleak"
)

// createCgroup creates a new cgroup v2 directory for this session
// under /sys/fs/cgroup/netleak/<sessionID> and returns its path.
func createCgroup(sessionID string) (string, error) {
	cgPath := filepath.Join(cgroupBase, cgroupNS, sessionID)
	if err := os.MkdirAll(cgPath, 0755); err != nil {
		return "", fmt.Errorf("create cgroup %s: %w", cgPath, err)
	}
	return cgPath, nil
}

// getCgroupID returns the cgroup v2 ID, which is the inode number
// of the cgroup directory on cgroupfs.
func getCgroupID(cgPath string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(cgPath, &st); err != nil {
		return 0, err
	}
	return st.Ino, nil
}

// joinCgroup moves the given PID into the cgroup.
func joinCgroup(cgPath string, pid int) error {
	p := filepath.Join(cgPath, "cgroup.procs")
	return os.WriteFile(p, []byte(strconv.Itoa(pid)), 0644)
}

// cleanupCgroup moves the calling process back to the parent cgroup
// and removes the session cgroup directory.
func cleanupCgroup(cgPath string) {
	parent := filepath.Dir(cgPath)
	_ = joinCgroup(parent, os.Getpid())
	if err := os.Remove(cgPath); err != nil {
		log.Printf("Warning: remove cgroup: %v", err)
	}
}
