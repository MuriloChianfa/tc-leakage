#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Aggregate per-scenario benchmark JSON results into a combined results.json
# and results.csv. Optionally compare against a baseline to detect regressions.
#
# Usage: aggregate-results.sh [results-dir] [baseline-file]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

RESULTS_DIR="${1:-${E2E_DIR}/results}"
BASELINE_FILE="${2:-${E2E_DIR}/baseline.json}"
REGRESSION_THRESHOLD="${BENCH_REGRESSION_THRESHOLD:-5}"

if [ ! -d "${RESULTS_DIR}" ] || [ -z "$(ls -A "${RESULTS_DIR}"/*.json 2>/dev/null)" ]; then
    echo "No benchmark results found in ${RESULTS_DIR}"
    exit 1
fi

# ---------------------------------------------------------------------------
# Merge per-scenario JSON files into results.json
# ---------------------------------------------------------------------------

echo "Aggregating benchmark results from ${RESULTS_DIR}..."

COMBINED="${RESULTS_DIR}/results.json"
jq -s '.' "${RESULTS_DIR}"/wireguard.json \
          "${RESULTS_DIR}"/openvpn.json \
          "${RESULTS_DIR}"/strongswan.json \
          "${RESULTS_DIR}"/softether.json \
          2>/dev/null > "${COMBINED}" || \
    jq -s '.' "${RESULTS_DIR}"/*.json > "${COMBINED}"

echo "Combined results written to ${COMBINED}"

# ---------------------------------------------------------------------------
# Generate CSV
# ---------------------------------------------------------------------------

CSV="${RESULTS_DIR}/results.csv"

echo "scenario,metric,baseline,netleak,delta_pct" > "${CSV}"

jq -r '.[] |
    .scenario as $s |
    (.throughput_single | "\($s),throughput_single_mbps,\(.baseline_mbps),\(.netleak_mbps),\(.delta_pct)"),
    (.throughput_multi  | "\($s),throughput_multi_mbps,\(.baseline_mbps),\(.netleak_mbps),\(.delta_pct)"),
    (.latency           | "\($s),latency_avg_ms,\(.baseline_avg_ms),\(.netleak_avg_ms),\(.delta_pct)")
' "${COMBINED}" >> "${CSV}"

echo "CSV results written to ${CSV}"

# ---------------------------------------------------------------------------
# Print summary table
# ---------------------------------------------------------------------------

echo ""
echo "=== Benchmark Summary ==="
printf "%-12s %-24s %12s %12s %10s\n" "Scenario" "Metric" "Baseline" "Netleak" "Delta%"
printf "%-12s %-24s %12s %12s %10s\n" "--------" "------" "--------" "-------" "------"

tail -n +2 "${CSV}" | while IFS=',' read -r scenario metric baseline netleak delta; do
    printf "%-12s %-24s %12s %12s %9s%%\n" \
        "${scenario}" "${metric}" "${baseline}" "${netleak}" "${delta}"
done

echo ""

# ---------------------------------------------------------------------------
# Regression detection against baseline
# ---------------------------------------------------------------------------

if [ ! -f "${BASELINE_FILE}" ]; then
    echo "No baseline file at ${BASELINE_FILE} — skipping regression check."
    echo "To enable regression detection, run benchmarks and copy results.json to baseline.json."
    exit 0
fi

echo "Checking for regressions against ${BASELINE_FILE} (threshold: ${REGRESSION_THRESHOLD}%)..."

REGRESSION_LOG=$(mktemp)
trap 'rm -f "${REGRESSION_LOG}"' EXIT

check_regression() {
    local scenario="$1" metric="$2" current_delta="$3"
    local abs_delta
    abs_delta=$(awk "BEGIN { d = ${current_delta}; if (d < 0) d = -d; printf \"%.2f\", d }")

    local baseline_delta
    baseline_delta=$(jq -r --arg s "${scenario}" --arg m "${metric}" '
        .[] | select(.scenario == $s) |
        if $m == "throughput_single" then .throughput_single.delta_pct
        elif $m == "throughput_multi" then .throughput_multi.delta_pct
        elif $m == "latency" then .latency.delta_pct
        else 0 end
    ' "${BASELINE_FILE}" 2>/dev/null || echo "0")

    if [ -z "${baseline_delta}" ] || [ "${baseline_delta}" = "null" ]; then
        return
    fi

    local abs_baseline
    abs_baseline=$(awk "BEGIN { d = ${baseline_delta}; if (d < 0) d = -d; printf \"%.2f\", d }")

    local worsened
    worsened=$(awk "BEGIN {
        diff = ${abs_delta} - ${abs_baseline};
        if (diff > ${REGRESSION_THRESHOLD}) print 1; else print 0
    }")

    if [ "${worsened}" = "1" ]; then
        echo "  REGRESSION: ${scenario}/${metric}: delta ${current_delta}% (was ${baseline_delta}%, threshold ${REGRESSION_THRESHOLD}%)"
        echo "1" >> "${REGRESSION_LOG}"
    fi
}

while read -r s m d; do check_regression "$s" "$m" "$d"; done < \
    <(jq -r '.[] | "\(.scenario) throughput_single \(.throughput_single.delta_pct)"' "${COMBINED}")

while read -r s m d; do check_regression "$s" "$m" "$d"; done < \
    <(jq -r '.[] | "\(.scenario) throughput_multi \(.throughput_multi.delta_pct)"' "${COMBINED}")

while read -r s m d; do check_regression "$s" "$m" "$d"; done < \
    <(jq -r '.[] | "\(.scenario) latency \(.latency.delta_pct)"' "${COMBINED}")

regressions=$(wc -l < "${REGRESSION_LOG}" 2>/dev/null || echo 0)

if [ "${regressions}" -gt 0 ]; then
    echo ""
    echo "FAIL: ${regressions} regression(s) detected!"
    exit 1
fi

echo "No regressions detected."
exit 0
