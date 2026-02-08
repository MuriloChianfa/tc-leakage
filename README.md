# netleak

**cgroup-based eBPF per-process traffic redirection**

A proxychains-like tool at kernel level: route all traffic from a process (and its children) through an arbitrary network interface, with kill-switch semantics that drop packets instead of leaking when the interface goes down.

## Prerequisites

- Linux kernel >= 5.8 (cgroup v2, cgroup-skb eBPF)
- cgroup v2 mounted at `/sys/fs/cgroup` (default on modern distros)
- `clang` and `llvm` >= 10
- `golang` >= 1.24
- `libbpf-dev`
- `make`

### Installing Dependencies

```bash
sudo apt update
sudo apt install -y clang llvm libbpf-dev libelf-dev make golang
```

## Building

```bash
git clone git@github.com:MuriloChianfa/netleak.git
cd netleak
make
```

## Usage

```
sudo netleak <interface> <command> [args...]
```

### Examples

Route curl through a specific interface:

```bash
sudo netleak ppp0 curl ifconfig.me
```

Run a shell with all traffic going through a given interface:

```bash
sudo netleak wg0 bash
```

> Everything launched from that shell (and its children*), will have traffic routed through `wg0`.

### Verify

```bash
# This should show the IP of the target interface
sudo netleak ppp0 curl ifconfig.me

# Other system processes remain unaffected
curl ifconfig.me  # shows your real IP
```

## Architecture

```
                    ┌─────────────────────────────┐
                    │         netleak CLI          │
                    │  - create cgroup v2          │
                    │  - load BPF programs         │
                    │  - setup policy routing      │
                    │  - monitor target interface  │
                    └──────────┬──────────────────┘
                               │ fork + exec
                    ┌──────────▼──────────────────┐
                    │     target command           │
                    │  (inherits cgroup)           │
                    └──────────┬──────────────────┘
                               │ egress packet
                    ┌──────────▼──────────────────┐
                    │  eBPF (cgroup/sock_create)   │
                    │  sk->mark = fwmark           │
                    │                              │
                    │  eBPF (cgroup_skb/egress)    │
                    │  if kill_switch → DROP        │
                    └──────────┬──────────────────┘
                               │ fwmark routing
                    ┌──────────▼──────────────────┐
                    │   Policy Routing Table 100   │
                    │   default dev <interface>    │
                    └─────────────────────────────┘
```

### BPF Map

| Key (u64)  | Value                          |
|------------|--------------------------------|
| cgroup_id  | `{ fwmark: u32, flags: u32 }`  |

### Kill-Switch

When the target interface goes down:
- Userspace detects it via netlink subscription
- Sets `FLAG_KILL_SWITCH` in the BPF map entry
- The eBPF program drops all packets from the affected cgroup
- No fallback to the default route ever occurs

When the interface comes back up, the flag is cleared and traffic resumes.

## Security

- Root privileges required (eBPF + cgroups + routing)
- No fallback routing under any failure condition
- No source address spoofing is needed

## Non-Goals

- No network namespace isolation
- No container runtime integration
- No raw AF_PACKET injection support

## License

[MIT License](LICENSE)
