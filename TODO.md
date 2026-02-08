# netleak Roadmap

## v1.0.0 - Testing and CI

- [x] Unit tests for Go code (cgroup, routing, BPF loading)
- [x] Fuzz tests for argument parsing and edge cases
- [x] CI pipelines (multi-distro build matrix)
- [x] CodeQL security scanning (Go + C)
- [x] `.deb` and `.rpm` packaging
- [x] Automated release workflow (tag-triggered)
- [x] Dockerfile with multi-stage build

## v1.1.0 - IPv6 and Multi-Interface

- [ ] IPv6 policy routing support
- [ ] Multiple interface support (route different cgroups through different interfaces)
- [ ] Ingress filtering (drop inbound traffic not from the target interface)
- [ ] Interface auto-detection by gateway or IP range
- [ ] Stable CLI and BPF map ABI
- [ ] ARM64 and cross-compilation support

## v1.2.0 - Daemon Mode

- [ ] Daemon mode (`netleakd`) with D-Bus or Unix socket API
- [ ] NetworkManager plugin integration
- [ ] Systemd service unit
- [ ] Per-user and per-application policies
- [ ] Session persistence across reboots
- [ ] Configuration file support (TOML or YAML)
- [ ] Man pages (`netleak(8)`)
- [ ] Full documentation site
