import os
from bisect import bisect_left

import matplotlib as mpl
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
from matplotlib import gridspec


mpl.rcParams["xtick.direction"] = "in"
mpl.rcParams["ytick.direction"] = "in"
plt.rcParams["font.family"] = "Times New Roman"
plt.rcParams["mathtext.fontset"] = "stix"
plt.rcParams["font.size"] = 20

COLORS = {
    "Nomos": "#2ca02c",
    "MC-ODXT": "#1f77b4",
    "VQNomos": "#d62728",
}

MARKERS = {
    "Nomos": "D",
    "MC-ODXT": "s",
    "VQNomos": "v",
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

    sampled_indices = {0, len(x_array) - 1}
    prev_ratio = 0.0
    first_x = float(x_array[0])

    for end_ratio, point_count in cfg["sample_segments"]:
        start_x = first_x if prev_ratio == 0.0 else x_max * prev_ratio
        end_x = x_max * end_ratio

        if end_x <= start_x or point_count <= 0:
            prev_ratio = end_ratio
            continue

        segment_targets = np.linspace(start_x, end_x, point_count, endpoint=True)
        for target in segment_targets:
            sampled_indices.add(nearest_index(x_array, float(target)))

        prev_ratio = end_ratio

    sampled_indices = sorted(sampled_indices)
    return x_values.iloc[sampled_indices], y_values.iloc[sampled_indices]


def plot_dataset(dataset_name, cfg, y_column, y_label, output_dir, output_suffix):
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

            x_values = df["upd_w1"]
            y_values = df[y_column] / 1000.0
            x_sampled, y_sampled = sample_points(x_values, y_values, cfg)

            line, = ax_main.plot(
                x_sampled,
                y_sampled,
                color=COLORS[scheme],
                linestyle="-",
                label=scheme,
                marker=MARKERS[scheme],
                markersize=10,
                markerfacecolor=COLORS[scheme],
                markeredgecolor=COLORS[scheme],
                markeredgewidth=0.5,
                linewidth=4,
                antialiased=True,
            )

            handles.append(line)
            labels.append(scheme)
        except Exception as exc:
            print(f"处理 {dataset_name} - {scheme} 数据时出错: {exc}")

    ax_main.set_xlabel(r"$|upd(w_{1})|$", fontsize=42, fontweight="bold", labelpad=1)
    ax_main.set_ylabel(y_label, fontsize=42, fontweight="bold", labelpad=1, rotation=90)

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
    output_file = os.path.join(output_dir, f"{cfg['output_name']}_{output_suffix}.pdf")
    plt.savefig(output_file, dpi=3600, format="pdf")
    plt.close(fig)


def plot_all(output_dir, dataset_config, y_column, y_label, output_suffix):
    os.makedirs(output_dir, exist_ok=True)
    for dataset_name, cfg in dataset_config.items():
        plot_dataset(dataset_name, cfg, y_column, y_label, output_dir, output_suffix)
