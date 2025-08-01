from matplotlib import ticker
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
from matplotlib import gridspec
from matplotlib.patches import ConnectionPatch

# 设置全局刻度线朝向内侧
mpl.rcParams['xtick.direction'] = 'in'
mpl.rcParams['ytick.direction'] = 'in'
# 设置字体为Times New Roman
plt.rcParams['font.family'] = 'Times New Roman'
plt.rcParams['mathtext.fontset'] = 'stix'  # 使数学文本也使用相似的字体
plt.rcParams['font.size'] = 20  # 增加基础字体大小

# 预设文件地址（实际使用时替换）
file_paths = {
    "FDXT": "result/Search/FDXT/Crime_USENIX_REV/2025-03-15_17-07-57.csv",
    "ODXT": "result/Search/ODXT/Crime_USENIX_REV/2025-03-19_10-36-32.csv",
    "SDSSE-CQ": "result/Search/SDSSE-CQ/Crime_USENIX_REV/processed_2025-03-15_18-15-39.csv"
}

# 设置不同文件的颜色方案
colors = {
    "FDXT": '#2ca02c',
    "ODXT": '#1f77b4',
    "SDSSE-CQ": '#d62728'
}

# 设置不同的标记形状
markers = {
    "FDXT": 'D',      # 菱形
    "ODXT": 's',      # 方形
    "SDSSE-CQ": 'v'   # 倒三角
}

# 设置固定的时间值（微秒）
fixed_times = {
    "FDXT": 65,      # 65微秒
    "ODXT": 60,      # 60微秒
    "SDSSE-CQ": 800000  # 800000微秒
}

fig = plt.figure(figsize=(10, 10), dpi=3600)  # 调整图形高度以适应三个子图

# 创建网格布局 - 顶部小区域用于图例，中部和底部用于断轴图
gs = gridspec.GridSpec(3, 1, height_ratios=[1, 5, 3], hspace=0.1)

# 创建顶部的子图用于放置图例
ax_legend = plt.subplot(gs[0])
ax_legend.axis('off')  # 关闭坐标轴显示
# 设置图例子图的范围与主图一致
ax_legend.set_xlim(0, 17000)

# 创建上部子图用于SDSSE-CQ (1000-1300ms)
ax_top = plt.subplot(gs[1])
# 创建下部子图用于FDXT和ODXT (0-30ms)
ax_bottom = plt.subplot(gs[2])

# 用于存储图例句柄和标签
handles = []
labels = []

# 处理每个方案的数据并在相应子图中绘图
for scheme, file_path in file_paths.items():
    try:
        # 读取CSV文件
        df = pd.read_csv(file_path)
        df.columns = df.columns.str.strip()  # 去掉列名前后空格
        
        # 使用w2列作为x轴（关键词数量）
        x_values = df['w2']
        
        # 使用serverTime列作为y轴（毫秒）
        y_values = df['serverTime'] / 1000.0
        
        # 根据方案不同，选择不同的子图进行绘制
        if scheme == "SDSSE-CQ":
            ax_plot = ax_top
        else:
            ax_plot = ax_bottom
            
        # 绘制线条并获取线条对象
        line, = ax_plot.plot(x_values, y_values, 
                 color=colors[scheme], 
                 linestyle='-', 
                 label=scheme, 
                 marker=markers[scheme],
                 markersize=8,  
                 markerfacecolor=colors[scheme],
                 markeredgecolor=colors[scheme],
                 markeredgewidth=0.5,
                 linewidth=2,  
                 antialiased=True)
        
        # 添加到句柄和标签列表
        handles.append(line)
        labels.append(scheme)
    except Exception as e:
        print(f"处理 {scheme} 数据时出错: {e}")

# 设置x轴范围和刻度（两个子图共用的x轴设置）
x_ticks = [0, 4000, 8000, 12000, 16000]
for ax in [ax_top, ax_bottom]:
    ax.set_xlim(0, 17000)
    ax.set_xticks(x_ticks)
    ax.tick_params(axis='x', which='major', labelsize=30)
    
    # 设置次要刻度线
    ax.xaxis.set_minor_locator(plt.MultipleLocator(2000))
    
    # 强制设置刻度线朝向内侧
    ax.tick_params(axis='both', which='both', direction='in')
    ax.tick_params(axis='x', which='major', bottom=True, top=True, length=8, width=1.5)
    ax.tick_params(axis='y', which='major', left=True, right=True, length=8, width=1.5)
    ax.tick_params(axis='x', which='minor', bottom=True, top=True, length=4, width=1)
    ax.tick_params(axis='y', which='minor', left=True, right=True, length=4, width=1)

# 设置上部子图的y轴范围和刻度 (SDSSE-CQ)
ax_top.set_ylim(800, 1400)
ax_top.set_yticks([800, 900, 1000, 1100, 1200, 1300, 1400])
ax_top.set_yticklabels(['$800$', '$900$', '$1000$', '$1100$', '$1200$', '$1300$', '$1400$'], fontsize=30)
ax_top.yaxis.set_minor_locator(plt.MultipleLocator(50))  # 每50ms一个次要刻度

# 设置下部子图的y轴范围和刻度 (FDXT & ODXT)
ax_bottom.set_ylim(0, 30)
ax_bottom.set_yticks([0, 10, 20, 30])
ax_bottom.set_yticklabels(['$0$', '$10$', '$20$', '$30$'], fontsize=30)
ax_bottom.yaxis.set_minor_locator(plt.MultipleLocator(5))  # 每5ms一个次要刻度

# 隐藏上部子图的x轴标签
ax_top.set_xticklabels([])

# 设置下部子图的x轴标签
ax_bottom.set_xlabel('|Upd($w_2$)|', fontsize=42, fontweight='bold', labelpad=10)
ax_bottom.set_xticklabels([f'{x}' for x in x_ticks], fontsize=30)

# 设置y轴标签在图的左侧中央位置
fig.text(0.02, 0.5, 'Server Computation Time(ms)', fontsize=42, fontweight='bold', 
         va='center', rotation='vertical')

# 添加断轴标记
d = 0.01  # 断轴标记的大小
kwargs = dict(transform=ax_top.transAxes, color='k', clip_on=False)
ax_top.plot((-d, +d), (-d, +d), **kwargs)        # 左下角
ax_top.plot((1-d, 1+d), (-d, +d), **kwargs)      # 右下角

kwargs.update(transform=ax_bottom.transAxes)
ax_bottom.plot((-d, +d), (1-d, 1+d), **kwargs)   # 左上角
ax_bottom.plot((1-d, 1+d), (1-d, 1+d), **kwargs) # 右上角

# 添加图例到顶部的子图
leg = ax_legend.legend(handles, labels, loc='center', ncol=3, frameon=True, 
                      fontsize=30, handlelength=2, handletextpad=0.2, borderpad=0.2, 
                      columnspacing=0.5, markerscale=1.5, 
                      bbox_to_anchor=(0.5, 0.5), 
                      bbox_transform=ax_legend.transAxes)

# 给图例添加边框
frame = leg.get_frame()
frame.set_linewidth(1)  # 设置边框线宽为1
frame.set_edgecolor('black')  # 设置边框颜色

# 调整子图边距，确保所有元素对齐
plt.subplots_adjust(left=0.15, right=0.9, top=0.95, bottom=0.1)

# 保存文件
plt.savefig("pic/w1-server/Crime_server_computation_time_comparison_broken_axis.pdf", dpi=3600, bbox_inches='tight', format='pdf')
# plt.savefig("pic/w1-server/Crime_server_computation_time_comparison_broken_axis.svg", dpi=3600, bbox_inches='tight', format='svg')
# plt.savefig("pic/w1-server/Crime_server_computation_time_comparison_broken_axis.png", dpi=3600, bbox_inches='tight', format='png')

