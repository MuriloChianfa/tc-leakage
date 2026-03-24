#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Shared helpers for e2e VPN test scripts.

set -euo pipefail

TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

SSH_OPTS="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR -o ConnectTimeout=10"

tap_begin() {
    echo "TAP version 13"
}

tap_ok() {
    local desc="$1"
    TEST_COUNT=$((TEST_COUNT + 1))
    PASS_COUNT=$((PASS_COUNT + 1))
    echo "ok ${TEST_COUNT} - ${desc}"
}

tap_not_ok() {
    local desc="$1"
    TEST_COUNT=$((TEST_COUNT + 1))
    FAIL_COUNT=$((FAIL_COUNT + 1))
    echo "not ok ${TEST_COUNT} - ${desc}"
}

tap_end() {
    echo "1..${TEST_COUNT}"
    if [ "${FAIL_COUNT}" -gt 0 ]; then
        echo "# ${FAIL_COUNT}/${TEST_COUNT} tests failed"
        return 1
    fi
    echo "# All ${TEST_COUNT} tests passed"
    return 0
}

vm_ssh() {
    local key="$1" port="$2"
    shift 2
    # shellcheck disable=SC2086
    ssh ${SSH_OPTS} -i "${key}" -p "${port}" root@localhost "$@"
}

vm_scp() {
    local key="$1" port="$2" src="$3" dst="$4"
    # shellcheck disable=SC2086
    scp ${SSH_OPTS} -i "${key}" -P "${port}" "${src}" "root@localhost:${dst}"
}

wait_for_ssh() {
    local key="$1" port="$2" timeout="${3:-120}"
    local start elapsed
    start=$(date +%s)
    echo "Waiting for VM SSH (port ${port})..."
    while true; do
        if vm_ssh "${key}" "${port}" "true" 2>/dev/null; then
            echo "VM SSH is ready."
            return 0
        fi
        elapsed=$(( $(date +%s) - start ))
        if [ "${elapsed}" -ge "${timeout}" ]; then
            echo "Timeout waiting for VM SSH after ${timeout}s"
            return 1
        fi
        sleep 2
    done
}

assert_curl_ok() {
    local key="$1" port="$2" iface="$3" url="$4" desc="${5:-routing through ${iface}}"
    local max_attempts=6
    local attempt
    for attempt in $(seq 1 "${max_attempts}"); do
        if vm_ssh "${key}" "${port}" "netleak ${iface} curl -sf --connect-timeout 10 ${url}" 2>/dev/null; then
            tap_ok "${desc}"
            return
        fi
        [ "${attempt}" -lt "${max_attempts}" ] && sleep 3
    done
    tap_not_ok "${desc}"
}

assert_curl_fails() {
    local key="$1" port="$2" iface="$3" url="$4" desc="${5:-kill-switch on ${iface}}"
    if vm_ssh "${key}" "${port}" "netleak ${iface} curl -sf --connect-timeout 5 ${url}" 2>/dev/null; then
        tap_not_ok "${desc}"
    else
        tap_ok "${desc}"
    fi
}

wait_for_http() {
    local key="$1" port="$2" url="$3" timeout="${4:-30}"
    local start elapsed
    start=$(date +%s)
    while true; do
        if vm_ssh "${key}" "${port}" "curl -sf --connect-timeout 3 ${url}" 2>/dev/null; then
            return 0
        fi
        elapsed=$(( $(date +%s) - start ))
        if [ "${elapsed}" -ge "${timeout}" ]; then
            echo "# Warning: HTTP listener at ${url} not reachable after ${timeout}s"
            return 1
        fi
        sleep 2
    done
}

detect_kvm() {
    if [ -w /dev/kvm ]; then
        echo "-enable-kvm"
    else
        echo ""
    fi
}
