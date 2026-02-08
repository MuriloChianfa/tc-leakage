# CodeQL Security Analysis

This repository uses CodeQL for automated security scanning of Go and C source code.

## Overview

The CodeQL workflow (`codeql.yml`) performs static analysis security testing (SAST) on:

- **Go Code**: CLI and userspace components in `cmd/`
- **C Code**: eBPF programs in `bpf/`

## Workflow Triggers

The CodeQL analysis runs automatically on:

1. **Push events** to `main` branch
2. **Pull requests** targeting `main` branch
3. **Scheduled runs** every Monday at 2:30 AM UTC
4. **Manual dispatch** via GitHub Actions UI

## What Gets Analyzed

### Go Analysis

- CLI entry point and orchestration (`cmd/main.go`)
- BPF loader and attachment (`cmd/bpf.go`)
- Cgroup v2 management (`cmd/cgroup.go`)
- Policy routing setup (`cmd/route.go`)
- Interface monitoring (`cmd/monitor.go`)
- Checks for:
  - Command injection
  - Path traversal
  - Resource leaks
  - Unsafe system calls
  - Privilege escalation patterns

### C/C++ Analysis

- eBPF programs (`bpf/netleak.c`)
- BPF header definitions (`bpf/netleak.h`)
- Checks for:
  - Buffer overflows
  - Integer overflows
  - Out-of-bounds access
  - Unsafe BPF map operations

## Query Suites

The workflow uses enhanced query suites:

- `security-extended`: Extended set of security queries
- `security-and-quality`: Security queries plus code quality checks

## Build Process

### Go Build

The Go code uses CodeQL's autobuild feature, which automatically detects and builds Go modules.

### C Build

The C (eBPF) code is built manually using:

```bash
clang -O2 -g -target bpf -Wall -Wextra -I/usr/include/bpf -c bpf/netleak.c -o bpf/netleak.o
```

## Viewing Results

### Pull Requests

- CodeQL findings appear as annotations directly in PRs
- New issues are highlighted in the diff view

### Security Tab

- Navigate to **Security** > **Code scanning alerts** in the repository
- Filter by language, severity, or rule

## Configuration

The analysis is configured in:

- `.github/workflows/codeql.yml` - Main workflow
- `.github/workflows/codeql/codeql-config.yml` - Path configuration and query suites

### Excluding Paths

To exclude certain paths from analysis, edit `.github/workflows/codeql/codeql-config.yml`:

```yaml
paths-ignore:
  - tests
  - pkg
  - '**/*.md'
```

## Best Practices

1. **Review findings promptly**: Address security issues as soon as they appear
2. **Don't dismiss without investigation**: Understand each finding before closing
3. **Monitor scheduled scans**: Weekly scans catch newly discovered vulnerability patterns

## Resources

- [CodeQL Documentation](https://codeql.github.com/docs/)
- [CodeQL for C/C++](https://codeql.github.com/docs/codeql-language-guides/codeql-for-cpp/)
- [CodeQL for Go](https://codeql.github.com/docs/codeql-language-guides/codeql-for-go/)
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)

## Support

For issues with CodeQL analysis:

1. Check the [Actions logs](../../actions)
2. Review [Security alerts](../../security/code-scanning)
3. Open an issue with the `security` label
4. Contact repository maintainers
