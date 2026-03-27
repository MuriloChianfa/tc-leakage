#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Multi-interface e2e test: verify two concurrent netleak sessions
# (wg0 + tun0) are isolated. Kill-switch on one interface must not
# affect the other.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

SSH_KEY="$1"
SSH_PORT="$2"
WG_IFACE="wg0"
WG_URL="http://10.10.0.1:8080/"
OVPN_IFACE="tun0"
OVPN_URL="http://10.20.0.1:8081/"

tap_begin

wait_for_http "${SSH_KEY}" "${SSH_PORT}" "${WG_URL}"
wait_for_http "${SSH_KEY}" "${SSH_PORT}" "${OVPN_URL}"

# Test 1: WireGuard session routes through wg0
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${WG_IFACE}" "${WG_URL}" \
    "wg0 session routes through WireGuard"

# Test 2: OpenVPN session routes through tun0
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${OVPN_IFACE}" "${OVPN_URL}" \
    "tun0 session routes through OpenVPN"

# Test 3: Bring down wg0, tun0 session still works
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${WG_IFACE} down"
sleep 2
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${OVPN_IFACE}" "${OVPN_URL}" \
    "tun0 unaffected when wg0 goes down"

# Test 4: wg0 session fails (kill-switch)
assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${WG_IFACE}" "${WG_URL}" \
    "wg0 kill-switch while tun0 still active"

# Test 5: Bring wg0 back up, both sessions work
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${WG_IFACE} up"
sleep 3
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${WG_IFACE}" "${WG_URL}" \
    "wg0 session recovers"

assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${OVPN_IFACE}" "${OVPN_URL}" \
    "tun0 session still works after wg0 recovery"

tap_end
