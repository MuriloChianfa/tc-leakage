#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Shared helpers for VPN throughput/latency benchmarks.
# Sourced by per-scenario bench-*.sh scripts.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

BENCH_DURATION="${BENCH_DURATION:-10}"
BENCH_STREAMS="${BENCH_STREAMS:-4}"
BENCH_PING_COUNT="${BENCH_PING_COUNT:-20}"
BENCH_IPERF_PORT="${BENCH_IPERF_PORT:-5201}"
RESULTS_DIR="${RESULTS_DIR:-$(cd "${SCRIPT_DIR}/.." && pwd)/results}"

mkdir -p "${RESULTS_DIR}"

# ---------------------------------------------------------------------------
# iperf3 helpers
# ---------------------------------------------------------------------------

run_iperf3_baseline() {
    local key="$1" port="$2" server_ip="$3" extra_args="${4:-}"
    vm_ssh "${key}" "${port}" \
        "iperf3 -c ${server_ip} -p ${BENCH_IPERF_PORT} -t ${BENCH_DURATION} --json ${extra_args}" 2>/dev/null
}

run_iperf3_netleak() {
    local key="$1" port="$2" iface="$3" server_ip="$4" extra_args="${5:-}"
    vm_ssh "${key}" "${port}" \
        "netleak ${iface} iperf3 -c ${server_ip} -p ${BENCH_IPERF_PORT} -t ${BENCH_DURATION} --json ${extra_args}" 2>/dev/null
}

extract_throughput() {
    local json="$1"
    printf '%s' "${json}" | jq -r '
        .end.sum_sent.bits_per_second / 1000000 |
        . * 100 | floor / 100
    '
}

extract_throughput_received() {
    local json="$1"
    printf '%s' "${json}" | jq -r '
        .end.sum_received.bits_per_second / 1000000 |
        . * 100 | floor / 100
    '
}

# ---------------------------------------------------------------------------
# ping / latency helpers
# ---------------------------------------------------------------------------

run_ping_baseline() {
    local key="$1" port="$2" target_ip="$3" count="${4:-${BENCH_PING_COUNT}}"
    vm_ssh "${key}" "${port}" \
        "ping -c ${count} -q ${target_ip}" 2>/dev/null
}

run_ping_netleak() {
    local key="$1" port="$2" iface="$3" target_ip="$4" count="${5:-${BENCH_PING_COUNT}}"
    vm_ssh "${key}" "${port}" \
        "netleak ${iface} ping -c ${count} -q ${target_ip}" 2>/dev/null
}

extract_ping_avg() {
    local output="$1"
    printf '%s' "${output}" | \
        grep 'rtt\|round-trip' | \
        sed -E 's|.*= ([0-9.]+)/([0-9.]+)/([0-9.]+).*|\2|'
}

# ---------------------------------------------------------------------------
# delta / comparison
# ---------------------------------------------------------------------------

compute_delta() {
    local baseline="$1" measured="$2"
    awk "BEGIN {
        if (${baseline} == 0) { printf \"0.00\"; exit }
        printf \"%.2f\", ((${measured} - ${baseline}) / ${baseline}) * 100
    }"
}

# ---------------------------------------------------------------------------
# JSON result emitter
# ---------------------------------------------------------------------------

emit_result() {
    local scenario="$1" json_blob="$2"
    local outfile="${RESULTS_DIR}/${scenario}.json"
    printf '%s\n' "${json_blob}" > "${outfile}"
    echo "# Benchmark results written to ${outfile}"
}

# ---------------------------------------------------------------------------
# Full benchmark suite for a single scenario
# ---------------------------------------------------------------------------

run_bench_suite() {
    local key="$1" port="$2" iface="$3" server_ip="$4" scenario="$5"

    echo "# ============================================="
    echo "# Benchmark: ${scenario} (${iface}) -> ${server_ip}"
    echo "# Duration: ${BENCH_DURATION}s | Streams: ${BENCH_STREAMS}"
    echo "# ============================================="

    # --- Single-stream TCP throughput ---
    echo "# [1/6] Single-stream TCP baseline..."
    local baseline_single_json
    baseline_single_json=$(run_iperf3_baseline "${key}" "${port}" "${server_ip}")
    local baseline_single_mbps
    baseline_single_mbps=$(extract_throughput "${baseline_single_json}")
    echo "#   Baseline: ${baseline_single_mbps} Mbps"

    echo "# [2/6] Single-stream TCP with netleak..."
    local netleak_single_json
    netleak_single_json=$(run_iperf3_netleak "${key}" "${port}" "${iface}" "${server_ip}")
    local netleak_single_mbps
    netleak_single_mbps=$(extract_throughput "${netleak_single_json}")
    echo "#   Netleak:  ${netleak_single_mbps} Mbps"

    local delta_single
    delta_single=$(compute_delta "${baseline_single_mbps}" "${netleak_single_mbps}")
    echo "#   Delta:    ${delta_single}%"

    # --- Ping latency ---
    echo "# [3/6] Ping latency baseline (${BENCH_PING_COUNT} pings)..."
    local baseline_ping_out
    baseline_ping_out=$(run_ping_baseline "${key}" "${port}" "${server_ip}")
    local baseline_ping_avg
    baseline_ping_avg=$(extract_ping_avg "${baseline_ping_out}")
    echo "#   Baseline avg RTT: ${baseline_ping_avg} ms"

    echo "# [4/6] Ping latency with netleak..."
    local netleak_ping_out
    netleak_ping_out=$(run_ping_netleak "${key}" "${port}" "${iface}" "${server_ip}")
    local netleak_ping_avg
    netleak_ping_avg=$(extract_ping_avg "${netleak_ping_out}")
    echo "#   Netleak avg RTT:  ${netleak_ping_avg} ms"

    local delta_latency
    delta_latency=$(compute_delta "${baseline_ping_avg}" "${netleak_ping_avg}")
    echo "#   Delta:            ${delta_latency}%"

    # --- Multi-stream TCP throughput ---
    echo "# [5/6] Multi-stream TCP baseline (${BENCH_STREAMS} streams)..."
    local baseline_multi_json
    baseline_multi_json=$(run_iperf3_baseline "${key}" "${port}" "${server_ip}" "-P ${BENCH_STREAMS}")
    local baseline_multi_mbps
    baseline_multi_mbps=$(extract_throughput "${baseline_multi_json}")
    echo "#   Baseline: ${baseline_multi_mbps} Mbps"

    echo "# [6/6] Multi-stream TCP with netleak (${BENCH_STREAMS} streams)..."
    local netleak_multi_json
    netleak_multi_json=$(run_iperf3_netleak "${key}" "${port}" "${iface}" "${server_ip}" "-P ${BENCH_STREAMS}")
    local netleak_multi_mbps
    netleak_multi_mbps=$(extract_throughput "${netleak_multi_json}")
    echo "#   Netleak:  ${netleak_multi_mbps} Mbps"

    local delta_multi
    delta_multi=$(compute_delta "${baseline_multi_mbps}" "${netleak_multi_mbps}")
    echo "#   Delta:    ${delta_multi}%"

    # --- Build result JSON ---
    local arch
    arch=$(uname -m)
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    local result_json
    result_json=$(jq -n \
        --arg scenario "${scenario}" \
        --arg iface "${iface}" \
        --arg arch "${arch}" \
        --arg ts "${timestamp}" \
        --argjson duration "${BENCH_DURATION}" \
        --argjson streams "${BENCH_STREAMS}" \
        --argjson qemu_mem "${BENCH_QEMU_MEM:-2048}" \
        --argjson qemu_cpus "${BENCH_QEMU_CPUS:-2}" \
        --argjson bs "${baseline_single_mbps}" \
        --argjson ns "${netleak_single_mbps}" \
        --argjson ds "${delta_single}" \
        --argjson bm "${baseline_multi_mbps}" \
        --argjson nm "${netleak_multi_mbps}" \
        --argjson dm "${delta_multi}" \
        --argjson bla "${baseline_ping_avg}" \
        --argjson nla "${netleak_ping_avg}" \
        --argjson dl "${delta_latency}" \
        '{
            scenario: $scenario,
            interface: $iface,
            arch: $arch,
            timestamp: $ts,
            config: {
                duration_s: $duration,
                streams: $streams,
                qemu_mem_mb: $qemu_mem,
                qemu_cpus: $qemu_cpus
            },
            throughput_single: {
                baseline_mbps: $bs,
                netleak_mbps: $ns,
                delta_pct: $ds
            },
            throughput_multi: {
                streams: $streams,
                baseline_mbps: $bm,
                netleak_mbps: $nm,
                delta_pct: $dm
            },
            latency: {
                baseline_avg_ms: $bla,
                netleak_avg_ms: $nla,
                delta_pct: $dl
            }
        }')

    emit_result "${scenario}" "${result_json}"

    echo "# ============================================="
    echo "# ${scenario} benchmark complete"
    echo "# ============================================="
}
