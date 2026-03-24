# netleak Roadmap

## v1.0.0 - Testing and CI

- [x] Unit tests for Go code (cgroup, routing, BPF loading)
- [x] Fuzz tests for argument parsing and edge cases
- [x] CI pipelines (multi-distro build matrix)
- [x] CodeQL security scanning (Go + C)
- [x] `.deb` and `.rpm` packaging
- [x] Automated release workflow (tag-triggered)
- [x] Dockerfile with multi-stage build

## v1.0.1 - Integration Testing (QEMU/KVM)

- [ ] QEMU/KVM test harness
  - [ ] VM boot scripts (QEMU launch with KVM auto-detection, TCG fallback)
  - [ ] Alpine cloud image download and cloud-init seed ISO generation
  - [ ] SSH-based VM access via user-mode networking port forward
  - [ ] Ansible provisioning for declarative, idempotent VM configuration
  - [ ] Shared test helpers and TAP-style output reporting
- [ ] WireGuard integration test
  - [ ] Ansible role: namespace + veth underlay, keypair generation, `wg0` tunnel
  - [ ] Traffic routing verification (`netleak wg0 curl ...`)
  - [ ] Kill-switch test (interface down -> traffic dropped)
  - [ ] Recovery test (interface up -> traffic resumes)
- [ ] OpenVPN integration test
  - [ ] Ansible role: ephemeral PKI (CA + certs), TLS tunnel with `--dev tun`
  - [ ] Traffic routing verification (`netleak tun0 curl ...`)
  - [ ] Kill-switch and recovery validation
- [ ] strongSwan/IPsec integration test
  - [ ] Ansible role: IKEv2 with PSK, route-based VTI (`vti0` or `xfrm`) interface
  - [ ] Traffic routing verification (`netleak vti0 curl ...`)
  - [ ] Kill-switch and recovery validation
- [ ] SoftEther VPN integration test
  - [ ] Ansible role: SoftEther server with virtual hub, `tap0` bridge interface
  - [ ] Traffic routing verification (`netleak tap0 curl ...`)
  - [ ] Kill-switch and recovery validation (layer-2 tap device)
- [ ] GitHub Actions CI workflow for e2e tests (TCG mode, parallel job matrix)
- [ ] Test artifact collection and log upload

## v1.1.0 - IPv6 and Multi-Interface

- [ ] IPv6 policy routing support
- [ ] Multiple interface support (route different cgroups through different interfaces)
- [ ] Ingress filtering (drop inbound traffic not from the target interface)
- [ ] Interface auto-detection by gateway or IP range
- [ ] Stable CLI and BPF map ABI
- [ ] ARM64 and cross-compilation support

## v1.1.1 - VPN Throughput Benchmarks

- [ ] Benchmark harness (QEMU/KVM, reuses e2e VM infrastructure, for both x86 and arm)
  - [ ] `iperf3` server/client inside VPN tunnel namespaces
  - [ ] Baseline measurement: `iperf3` direct through VPN interface (no netleak)
  - [ ] Netleak measurement: `netleak <iface> iperf3 -c ...`
  - [ ] Automated comparison and delta-percentage calculation
- [ ] WireGuard (`wg0`) throughput benchmark
- [ ] OpenVPN (`tun0`) throughput benchmark
- [ ] strongSwan/IPsec (`vti0`) throughput benchmark
- [ ] SoftEther (`tap0`) throughput benchmark
- [ ] Latency benchmarks (RTT via `ping` with and without netleak)
- [ ] Multi-stream benchmarks (parallel `iperf3` sessions)
- [ ] Results output in machine-readable format (JSON/CSV)
- [ ] CI job: run benchmarks and upload results as artifacts
- [ ] Historical tracking: compare against previous runs to detect regressions

## v1.2.0 - Daemon Mode

- [ ] Daemon mode (`netleakd`) with D-Bus or Unix socket API
- [ ] NetworkManager plugin integration
- [ ] Systemd service unit
- [ ] Per-user and per-application policies
- [ ] Session persistence across reboots
- [ ] Configuration file support (TOML or YAML)
- [ ] Man pages (`netleak(8)`)
- [ ] Full documentation site
