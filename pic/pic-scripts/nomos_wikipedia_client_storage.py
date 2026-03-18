import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
from matplotlib import gridspec
from matplotlib.ticker import FixedLocator
import os

# 创建输出目录
os.makedirs("pic/nomos_client_storage", exist_ok=True)

# 设置全局刻度线朝向内侧
mpl.rcParams['xtick.direction'] = 'in'
mpl.rcParams['ytick.direction'] = 'in'
# 设置字体为Times New Roman
plt.rcParams['font.family'] = 'Times New Roman'
plt.rcParams['mathtext.fontset'] = 'stix'
plt.rcParams['font.size'] = 20

# 预设文件地址
file_paths = {
    "Nomos": "pic/nomos_client_data/Nomos_Wikipedia.csv",
    "MC-ODXT": "pic/nomos_client_data/MC-ODXT_Wikipedia.csv",
    "VQNomos": "pic/nomos_client_data/VQNomos_Wikipedia.csv"
}

# 设置不同文件的颜色方案
colors = {
    "Nomos": '#2ca02c',
    "MC-ODXT": '#1f77b4',
    "VQNomos": '#d62728'
}

# 设置不同的标记形状
markers = {
    "Nomos": 'D',      # 菱形
    "MC-ODXT": 's',      # 方形
    "VQNomos": 'v'   # 倒三角
}

fig = plt.figure(figsize=(18, 12), dpi=3600)

# 创建网格布局 - 顶部小区域用于图例，底部大区域用于主图
gs = gridspec.GridSpec(2, 1, height_ratios=[1, 8], hspace=0.05)

# 创建顶部的子图用于放置图例
ax_legend = plt.subplot(gs[0])
ax_legend.axis('off')
ax_legend.set_xlim(0, 11000)

# 创建底部的子图用于主图
ax_main = plt.subplot(gs[1])

# 用于存储图例句柄和标签
handles = []
labels = []

# 处理每个方案的数据并在主图中绘图
for scheme, file_path in file_paths.items():
    try:
        # 读取CSV文件
        df = pd.read_csv(file_path)
        df.columns = df.columns.str.strip()

        # 使用KeywordCount列作为x轴
        x_values = df['KeywordCount']

        # 转换为KB
        y_values = (df['Storage(Bits)']) / 8.0 / 1024.0

        # 选择每隔4一个点进行采样
        indices = np.arange(200, len(x_values), 400)
        x_sampled = x_values.iloc[indices]
        y_sampled = y_values.iloc[indices]

        # 确保最后一个点被包含
        x_sampled = pd.concat([x_sampled, x_values.iloc[[-1]]])
        y_sampled = pd.concat([y_sampled, y_values.iloc[[-1]]])

        # 绘制主图线并获取线条对象
        line, = ax_main.plot(x_sampled, y_sampled,
                 color=colors[scheme],
                 linestyle='-',
                 label=scheme,
                 marker=markers[scheme],
                 markersize=10,
                 markerfacecolor=colors[scheme],
                 markeredgecolor=colors[scheme],
                 markeredgewidth=0.5,
                 linewidth=4,
                 antialiased=True)

        # 添加到句柄和标签列表
        handles.append(line)
        labels.append(scheme)
    except Exception as e:
        print(f"处理 {scheme} 数据时出错: {e}")

# 设置主图属性
ax_main.set_xlabel('Keywords Number', fontsize=42, fontweight='bold', labelpad=1)
ax_main.set_ylabel('Client Storage Cost (KB)', fontsize=42, fontweight='bold', labelpad=1, rotation=90)

# 设置主图的x轴范围和刻度
ax_main.set_xlim(0, 10500)
x_ticks = [0, 2000, 4000, 6000, 8000, 10000]
ax_main.set_xticks(x_ticks)
ax_main.set_xticklabels([f'{int(x/1000)}k' if x != 0 else '0' for x in x_ticks], fontsize=50)

# 设置y轴为正常线性刻度
y_ticks = [30, 60, 90, 120]
y_labels = [str(int(y)) for y in y_ticks]
plt.yticks(y_ticks, y_labels, fontsize=50)
plt.ylim(0, 120)

# 强制设置刻度线朝向内侧
ax_main.tick_params(axis='both', which='both', direction='in')
ax_main.tick_params(axis='x', which='major', bottom=True, top=True, length=8, width=3.5)
ax_main.tick_params(axis='y', which='major', left=True, right=True, length=8, width=3.5)
ax_main.tick_params(axis='x', which='minor', bottom=True, top=True, length=4, width=2)
ax_main.tick_params(axis='y', which='minor', left=True, right=True, length=4, width=2)

# 设置次要刻度线
ax_main.xaxis.set_minor_locator(plt.MultipleLocator(1000))
# y轴次刻度线：每两个主刻度线之间放一个次刻度线
ax_main.yaxis.set_minor_locator(FixedLocator([15, 45, 75, 105]))

# 添加图例到上面的子图
leg = ax_legend.legend(handles, labels, loc='center', ncol=3, frameon=True,
                      fontsize=45, handlelength=3.6, handletextpad=0.2, borderpad=0.2,
                      columnspacing=0.5, markerscale=1.5,
                      bbox_to_anchor=(0.5, 0.5),
                      bbox_transform=ax_legend.transAxes)

# 给图例添加边框
frame = leg.get_frame()
frame.set_linewidth(1)
frame.set_edgecolor('black')

# 调整子图边距
plt.subplots_adjust(left=0.12, right=0.97, top=0.98, bottom=0.12)

# 保存文件
plt.savefig("pic/nomos_client_storage/Wikipedia_client_storage_comparison.pdf", dpi=3600, format='pdf')
