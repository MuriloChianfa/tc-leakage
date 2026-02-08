---
name: Bug Report
about: Report a bug or unexpected behavior in netleak
title: "[BUG] "
labels: bug
assignees: ''

---

## Bug Description

<!-- A clear and concise description of what the bug is -->

## Environment

**Operating System:**
<!-- e.g., Ubuntu 24.04, Debian 12, Fedora 40 -->

**Kernel Version:**
<!-- e.g., 6.8.0-generic (run `uname -r`) -->

**Go Version:**
<!-- e.g., 1.24.0 (run `go version`) -->

**clang/LLVM Version:**
<!-- e.g., 18.0.0 (run `clang --version`) -->

**netleak Version:**
<!-- e.g., 1.0.0, or commit hash if building from source -->

**cgroup v2 Status:**
<!-- Run: mount | grep cgroup2 -->

**BPF Filesystem:**
<!-- Run: mount | grep bpf -->

## Steps to Reproduce

<!-- Provide a minimal, reproducible example -->

1. Build with: <!-- e.g., make all -->
2. Run with: <!-- e.g., sudo netleak wg0 curl ifconfig.me -->
3. Observe: <!-- what happens -->

### Commands Used

```bash
# Commands used to reproduce the issue
```

## Expected Behavior

<!-- What you expected to happen -->

## Actual Behavior

<!-- What actually happened -->

## Error Messages

<!-- If applicable, paste any error messages, kernel logs, or debug output -->

```
Paste error messages here
```

### Kernel Logs

<!-- If applicable, run: dmesg | tail -50 -->

```
Paste dmesg output here
```

### BPF Status

<!-- If applicable, run: sudo bpftool prog show && sudo bpftool map list -->

```
Paste bpftool output here
```

## Additional Context

### Severity

- [ ] Critical - Traffic leaks to wrong interface / kill-switch fails
- [ ] High - Crash, panic, or data corruption
- [ ] Medium - Incorrect behavior in specific scenarios
- [ ] Low - Minor inconvenience

### Frequency

- [ ] Always reproducible
- [ ] Intermittent
- [ ] Rare

### Affected Components

<!-- Check all that apply -->

- [ ] BPF program loading
- [ ] BPF program attachment
- [ ] Cgroup creation/management
- [ ] Policy routing setup
- [ ] Interface monitoring
- [ ] Kill-switch behavior
- [ ] Signal handling
- [ ] Command execution
- [ ] Cleanup/teardown
- [ ] Build system
- [ ] Documentation
- [ ] Other: <!-- specify -->

## Workarounds

<!-- If you found any workarounds, please describe them -->

## Possible Fix

<!-- If you have suggestions on how to fix the bug -->

## Related Issues

<!-- Link any related issues or PRs -->

## Testing

<!-- Have you tested with different configurations? -->

- [ ] Tested with different kernel version
- [ ] Tested with different network interface type (wg, tun, ppp, etc.)
- [ ] Tested on different distribution
- [ ] Verified with latest version from main branch

---

**Note:** For security vulnerabilities, please **DO NOT** open a public issue. Instead, email murilo.chianfa@outlook.com. See [SECURITY.md](../SECURITY.md) for details.
