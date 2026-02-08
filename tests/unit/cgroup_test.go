// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"path/filepath"
	"strings"
	"testing"
)

const (
	cgroupBase = "/sys/fs/cgroup"
	cgroupNS   = "netleak"
)

// TestCgroupPathGeneration verifies cgroup paths are generated correctly.
func TestCgroupPathGeneration(t *testing.T) {
	testCases := []struct {
		sessionID string
		expected  string
	}{
		{"12345", "/sys/fs/cgroup/netleak/12345"},
		{"1", "/sys/fs/cgroup/netleak/1"},
		{"99999", "/sys/fs/cgroup/netleak/99999"},
	}

	for _, tc := range testCases {
		t.Run(tc.sessionID, func(t *testing.T) {
			cgPath := filepath.Join(cgroupBase, cgroupNS, tc.sessionID)
			if cgPath != tc.expected {
				t.Errorf("cgPath = %q, want %q", cgPath, tc.expected)
			}
		})
	}
}

// TestCgroupPathContainsNamespace verifies all cgroup paths include the netleak namespace.
func TestCgroupPathContainsNamespace(t *testing.T) {
	sessionID := "42"
	cgPath := filepath.Join(cgroupBase, cgroupNS, sessionID)

	if !strings.Contains(cgPath, "/netleak/") {
		t.Errorf("cgroup path %q does not contain /netleak/ namespace", cgPath)
	}
}

// TestCgroupPathIsAbsolute verifies cgroup paths are always absolute.
func TestCgroupPathIsAbsolute(t *testing.T) {
	sessionID := "1234"
	cgPath := filepath.Join(cgroupBase, cgroupNS, sessionID)

	if !filepath.IsAbs(cgPath) {
		t.Errorf("cgroup path %q is not absolute", cgPath)
	}
}

// TestCgroupParentPath verifies parent cgroup path computation for cleanup.
func TestCgroupParentPath(t *testing.T) {
	cgPath := filepath.Join(cgroupBase, cgroupNS, "12345")
	parent := filepath.Dir(cgPath)
	expected := filepath.Join(cgroupBase, cgroupNS)

	if parent != expected {
		t.Errorf("parent = %q, want %q", parent, expected)
	}
}
