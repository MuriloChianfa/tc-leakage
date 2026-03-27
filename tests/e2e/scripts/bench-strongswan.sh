#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# strongSwan/IPsec (vti0) throughput and latency benchmark.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/bench-common.sh"

SSH_KEY="$1"
SSH_PORT="$2"
IFACE="vti0"
SERVER_IP="10.10.0.1"
SCENARIO="strongswan"

wait_for_http "${SSH_KEY}" "${SSH_PORT}" "http://${SERVER_IP}:8080/"

run_bench_suite "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${SERVER_IP}" "${SCENARIO}"
