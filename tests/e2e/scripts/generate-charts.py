#!/usr/bin/env python3
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# Generate benchmark result charts from results.json.
# Outputs PNG files into the same results directory.
#
# Usage: generate-charts.py [results-dir]

import json
import sys
import os
from pathlib import Path

import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.ticker as ticker
import numpy as np


COLORS = {
    "baseline": "#5B8DEE",
    "netleak": "#2ECC71",
    "positive": "#2ECC71",
    "negative": "#E74C3C",
    "neutral": "#95A5A6",
    "text": "#2C3E50",
    "grid": "#ECF0F1",
    "bg": "#FAFBFC",
    "bar_edge": "#FFFFFF",
}

SCENARIO_LABELS = {
    "wireguard": "WireGuard\n(wg0)",
    "openvpn": "OpenVPN\n(tun0)",
    "strongswan": "strongSwan\n(vti0)",
    "softether": "SoftEther\n(tap0)",
}


def setup_style():
    plt.rcParams.update({
        "figure.facecolor": COLORS["bg"],
        "axes.facecolor": "#FFFFFF",
        "axes.edgecolor": "#DEE2E6",
        "axes.labelcolor": COLORS["text"],
        "axes.titlecolor": COLORS["text"],
        "axes.grid": True,
        "axes.axisbelow": True,
        "grid.color": COLORS["grid"],
        "grid.linewidth": 0.8,
        "xtick.color": COLORS["text"],
        "ytick.color": COLORS["text"],
        "font.family": "sans-serif",
        "font.size": 11,
        "axes.titlesize": 14,
        "axes.titleweight": "bold",
        "axes.labelsize": 12,
        "figure.titlesize": 16,
        "figure.titleweight": "bold",
    })


def load_results(results_dir):
    path = Path(results_dir) / "results.json"
    with open(path) as f:
        return json.load(f)


def fmt_mbps(val):
    if val >= 1000:
        return f"{val / 1000:.1f} Gbps"
    return f"{val:.0f} Mbps"


def make_grouped_bar(ax, scenarios, baseline_vals, netleak_vals, ylabel, title,
                     log_scale=False, value_formatter=None,
                     higher_is_better=True):
    x = np.arange(len(scenarios))
    width = 0.32

    bars_b = ax.bar(x - width / 2, baseline_vals, width,
                    label="Direct (baseline)", color=COLORS["baseline"],
                    edgecolor=COLORS["bar_edge"], linewidth=0.5, zorder=3)
    bars_n = ax.bar(x + width / 2, netleak_vals, width,
                    label="Through netleak", color=COLORS["netleak"],
                    edgecolor=COLORS["bar_edge"], linewidth=0.5, zorder=3)

    if log_scale:
        ax.set_yscale("log")
        ax.yaxis.set_major_formatter(ticker.FuncFormatter(
            lambda v, _: fmt_mbps(v) if v >= 1 else f"{v:.3f}"
        ))

    ax.set_xticks(x)
    ax.set_xticklabels([SCENARIO_LABELS.get(s, s) for s in scenarios],
                       fontsize=10)
    ax.set_ylabel(ylabel)
    ax.set_title(title, pad=14)
    ax.legend(loc="upper right", framealpha=0.9, edgecolor="#DEE2E6")

    if not value_formatter:
        value_formatter = lambda v: f"{v:,.0f}"

    for bar_b, bar_n, bv, nv in zip(bars_b, bars_n, baseline_vals, netleak_vals):
        if bv == 0:
            continue
        delta = ((nv - bv) / bv) * 100
        if higher_is_better:
            color = COLORS["positive"] if delta >= 0 else COLORS["negative"]
        else:
            color = COLORS["negative"] if delta >= 0 else COLORS["positive"]
        sign = "+" if delta >= 0 else ""
        peak = max(bar_b.get_height(), bar_n.get_height())

        if log_scale:
            y_off = peak * 1.35
        else:
            y_off = peak * 1.06

        ax.annotate(f"{sign}{delta:.1f}%",
                    xy=(bar_b.get_x() + width, y_off),
                    ha="center", va="bottom", fontsize=9, fontweight="bold",
                    color=color)

    if log_scale:
        ax.set_ylim(top=max(max(baseline_vals), max(netleak_vals)) * 4)
    else:
        ax.set_ylim(top=max(max(baseline_vals), max(netleak_vals)) * 1.18)


def chart_throughput_single(data, results_dir):
    fig, ax = plt.subplots(figsize=(10, 6))
    scenarios = [d["scenario"] for d in data]
    baseline = [d["throughput_single"]["baseline_mbps"] for d in data]
    netleak = [d["throughput_single"]["netleak_mbps"] for d in data]

    make_grouped_bar(ax, scenarios, baseline, netleak,
                     ylabel="Throughput",
                     title="Single-Stream TCP Throughput (iperf3, 10s)",
                     log_scale=True, value_formatter=fmt_mbps)

    fig.tight_layout()
    out = Path(results_dir) / "throughput-single.png"
    fig.savefig(out, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"  {out}")


def chart_throughput_multi(data, results_dir):
    fig, ax = plt.subplots(figsize=(10, 6))
    scenarios = [d["scenario"] for d in data]
    streams = data[0]["throughput_multi"]["streams"]
    baseline = [d["throughput_multi"]["baseline_mbps"] for d in data]
    netleak = [d["throughput_multi"]["netleak_mbps"] for d in data]

    make_grouped_bar(ax, scenarios, baseline, netleak,
                     ylabel="Throughput",
                     title=f"Multi-Stream TCP Throughput ({streams} streams, iperf3, 10s)",
                     log_scale=True, value_formatter=fmt_mbps)

    fig.tight_layout()
    out = Path(results_dir) / "throughput-multi.png"
    fig.savefig(out, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"  {out}")


def chart_latency(data, results_dir):
    fig, ax = plt.subplots(figsize=(10, 6))
    scenarios = [d["scenario"] for d in data]
    baseline = [d["latency"]["baseline_avg_ms"] for d in data]
    netleak = [d["latency"]["netleak_avg_ms"] for d in data]

    make_grouped_bar(ax, scenarios, baseline, netleak,
                     ylabel="Avg RTT (ms)",
                     title="Ping Latency -- Average Round-Trip Time",
                     log_scale=False,
                     value_formatter=lambda v: f"{v:.3f} ms",
                     higher_is_better=False)

    fig.tight_layout()
    out = Path(results_dir) / "latency.png"
    fig.savefig(out, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"  {out}")


def _is_improvement(delta, metric_key):
    """Positive throughput delta = good; positive latency delta = bad."""
    if "latency" in metric_key:
        return delta < 0
    return delta > 0


def chart_overhead(data, results_dir):
    metrics = [
        ("throughput_single", "Single-stream throughput"),
        ("throughput_multi", "Multi-stream throughput"),
        ("latency", "Latency (RTT)"),
    ]

    labels = []
    deltas = []
    metric_keys = []
    for d in data:
        name = SCENARIO_LABELS.get(d["scenario"], d["scenario"]).replace("\n", " ")
        for key, metric_label in metrics:
            labels.append(f"{name}  --  {metric_label}")
            deltas.append(d[key]["delta_pct"])
            metric_keys.append(key)

    labels.reverse()
    deltas.reverse()
    metric_keys.reverse()

    fig, ax = plt.subplots(figsize=(11, 7))
    y = np.arange(len(labels))

    bar_colors = []
    for d, mk in zip(deltas, metric_keys):
        if abs(d) < 0.5:
            bar_colors.append(COLORS["neutral"])
        elif _is_improvement(d, mk):
            bar_colors.append(COLORS["positive"])
        else:
            bar_colors.append(COLORS["negative"])

    bars = ax.barh(y, deltas, height=0.6, color=bar_colors,
                   edgecolor=COLORS["bar_edge"], linewidth=0.5, zorder=3)

    ax.axvline(x=0, color="#ADB5BD", linewidth=1.0, zorder=2)

    for bar, delta in zip(bars, deltas):
        sign = "+" if delta >= 0 else ""
        x_pos = bar.get_width()
        offset = 0.3 if delta >= 0 else -0.3
        ha = "left" if delta >= 0 else "right"
        ax.text(x_pos + offset, bar.get_y() + bar.get_height() / 2,
                f"{sign}{delta:.2f}%", va="center", ha=ha,
                fontsize=9, fontweight="bold", color=COLORS["text"])

    ax.set_yticks(y)
    ax.set_yticklabels(labels, fontsize=9)
    ax.set_xlabel("Change vs Baseline (%)")
    ax.set_title("netleak Overhead vs Direct Baseline", pad=14)

    margin = max(abs(min(deltas)), abs(max(deltas))) * 1.5
    ax.set_xlim(-max(margin, 2), max(margin, 2))

    ax.annotate("< negative change",
                xy=(ax.get_xlim()[0] * 0.55, len(labels) - 0.2),
                fontsize=8, fontstyle="italic", color=COLORS["text"],
                ha="center", alpha=0.6)
    ax.annotate("positive change >",
                xy=(ax.get_xlim()[1] * 0.55, len(labels) - 0.2),
                fontsize=8, fontstyle="italic", color=COLORS["text"],
                ha="center", alpha=0.6)

    fig.tight_layout()
    out = Path(results_dir) / "overhead.png"
    fig.savefig(out, dpi=150, bbox_inches="tight")
    plt.close(fig)
    print(f"  {out}")


def main():
    results_dir = sys.argv[1] if len(sys.argv) > 1 else str(
        Path(__file__).resolve().parent.parent / "results"
    )

    if not Path(results_dir, "results.json").exists():
        print(f"Error: {results_dir}/results.json not found", file=sys.stderr)
        print("Run benchmarks and aggregate-results.sh first.", file=sys.stderr)
        sys.exit(1)

    setup_style()
    data = load_results(results_dir)

    print("Generating benchmark charts...")
    chart_throughput_single(data, results_dir)
    chart_throughput_multi(data, results_dir)
    chart_latency(data, results_dir)
    chart_overhead(data, results_dir)
    print("Done.")


if __name__ == "__main__":
    main()
