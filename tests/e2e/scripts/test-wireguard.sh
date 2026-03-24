#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# WireGuard e2e test: verify netleak routes traffic through wg0,
# kill-switch activates on interface down, and recovers on interface up.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

SSH_KEY="$1"
SSH_PORT="$2"
IFACE="wg0"
URL="http://10.10.0.1:8080/"

tap_begin

wait_for_http "${SSH_KEY}" "${SSH_PORT}" "${URL}"

# Test 1: Traffic routes through WireGuard tunnel
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "traffic routes through ${IFACE}"

# Test 2: Kill-switch activates when wg0 goes down
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${IFACE} down"
sleep 2
assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "kill-switch drops traffic when ${IFACE} is down"

# Test 3: Traffic recovers when wg0 comes back up
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${IFACE} up"
sleep 3
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "traffic recovers when ${IFACE} comes back up"

tap_end
