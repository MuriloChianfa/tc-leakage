# netleak Roadmap

## v0.2.0 - Testing and CI

- [ ] Unit tests for Go code (cgroup, routing, BPF loading)
- [ ] Fuzz tests for argument parsing and edge cases
- [ ] CI pipelines (multi-distro build matrix)
- [ ] CodeQL security scanning (Go + C)
- [ ] `.deb` and `.rpm` packaging
- [ ] Automated release workflow (tag-triggered)
- [ ] Dockerfile with multi-stage build

## v0.3.0 - IPv6 and Multi-Interface

- [ ] IPv6 policy routing support
- [ ] Multiple interface support (route different cgroups through different interfaces)
- [ ] Ingress filtering (drop inbound traffic not from the target interface)
- [ ] Interface auto-detection by gateway or IP range

## v0.4.0 - DNS Leak Prevention

- [ ] DNS leak prevention (force DNS through the target interface)
- [ ] Split tunneling by destination (allow specific CIDRs to bypass)
- [ ] Custom fwmark and routing table via CLI flags
- [ ] Configuration file support (TOML or YAML)

## v0.5.0 - Daemon Mode

- [ ] Daemon mode (`netleakd`) with D-Bus or Unix socket API
- [ ] NetworkManager plugin integration
- [ ] Systemd service unit
- [ ] Per-user and per-application policies
- [ ] Session persistence across reboots

## v1.0.0 - Stable Release

- [ ] Stable CLI and BPF map ABI
- [ ] Man pages (`netleak(8)`)
- [ ] Full documentation site
- [ ] Security audit
- [ ] ARM64 and cross-compilation support
- [ ] Performance benchmarks and optimization pass
