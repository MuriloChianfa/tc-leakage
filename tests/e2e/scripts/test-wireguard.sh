#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# WireGuard e2e test: verify netleak routes traffic through wg0,
# kill-switch activates on interface down, and recovers on interface up.
# Tests both IPv4 and IPv6 traffic.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

SSH_KEY="$1"
SSH_PORT="$2"
IFACE="wg0"
URL4="http://10.10.0.1:8080/"
URL6="http://[fd10::1]:8080/"

tap_begin

wait_for_http "${SSH_KEY}" "${SSH_PORT}" "${URL4}"

# Test 1: IPv4 traffic routes through WireGuard tunnel
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL4}" \
    "IPv4 traffic routes through ${IFACE}"

# Test 2: IPv6 traffic routes through WireGuard tunnel
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL6}" \
    "IPv6 traffic routes through ${IFACE}"

# Test 3: Kill-switch activates when wg0 goes down (IPv4)
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${IFACE} down"
sleep 2
assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL4}" \
    "kill-switch drops IPv4 traffic when ${IFACE} is down"

# Test 4: Kill-switch activates when wg0 goes down (IPv6)
assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL6}" \
    "kill-switch drops IPv6 traffic when ${IFACE} is down"

# Test 5: IPv4 traffic recovers when wg0 comes back up
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${IFACE} up"
sleep 3
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL4}" \
    "IPv4 traffic recovers when ${IFACE} comes back up"

# Test 6: IPv6 traffic recovers when wg0 comes back up
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL6}" \
    "IPv6 traffic recovers when ${IFACE} comes back up"

# Test 7: Ingress filtering (opt-in via --ingress-filter) — traffic from
# the underlay IP is unreachable inside the netleak cgroup because the
# ingress BPF program blocks packets arriving on a non-tunnel interface.
if vm_ssh "${SSH_KEY}" "${SSH_PORT}" \
    "ip netns exec vpn-server sh -c 'cd /tmp && nohup python3 -m http.server 8082 --bind 10.0.0.1 > /tmp/http-underlay.log 2>&1 &'" 2>/dev/null; then
    sleep 1
    assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "http://10.0.0.1:8082/" \
        "ingress filter blocks non-tunnel traffic (--ingress-filter)" "--ingress-filter"
fi

tap_end
