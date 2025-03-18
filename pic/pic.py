import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
from matplotlib.ticker import LogLocator

# 设置全局刻度线朝向内侧
mpl.rcParams['xtick.direction'] = 'in'
mpl.rcParams['ytick.direction'] = 'in'

# 预设文件地址（实际使用时替换）
file_paths = {
    "ODXT": "result/Search/ODXT/Crime_USENIX_REV/processed_2025-03-15_20-43-23.csv",
    # "HDXT": "result/Search/HDXT/Crime_USENIX_REV/2025-03-06_17-54-01.csv",
    "FDXT": "result/Search/FDXT/Crime_USENIX_REV/processed_2025-03-15_17-08-07.csv",
    "SDSSE-CQ": "result/Search/SDSSE-CQ/Crime_USENIX_REV/2025-03-15_18-15-39.csv"
}

# 设置不同文件的颜色方案
colors = {
    "ODXT": '#1f77b4',
    # "HDXT": '#ff7f0e',
    "FDXT": '#2ca02c',
    "SDSSE-CQ": '#d62728'
}

# 设置不同的标记形状
markers = {
    "ODXT": 's',      # 方形
    # "HDXT": 'D',    # 菱形
    "FDXT": 'D',      # 菱形
    "SDSSE-CQ": 'v'   # 倒三角
}

# 创建高分辨率图表
plt.figure(figsize=(12, 8), dpi=1200)

# 用于存储图例句柄和标签
handles = []
labels = []

# 处理每个方案的数据并绘图
for scheme, file_path in file_paths.items():
    try:
        # 读取CSV文件
        df = pd.read_csv(file_path)
        df.columns = df.columns.str.strip()  # 去掉列名前后空格
        
        # 使用w2列作为x轴（关键词数量）
        x_values = df['w2']
        
        # 使用totalTime列作为y轴（查询时间，单位为毫秒）
        # 原始数据是微秒，需要转换为毫秒
        y_values = df['totalTime'] / 1000.0  # 微秒转毫秒
        
        # 绘制主图线并获取线条对象
        line, = plt.plot(x_values, y_values, 
                 color=colors[scheme], 
                 linestyle='-', 
                 label=scheme, 
                 marker=markers[scheme],
                 markersize=3,
                 markerfacecolor=colors[scheme],
                 markeredgecolor='black',  # 使用相同颜色，去掉黑色边框
                 markeredgewidth=0.2,  # 设置边框宽度为0
                 linewidth=1.5,
                 antialiased=True)  # 抗锯齿
        
        # 添加到句柄和标签列表
        handles.append(line)
        labels.append(scheme)
    except Exception as e:
        print(f"处理 {scheme} 数据时出错: {e}")

# 获取当前坐标轴
ax = plt.gca()

# 设置主图属性
plt.xlabel('Keyword Volume', fontsize=14, fontweight='bold')
plt.ylabel('Time cost(ms)', fontsize=14, fontweight='bold')
plt.grid(False)  # 不显示网格线
# plt.title('Performance Comparison', fontsize=16, fontweight='bold')

# 设置主图的x轴范围和刻度
plt.xlim(0, 16000)
x_ticks = [0, 4000, 8000, 12000, 16000]
plt.xticks(x_ticks, fontsize=12)

# 设置y轴为对数刻度
plt.yscale('log')

# 设置y轴刻度范围
y_ticks = [10**i for i in range(-1, 5)]  # 从10^-1到10^4
plt.yticks(y_ticks, [f'$10^{{{i}}}$' for i in range(-1, 5)], fontsize=12)

# 强制设置刻度线朝向内侧
ax.tick_params(axis='both', which='both', direction='in')

# 设置主要刻度线（在每个数上）
ax.tick_params(axis='x', which='major', bottom=True, top=False, length=6, width=1)
ax.tick_params(axis='y', which='major', left=True, right=True, length=6, width=1)

# 设置次要刻度线（在数与数之间）
ax.xaxis.set_minor_locator(plt.MultipleLocator(2000))  # 在主刻度之间添加次要刻度
ax.tick_params(axis='x', which='minor', bottom=True, top=False, length=3, width=0.5)

# 对于对数刻度的y轴，次要刻度已经自动设置，只需调整样式
ax.yaxis.set_minor_locator(LogLocator(base=10.0, subs=[2.0, 5.0]))
ax.tick_params(axis='y', which='minor', left=True, right=True, length=3, width=0.5)

# 使用收集的句柄和标签添加图例到主图右上角
plt.legend(handles, labels, loc='upper right', fontsize=10, frameon=True)

# 创建插图（放大部分区域）
ax_inset = plt.axes([0.5, 0.24, 0.35, 0.25])  # [x, y, width, height]

# 在插图中绘制相同的数据，但只显示特定范围
for scheme, file_path in file_paths.items():
    try:
        df = pd.read_csv(file_path)
        df.columns = df.columns.str.strip()
        
        x_values = df['w2']
        # 原始数据是微秒，需要转换为毫秒
        y_values = df['totalTime'] / 1000.0  # 微秒转毫秒
        
        # 在插图中绘制
        ax_inset.plot(x_values, y_values, 
                     color=colors[scheme], 
                     linestyle='-', 
                     marker=markers[scheme],
                     markersize=3,
                     markerfacecolor=colors[scheme],
                     markeredgecolor='black',  # 使用相同颜色，去掉黑色边框
                     markeredgewidth=0.2,  # 设置边框宽度为0
                     linewidth=1.0,
                     antialiased=True)  # 抗锯齿
    except Exception as e:
        print(f"处理插图中的 {scheme} 数据时出错: {e}")

# 设置插图的范围
ax_inset.set_xlim(1000, 4500)  # 小图x轴范围为1000到5000
ax_inset.set_ylim(0.06, 0.2)   # y轴范围设置为0.06到0.2ms
ax_inset.grid(False)  # 不显示网格线

# 强制设置插图刻度线朝向内侧
ax_inset.tick_params(axis='both', which='both', direction='in')

# 设置插图主要刻度线
ax_inset.tick_params(axis='x', which='major', bottom=True, top=False, length=4, width=0.5, labelsize=10)
ax_inset.tick_params(axis='y', which='major', left=True, right=True, length=4, width=0.5, labelsize=10)

# 设置插图次要刻度线
ax_inset.xaxis.set_minor_locator(plt.MultipleLocator(1000))  # 在主刻度之间添加次要刻度
ax_inset.tick_params(axis='x', which='minor', bottom=True, top=False, length=2, width=0.5)
ax_inset.tick_params(axis='y', which='minor', left=True, right=True, length=2, width=0.5)

# 为插图添加y轴标签，明确单位
ax_inset.set_ylabel('ms', fontsize=8)

# 保存为矢量格式，提高清晰度
plt.savefig("pic/result/performance_comparison.pdf", dpi=3600, bbox_inches='tight', format='pdf')
plt.savefig("pic/result/performance_comparison.svg", dpi=3600, bbox_inches='tight', format='svg')
# 同时保存PNG格式，用于兼容性
plt.savefig("pic/result/performance_comparison.png", dpi=3600, bbox_inches='tight', format='png')

plt.show()
