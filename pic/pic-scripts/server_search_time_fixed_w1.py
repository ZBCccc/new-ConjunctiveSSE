import os
from bisect import bisect_left

import matplotlib as mpl
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from matplotlib import gridspec


OUTPUT_DIR = "pic/server_search_time_fixed_w1_figures"
os.makedirs(OUTPUT_DIR, exist_ok=True)

mpl.rcParams["xtick.direction"] = "in"
mpl.rcParams["ytick.direction"] = "in"
plt.rcParams["font.family"] = "Times New Roman"
plt.rcParams["mathtext.fontset"] = "stix"
plt.rcParams["font.size"] = 20

colors = {
    "Nomos": "#2ca02c",
    "MC-ODXT": "#1f77b4",
    "VQNomos": "#d62728",
}

markers = {
    "Nomos": "D",
    "MC-ODXT": "s",
    "VQNomos": "v",
}

DATASET_CONFIG = {
    "Crime": {
        "file_paths": {
            "Nomos": "pic/server_search_time_fixed_w1/Nomos_Crime.csv",
            "MC-ODXT": "pic/server_search_time_fixed_w1/MC-ODXT_Crime.csv",
            "VQNomos": "pic/server_search_time_fixed_w1/VQNomos_Crime.csv",
        },
        "xlim": 17000,
        "x_ticks": [0, 4000, 8000, 12000, 16000],
        "x_minor": 2000,
        "ylim": (0, 4),
        "y_ticks": [0, 1, 2, 3, 4],
        "y_labels": ["", "1", "2", "3", "4"],
        "y_minor": 0.5,
        "legend_xlim": 16644,
        "include_first_point": False,
        "min_sample_x": 10,
        "sample_segments": [(0.12, 12), (0.50, 10), (1.00, 8)],
        "output_name": "Crime",
    },
    "Enron": {
        "file_paths": {
            "Nomos": "pic/server_search_time_fixed_w1/Nomos_Enron.csv",
            "MC-ODXT": "pic/server_search_time_fixed_w1/MC-ODXT_Enron.csv",
            "VQNomos": "pic/server_search_time_fixed_w1/VQNomos_Enron.csv",
        },
        "xlim": 28000,
        "x_ticks": [0, 5000, 10000, 15000, 20000, 25000],
        "x_minor": 2500,
        "ylim": (0, 4),
        "y_ticks": [0, 1, 2, 3, 4],
        "y_labels": ["", "1", "2", "3", "4"],
        "y_minor": 0.5,
        "legend_xlim": 26946,
        "include_first_point": True,
        "min_sample_x": 10,
        "sample_segments": [(0.12, 12), (0.45, 10), (1.00, 8)],
        "output_name": "Enron",
    },
    "Wikipedia": {
        "file_paths": {
            "Nomos": "pic/server_search_time_fixed_w1/Nomos_Wiki.csv",
            "MC-ODXT": "pic/server_search_time_fixed_w1/MC-ODXT_Wiki.csv",
            "VQNomos": "pic/server_search_time_fixed_w1/VQNomos_Wiki.csv",
        },
        "xlim": 10000,
        "x_ticks": [0, 2000, 4000, 6000, 8000],
        "x_minor": 1000,
        "ylim": (0, 4),
        "y_ticks": [0, 1, 2, 3, 4],
        "y_labels": ["", "1", "2", "3", "4"],
        "y_minor": 0.5,
        "legend_xlim": 9738,
        "include_first_point": False,
        "min_sample_x": 10,
        "sample_segments": [(0.12, 12), (0.60, 10), (1.00, 8)],
        "output_name": "Wikipedia",
    },
}


def nearest_index(x_array, target):
    pos = bisect_left(x_array, target)
    candidates = []

    if pos < len(x_array):
        candidates.append(pos)
    if pos > 0:
        candidates.append(pos - 1)

    return min(candidates, key=lambda idx: (abs(x_array[idx] - target), idx))


def sample_points(x_values, y_values, cfg):
    x_array = x_values.to_numpy(dtype=float)
    x_max = min(float(cfg["xlim"]), float(x_array[-1]))
    min_sample_x = float(cfg.get("min_sample_x", x_array[0]))
    valid_start_idx = bisect_left(x_array, min_sample_x)

    sampled_indices = {len(x_array) - 1}
    if cfg.get("include_first_point", True) and valid_start_idx < len(x_array):
        sampled_indices.add(valid_start_idx)
    prev_ratio = 0.0
    first_x = float(x_array[valid_start_idx]) if valid_start_idx < len(x_array) else float(x_array[0])

    for end_ratio, point_count in cfg["sample_segments"]:
        start_x = first_x if prev_ratio == 0.0 else x_max * prev_ratio
        end_x = x_max * end_ratio

        if end_x <= start_x or point_count <= 0:
            prev_ratio = end_ratio
            continue

        segment_targets = np.linspace(start_x, end_x, point_count, endpoint=True)
        for target in segment_targets:
            sampled_indices.add(nearest_index(x_array[valid_start_idx:], float(target)) + valid_start_idx)

        prev_ratio = end_ratio

    sampled_indices = sorted(sampled_indices)
    return x_values.iloc[sampled_indices], y_values.iloc[sampled_indices]


def plot_dataset(dataset_name, cfg):
    fig = plt.figure(figsize=(18, 12), dpi=3600)
    gs = gridspec.GridSpec(2, 1, height_ratios=[1, 8], hspace=0.05)

    ax_legend = plt.subplot(gs[0])
    ax_legend.axis("off")
    ax_legend.set_xlim(0, cfg["legend_xlim"])

    ax_main = plt.subplot(gs[1])
    handles = []
    labels = []

    for scheme, file_path in cfg["file_paths"].items():
        try:
            df = pd.read_csv(file_path)
            df.columns = df.columns.str.strip()

            x_values = df["upd_w2"]
            y_values = df["server_time_ms"]
            x_sampled, y_sampled = sample_points(x_values, y_values, cfg)

            line, = ax_main.plot(
                x_sampled,
                y_sampled,
                color=colors[scheme],
                linestyle="-",
                label=scheme,
                marker=markers[scheme],
                markersize=10,
                markerfacecolor=colors[scheme],
                markeredgecolor=colors[scheme],
                markeredgewidth=0.5,
                linewidth=4,
                antialiased=True,
            )

            handles.append(line)
            labels.append(scheme)
        except Exception as exc:
            print(f"处理 {dataset_name} - {scheme} 数据时出错: {exc}")

    ax_main.set_xlabel(r"$|upd(w_{2})|$", fontsize=42, fontweight="bold", labelpad=1)
    ax_main.set_ylabel("Server Search Time (ms)", fontsize=42, fontweight="bold", labelpad=1, rotation=90)

    ax_main.set_xlim(0, cfg["xlim"])
    ax_main.set_xticks(cfg["x_ticks"])
    ax_main.set_xticklabels([str(int(x)) for x in cfg["x_ticks"]], fontsize=50)

    plt.yscale("linear")
    plt.yticks(cfg["y_ticks"], cfg["y_labels"], fontsize=50)
    plt.ylim(cfg["ylim"][0], cfg["ylim"][1])

    ax_main.tick_params(axis="both", which="both", direction="in")
    ax_main.tick_params(axis="x", which="major", bottom=True, top=True, length=8, width=3.5)
    ax_main.tick_params(axis="y", which="major", left=True, right=True, length=8, width=3.5)
    ax_main.tick_params(axis="x", which="minor", bottom=True, top=True, length=4, width=2)
    ax_main.tick_params(axis="y", which="minor", left=True, right=True, length=4, width=2)

    ax_main.xaxis.set_minor_locator(plt.MultipleLocator(cfg["x_minor"]))
    ax_main.yaxis.set_minor_locator(plt.MultipleLocator(cfg["y_minor"]))

    leg = ax_legend.legend(
        handles,
        labels,
        loc="center",
        ncol=3,
        frameon=True,
        fontsize=45,
        handlelength=3.6,
        handletextpad=0.2,
        borderpad=0.2,
        columnspacing=0.5,
        markerscale=1.5,
        bbox_to_anchor=(0.5, 0.5),
        bbox_transform=ax_legend.transAxes,
    )

    frame = leg.get_frame()
    frame.set_linewidth(1)
    frame.set_edgecolor("black")

    plt.subplots_adjust(left=0.12, right=0.97, top=0.98, bottom=0.12)
    output_file = os.path.join(
        OUTPUT_DIR,
        f"{cfg['output_name']}_server_search_time_comparison.pdf",
    )
    plt.savefig(output_file, dpi=3600, format="pdf")
    plt.close(fig)


for dataset_name, cfg in DATASET_CONFIG.items():
    plot_dataset(dataset_name, cfg)
