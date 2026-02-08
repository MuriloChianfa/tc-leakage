# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa

# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM golang:1.24-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    llvm \
    libbpf-dev \
    libelf-dev \
    make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY . .

RUN make all

# Verify the build artifacts exist
RUN file /src/netleak && file /src/bpf/netleak.o

# =============================================================================
# Stage 2: Minimal Runtime
# =============================================================================
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    libbpf1 \
    libelf1 \
    && rm -rf /var/lib/apt/lists/*

# Install binary and BPF object
COPY --from=builder /src/netleak /usr/bin/netleak
COPY --from=builder /src/bpf/netleak.o /usr/lib/netleak/netleak.o

# NOTE: Running netleak requires:
#   --privileged (or CAP_SYS_ADMIN + CAP_NET_ADMIN + CAP_BPF)
#   --cgroupns=host (cgroup v2 access)
#   -v /sys/fs/bpf:/sys/fs/bpf (BPF filesystem)
#   -v /sys/fs/cgroup:/sys/fs/cgroup (cgroup filesystem)

ENTRYPOINT ["netleak"]
