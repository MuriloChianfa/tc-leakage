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

## Verifying Binary Signatures

netleak binaries are cryptographically signed with GPG for authenticity verification. To verify a downloaded binary:

### 1. Import the Public Key

Import the maintainer's public key directly from the keyserver using the key fingerprint:

```bash
gpg --keyserver keys.openpgp.org --recv-keys 3E1A1F401A1C47BC77D1705612D0D82387FC53B0
```

<details>
<summary><b>Alternative options</b></summary>

Using the shorter key ID:

```bash
gpg --keyserver keys.openpgp.org --recv-keys 12D0D82387FC53B0
```

**Alternative keyserver** (if `keys.openpgp.org` is unavailable):

```bash
gpg --keyserver hkp://keyserver.ubuntu.com --recv-keys 3E1A1F401A1C47BC77D1705612D0D82387FC53B0
```

</details>

You should see output confirming the key was imported:
```
gpg: key 12D0D82387FC53B0: public key "MuriloChianfa <murilo.chianfa@outlook.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1
```

### 2. Verify the Signature

Assuming you have downloaded both the binary package (e.g., `netleak_1.0.0_amd64.deb`) and its signature file (e.g., `netleak_1.0.0_amd64.deb.asc`):

```bash
gpg --verify netleak_1.0.0_amd64.deb.asc netleak_1.0.0_amd64.deb
```

If the signature is valid, you should see:
```
gpg: Signature made [date and time]
gpg:                using EDDSA key 3E1A1F401A1C47BC77D1705612D0D82387FC53B0
gpg: Good signature from "MuriloChianfa <murilo.chianfa@outlook.com>"
```

If you see "BAD signature", **do not use** the binary - it may have been tampered with or corrupted.

### 3. Verify Checksums (Additional Layer)

For extra security, you can also verify the checksums:

```bash
# Download checksums and signature
curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/SHA256SUMS
curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/SHA256SUMS.asc

# Verify the checksums signature
gpg --verify SHA256SUMS.asc SHA256SUMS

# Verify the package checksum
sha256sum -c SHA256SUMS --ignore-missing
```

## Non-Goals

- No network namespace isolation
- No container runtime integration
- No raw AF_PACKET injection support

## License

This project uses dual licensing:

- **Go source code** (`cmd/`): [MIT License](LICENSE)
- **eBPF/C source code** (`bpf/`): [GPL-3.0](LICENSE-GPL)

Each source file contains an SPDX license identifier header indicating its applicable license.
