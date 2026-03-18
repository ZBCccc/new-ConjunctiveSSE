import os
from bisect import bisect_left
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
from matplotlib import gridspec

# 创建输出目录
os.makedirs("pic/nomos_client_communication_w2", exist_ok=True)

# 设置全局刻度线朝向内侧
mpl.rcParams["xtick.direction"] = "in"
mpl.rcParams["ytick.direction"] = "in"
# 设置字体为Times New Roman
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
            "Nomos": "pic/nomos_client_communication_data_w2/Nomos_Crime.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w2/MC-ODXT_Crime.csv",
            "VQNomos": "pic/nomos_client_communication_data_w2/VQNomos_Crime.csv",
        },
        "xlim": 17000,
        "x_ticks": [0, 4000, 8000, 12000, 16000],
        "x_minor": 2000,
        "y_ticks": [0, 500, 1000, 1500, 2000, 2500],
        "y_labels": ["", "500", "1000", "1500", "2000", "2500"],
        "y_lim": (0, 2500),
        "y_minor": 250,
        "legend_xlim": 16644,
        "sample_segments": [(0.15, 12), (0.50, 10), (1.00, 8)],
    },
    "Enron": {
        "file_paths": {
            "Nomos": "pic/nomos_client_communication_data_w2/Nomos_Enron.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w2/MC-ODXT_Enron.csv",
            "VQNomos": "pic/nomos_client_communication_data_w2/VQNomos_Enron.csv",
        },
        "xlim": 28000,
        "x_ticks": [0, 5000, 10000, 15000, 20000, 25000],
        "x_minor": 2500,
        "y_ticks": [0, 700, 1400, 2100, 2800, 3500],
        "y_labels": ["", "700", "1400", "2100", "2800", "3500"],
        "y_lim": (0, 3500),
        "y_minor": 350,
        "legend_xlim": 26946,
        "sample_segments": [(0.12, 12), (0.45, 10), (1.00, 8)],
    },
    "Wikipedia": {
        "file_paths": {
            "Nomos": "pic/nomos_client_communication_data_w2/Nomos_Wikipedia.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w2/MC-ODXT_Wikipedia.csv",
            "VQNomos": "pic/nomos_client_communication_data_w2/VQNomos_Wikipedia.csv",
        },
        "xlim": 10000,
        "x_ticks": [0, 2000, 4000, 6000, 8000],
        "x_minor": 1000,
        "y_ticks": [0, 300, 600, 900, 1200, 1500],
        "y_labels": ["", "300", "600", "900", "1200", "1500"],
        "y_lim": (0, 1500),
        "y_minor": 150,
        "legend_xlim": 9738,
        "sample_segments": [(0.20, 12), (0.60, 10), (1.00, 8)],
    },
}

Y_TICKS = [0, 500, 1000, 1500, 2000, 2500, 3000, 3500]
Y_LABELS = ["0", "500", "1000", "1500", "2000", "2500", "3000", "3500"]
Y_LIM = (0, 3500)


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


def plot_dataset(dataset_name, cfg):
    fig = plt.figure(figsize=(18, 12), dpi=3600)
    gs = gridspec.GridSpec(2, 1, height_ratios=[1, 8], hspace=0.05)

    ax_legend = plt.subplot(gs[0])
    ax_legend.axis("off")
    ax_legend.set_xlim(0, cfg["legend_xlim"])

    ax_main = plt.subplot(gs[1])
    y_ticks = cfg.get("y_ticks", Y_TICKS)
    y_labels = cfg.get("y_labels", Y_LABELS)
    y_lim = cfg.get("y_lim", Y_LIM)
    y_minor = cfg.get("y_minor", 250)

    handles = []
    labels = []

    for scheme, file_path in cfg["file_paths"].items():
        try:
            df = pd.read_csv(file_path)
            df.columns = df.columns.str.strip()

            x_values = df["KeywordCount"]
            y_values = df["Storage(Bytes)"] / 1024.0

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
        except Exception as e:
            print(f"处理 {dataset_name} - {scheme} 数据时出错: {e}")

    ax_main.set_xlabel(r"$|upd(w_{2})|$", fontsize=42, fontweight="bold", labelpad=1)
    ax_main.set_ylabel("Communication Cost (KB)", fontsize=42, fontweight="bold", labelpad=1, rotation=90)

    ax_main.set_xlim(0, cfg["xlim"])
    ax_main.set_xticks(cfg["x_ticks"])
    ax_main.set_xticklabels([str(int(x)) for x in cfg["x_ticks"]], fontsize=50)

    plt.yscale("linear")
    plt.yticks(y_ticks, y_labels, fontsize=50)
    plt.ylim(y_lim[0], y_lim[1])

    ax_main.tick_params(axis="both", which="both", direction="in")
    ax_main.tick_params(axis="x", which="major", bottom=True, top=True, length=8, width=3.5)
    ax_main.tick_params(axis="y", which="major", left=True, right=True, length=8, width=3.5)
    ax_main.tick_params(axis="x", which="minor", bottom=True, top=True, length=4, width=2)
    ax_main.tick_params(axis="y", which="minor", left=True, right=True, length=4, width=2)

    ax_main.xaxis.set_minor_locator(plt.MultipleLocator(cfg["x_minor"]))
    ax_main.yaxis.set_minor_locator(plt.MultipleLocator(y_minor))

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
    output_file = f"pic/nomos_client_communication_w2/{dataset_name}_client_communication_comparison.pdf"
    plt.savefig(output_file, dpi=3600, format="pdf")
    plt.close(fig)


for dataset_name, cfg in DATASET_CONFIG.items():
    plot_dataset(dataset_name, cfg)
