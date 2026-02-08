# Contributing to netleak

Thank you for your interest in contributing to netleak! We welcome contributions from the community and appreciate your efforts to make this tool better.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to murilo.chianfa@outlook.com.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/netleak.git
   cd netleak
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/MuriloChianfa/netleak.git
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Environment

### Requirements

- Linux kernel >= 5.8 (cgroup v2, cgroup-skb eBPF)
- cgroup v2 mounted at `/sys/fs/cgroup` (default on modern distros)
- `clang` and `llvm` >= 10
- `golang` >= 1.24
- `libbpf-dev`
- `libelf-dev`
- `make`

### Ubuntu/Debian

```bash
sudo apt update
sudo apt install -y clang llvm libbpf-dev libelf-dev make golang
```

### Fedora

```bash
sudo dnf install -y clang llvm libbpf-devel elfutils-libelf-devel golang make
```

### Build

```bash
make all
```

This compiles the BPF object (`bpf/netleak.o`) and the Go binary (`netleak`).

### Using Docker

You can also build using the provided Dockerfile:

```bash
docker build -t netleak-builder .
```

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

- **Bug fixes**: Fix issues reported in the issue tracker
- **New features**: Add new functionality or optimizations
- **Documentation**: Improve README, man pages, or code comments
- **Tests**: Add test cases or improve test coverage
- **Performance optimizations**: BPF program improvements, Go optimizations
- **Packaging**: Improve .deb/.rpm packaging or add new distribution support

### Before You Start

1. **Check existing issues** to see if someone is already working on it
2. **Open an issue** for discussion if you're proposing a significant change
3. **Keep changes focused**: One feature/fix per pull request
4. **Follow the coding standards** described below

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`gofmt`)
- Use `go vet` and `staticcheck` for static analysis
- Naming conventions: follow [Effective Go](https://go.dev/doc/effective_go)
- Keep functions focused and short
- Document exported symbols

### BPF/C Code Style

- Use tabs for indentation
- Maximum line length: 120 characters
- Follow kernel BPF coding conventions
- Document BPF program sections clearly
- Include comments explaining eBPF verifier constraints

### Include Order (C)

```c
#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>
#include "netleak.h"
```

### Commit Messages

```
Brief summary (50 chars or less)

More detailed explanation if needed. Explain the problem
this commit solves and why you chose this approach.

Closes #123
```

## Testing Guidelines

### Running Tests

```bash
# Unit tests
go test ./tests/unit/...

# Fuzz tests
go test -fuzz=. ./tests/fuzzing/ -fuzztime=30s
```

### Writing Tests

- Add tests in the `tests/` directory
- Use descriptive test names
- Test edge cases: invalid interfaces, missing permissions, bad arguments
- Note: actual BPF loading requires root and a compatible kernel; use build verification in CI

## Pull Request Process

### Before Submitting

1. **Update your branch** with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests**:
   ```bash
   go test ./tests/...
   ```

3. **Run linters**:
   ```bash
   gofmt -d ./cmd/
   go vet ./cmd/...
   ```

4. **Write clear commit messages**

### Submitting the PR

1. **Push your branch** to your fork
2. **Open a pull request** against the `main` branch
3. **Fill out the PR template** completely
4. **Link related issues** using "Closes #123" or "Fixes #456"
5. **Wait for CI** to complete
6. **Respond to review comments** promptly

### PR Templates

We provide templates for different types of PRs:

- General changes: Default template
- Bug fixes: `BUG_FIX_TEMPLATE.md`
- New features: `FEATURE_TEMPLATE.md`

Select the appropriate template when creating your PR.

### Review Process

- A maintainer will review your PR within a few days
- Address review comments by pushing new commits
- Once approved, a maintainer will merge your PR
- PRs require at least one approval from a maintainer

## Reporting Bugs

### Security Vulnerabilities

**DO NOT** open public issues for security vulnerabilities. Use email instead. See [SECURITY.md](SECURITY.md) for details.

### Regular Bugs

Use the **Bug Report** issue template and include:

1. **Description**: Clear description of the bug
2. **Environment**: OS, kernel version, Go version, clang version
3. **Steps to reproduce**: Minimal steps to reproduce the issue
4. **Expected behavior**: What you expected to happen
5. **Actual behavior**: What actually happened
6. **Additional context**: Logs, dmesg output, etc.

## Feature Requests

Use the **Feature Request** issue template and include:

1. **Problem statement**: What problem does this solve?
2. **Proposed solution**: Your suggested implementation
3. **Alternatives considered**: Other approaches you've thought about
4. **Additional context**: Use cases, examples, etc.

## Recognition

Contributors will be:

- Credited in release notes for significant contributions
- Mentioned in commit history

Thank you for contributing to netleak!
