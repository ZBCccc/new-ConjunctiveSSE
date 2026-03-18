import os
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
import re
from matplotlib import gridspec
from matplotlib.ticker import LogLocator

# 创建输出目录
os.makedirs("pic/nomos_client_communication_w1", exist_ok=True)

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
    "Nomos": "D",    # 菱形
    "MC-ODXT": "s",  # 方形
    "VQNomos": "v",  # 倒三角
}

DATASET_CONFIG = {
    "Crime": {
        "file_paths": {
            "Nomos": "pic/nomos_client_communication_data_w1/Nomos_Crime.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w1/MC-ODXT_Crime.csv",
            "VQNomos": "pic/nomos_client_communication_data_w1/VQNomos_Crime.csv",
        },
        "xlim": 17000,
        "x_ticks": [0, 4000, 8000, 12000, 16000],
        "x_tick_labels": ["0", "4k", "8k", "12k", "16k"],
        "x_minor": 2000,
        "ylim": (1, 40),
        "y_ticks": [2, 5, 10, 20, 40],
        "y_labels": ["2", "5", "10", "20", "40"],
        "legend_xlim": 16644,
        "sample_mode": "crime",
        "raw_count_file": "pic/raw_data/Crime_filecnt_sorted.json",
    },
    "Enron": {
        "file_paths": {
            "Nomos": "pic/nomos_client_communication_data_w1/Nomos_Enron.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w1/MC-ODXT_Enron.csv",
            "VQNomos": "pic/nomos_client_communication_data_w1/VQNomos_Enron.csv",
        },
        "xlim": 28000,
        "x_ticks": [0, 5000, 10000, 15000, 20000, 25000],
        "x_tick_labels": ["0", "5k", "10k", "15k", "20k", "25k"],
        "x_minor": 2500,
        "ylim": (1, 40),
        "y_ticks": [2, 5, 10, 20, 40],
        "y_labels": ["2", "5", "10", "20", "40"],
        "legend_xlim": 26946,
        "sample_mode": "enron",
        "raw_count_file": "pic/raw_data/Enron_filecnt_sorted.json",
    },
    "Wikipedia": {
        "file_paths": {
            "Nomos": "pic/nomos_client_communication_data_w1/Nomos_Wikipedia.csv",
            "MC-ODXT": "pic/nomos_client_communication_data_w1/MC-ODXT_Wikipedia.csv",
            "VQNomos": "pic/nomos_client_communication_data_w1/VQNomos_Wikipedia.csv",
        },
        "xlim": 10000,
        "x_ticks": [0, 2000, 4000, 6000, 8000],
        "x_tick_labels": ["0", "2k", "4k", "6k", "8k"],
        "x_minor": 1000,
        "ylim": (1, 40),
        "y_ticks": [2, 5, 10, 20, 40],
        "y_labels": ["2", "5", "10", "20", "40"],
        "legend_xlim": 9738,
        "sample_mode": "wikipedia",
        "raw_count_file": "pic/raw_data/Wiki_filecnt_sorted.json",
    },
}


def read_doc_counts(json_path):
    pattern = re.compile(r'"[^"]+":\s*(\d+)')
    counts = []
    with open(json_path, "r", encoding="utf-8") as f:
        for line in f:
            m = pattern.search(line)
            if m:
                counts.append(int(m.group(1)))
    return counts


def sample_points(x_values, y_values, mode):
    if mode == "crime":
        x_below = x_values[x_values < 10000]
        y_below = y_values[x_values < 10000]
        x_above = x_values[x_values >= 10000]
        y_above = y_values[x_values >= 10000]

        idx_below = np.arange(1000, len(x_below), 2000)
        idx_above = np.arange(1000, len(x_above), 5000)

        x_sampled_below = x_below.iloc[idx_below]
        y_sampled_below = y_below.iloc[idx_below]
        x_sampled_above = x_above.iloc[idx_above]
        y_sampled_above = y_above.iloc[idx_above]

        last_point = x_above.iloc[[-1]] if len(x_above) > 0 else x_values.iloc[[-1]]
        last_y = y_above.iloc[[-1]] if len(y_above) > 0 else y_values.iloc[[-1]]

        x_sampled = pd.concat([x_sampled_below, x_sampled_above, last_point])
        y_sampled = pd.concat([y_sampled_below, y_sampled_above, last_y])
        return x_sampled, y_sampled

    if mode == "enron":
        x_below = x_values[x_values < 4000]
        y_below = y_values[x_values < 4000]
        x_above = x_values[x_values >= 4000]
        y_above = y_values[x_values >= 4000]

        idx_below = np.arange(500, len(x_below), 500)
        idx_above = np.arange(1000, len(x_above), 1000)

        x_sampled_below = x_below.iloc[idx_below]
        y_sampled_below = y_below.iloc[idx_below]
        x_sampled_above = x_above.iloc[idx_above]
        y_sampled_above = y_above.iloc[idx_above]

        last_point = x_above.iloc[[-1]] if len(x_above) > 0 else x_values.iloc[[-1]]
        last_y = y_above.iloc[[-1]] if len(y_above) > 0 else y_values.iloc[[-1]]

        x_sampled = pd.concat([x_sampled_below, x_sampled_above, last_point])
        y_sampled = pd.concat([y_sampled_below, y_sampled_above, last_y])
        return x_sampled, y_sampled

    # wikipedia
    idx = np.arange(550, len(x_values), 400)
    x_sampled = x_values.iloc[idx]
    y_sampled = y_values.iloc[idx]
    x_sampled = pd.concat([x_sampled, x_values.iloc[[-1]]])
    y_sampled = pd.concat([y_sampled, y_values.iloc[[-1]]])
    return x_sampled, y_sampled


def plot_dataset(dataset_name, cfg):
    fig = plt.figure(figsize=(18, 12), dpi=3600)
    gs = gridspec.GridSpec(2, 1, height_ratios=[1, 8], hspace=0.05)

    ax_legend = plt.subplot(gs[0])
    ax_legend.axis("off")
    ax_legend.set_xlim(0, cfg["legend_xlim"])

    ax_main = plt.subplot(gs[1])

    handles = []
    labels = []
    doc_counts = read_doc_counts(cfg["raw_count_file"])

    for scheme, file_path in cfg["file_paths"].items():
        try:
            df = pd.read_csv(file_path)
            df.columns = df.columns.str.strip()

            x_values = df["KeywordCount"]
            # 原始单位是 Byte，这里只做 Byte -> KB，不做 bit 转换
            y_values = df["Storage(Bytes)"] / 1024.0

            x_sampled, y_sampled = sample_points(x_values, y_values, cfg["sample_mode"])

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
    x_tick_labels = [str(int(x)) for x in cfg["x_ticks"]]
    ax_main.set_xticklabels(x_tick_labels, fontsize=50)

    plt.yscale("log")
    plt.yticks(cfg["y_ticks"], cfg["y_labels"], fontsize=50)
    plt.ylim(cfg["ylim"][0], cfg["ylim"][1])

    ax_main.tick_params(axis="both", which="both", direction="in")
    ax_main.tick_params(axis="x", which="major", bottom=True, top=True, length=8, width=3.5)
    ax_main.tick_params(axis="y", which="major", left=True, right=True, length=8, width=3.5)
    ax_main.tick_params(axis="x", which="minor", bottom=True, top=True, length=4, width=2)
    ax_main.tick_params(axis="y", which="minor", left=True, right=True, length=4, width=2)

    ax_main.xaxis.set_minor_locator(plt.MultipleLocator(cfg["x_minor"]))
    ax_main.yaxis.set_minor_locator(LogLocator(base=10.0, subs="auto", numticks=100))

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
    output_file = f"pic/nomos_client_communication_w1/{dataset_name}_client_communication_comparison.pdf"
    plt.savefig(output_file, dpi=3600, format="pdf")
    plt.close(fig)


for dataset_name, cfg in DATASET_CONFIG.items():
    plot_dataset(dataset_name, cfg)
