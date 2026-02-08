<h1 align="center">netleak</h1>

<p align="center">
  A kernel-level, proxychains-like tool for per-process traffic redirection on Linux.<br><br>
  Built on top of <strong>cgroup v2</strong> and <strong>eBPF</strong>, netleak forces all network traffic from a process
  and its entire child tree through a specific network interface, completely
  transparently to the application. If the target interface goes down, a
  kernel-enforced <strong>kill-switch</strong> drops every packet instead of falling back to the
  default route, guaranteeing zero traffic leakage under any failure condition.
</p>

---

## Install from Package

<details name="install" open>
  <summary style="font-size: 16px;"><strong>Ubuntu/Debian</strong></summary>

  Download the `.deb` package from the [latest release](https://github.com/MuriloChianfa/netleak/releases/latest) and install it:

  ```bash
  curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/netleak_1.0.0_amd64.deb
  sudo dpkg -i netleak_1.0.0_amd64.deb
  ```
</details>
<details name="install">
  <summary style="font-size: 16px;"><strong>Fedora/RHEL/Rocky/AlmaLinux</strong></summary>

  Download the `.rpm` package from the [latest release](https://github.com/MuriloChianfa/netleak/releases/latest) and install it:

  ```bash
  curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/netleak-1.0.0-1.x86_64.rpm
  sudo rpm -i netleak-1.0.0-1.x86_64.rpm
  ```
</details>

## Or Build & Install from Source

**Requirements:** Linux kernel >= 5.8, cgroup v2, clang/llvm >= 10, Go >= 1.24

<details name="build-source" open>
  <summary style="font-size: 16px;"><strong>Ubuntu/Debian</strong></summary>

  ```bash
  # Install build dependencies
  sudo apt update
  sudo apt install -y clang llvm libbpf-dev libelf-dev make golang

  # Clone and build
  git clone https://github.com/MuriloChianfa/netleak.git
  cd netleak
  make
  ```
</details>
<details name="build-source">
  <summary style="font-size: 16px;"><strong>Fedora/RHEL/Rocky Linux</strong></summary>

  ```bash
  # Install build dependencies
  sudo dnf install -y clang llvm libbpf-devel elfutils-libelf-devel make golang

  # Clone and build
  git clone https://github.com/MuriloChianfa/netleak.git
  cd netleak
  make
  ```
</details>

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


### Kill-Switch

When the target interface goes down:
- Userspace detects it via netlink subscription
- Sets `FLAG_KILL_SWITCH` in the BPF map entry
- The eBPF program drops all packets from the affected cgroup
- No fallback to the default route ever occurs

When the interface comes back up, the flag is cleared and traffic resumes.

## Verifying Binary Signatures

netleak binaries are cryptographically signed with GPG for authenticity verification. To verify a downloaded binary:

### 1. Verify Checksums

Download the checksum file and verify the integrity of your package:

```bash
# Download checksums and signature
curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/SHA256SUMS
curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.0.0/SHA256SUMS.asc

# Verify the package checksum
sha256sum -c SHA256SUMS --ignore-missing
```

### 2. Import the Public Key

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

### 3. Verify the GPG Signature

Verify both the checksums file signature and the package signature:

```bash
# Verify the checksums signature
gpg --verify SHA256SUMS.asc SHA256SUMS
```

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

## Security

For details on reporting vulnerabilities and our security practices, see the [Security Policy](https://github.com/MuriloChianfa/netleak/security/policy).

## License

This project is dual-licensed:

| Component | License | File |
|---|---|---|
| Go source | MIT | [LICENSE](LICENSE) |
| eBPF/C source | GPL-3.0-only | [LICENSE-GPL](LICENSE-GPL) |

Each source file contains an SPDX license identifier header.
