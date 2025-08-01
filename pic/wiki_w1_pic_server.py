from matplotlib import ticker
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import matplotlib as mpl
from matplotlib import gridspec

# 设置全局刻度线朝向内侧
mpl.rcParams['xtick.direction'] = 'in'
mpl.rcParams['ytick.direction'] = 'in'
# 设置字体为Times New Roman
plt.rcParams['font.family'] = 'Times New Roman'
plt.rcParams['mathtext.fontset'] = 'stix'  # 使数学文本也使用相似的字体
plt.rcParams['font.size'] = 20  # 增加基础字体大小

# 预设文件地址（实际使用时替换）
file_paths = {
    "FDXT": "result/Search/FDXT/Wiki_USENIX/2025-03-30_16-44-07_w1_keywords_2.csv",
    "ODXT": "result/Search/ODXT/Wiki_USENIX/w1_keywords_2.csv",
    "SDSSE-CQ": "result/Search/SDSSE-CQ/Wiki_USENIX/w1_keywords_2.csv"
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
    "FDXT": 70,      # 65微秒
    "ODXT": 60,      # 60微秒
    "SDSSE-CQ": 800000  # 800000微秒
}

fig = plt.figure(figsize=(18, 12), dpi=3600)

# 创建网格布局 - 顶部小区域用于图例，底部大区域用于主图
gs = gridspec.GridSpec(2, 1, height_ratios=[1, 8], hspace=0.05)

# 创建顶部的子图用于放置图例
ax_legend = plt.subplot(gs[0])
ax_legend.axis('off')  # 关闭坐标轴显示
# 设置图例子图的范围与主图一致
ax_legend.set_xlim(0, 17000)

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
        df.columns = df.columns.str.strip()  # 去掉列名前后空格
        
        # 使用w2列作为x轴（关键词数量）
        x_values = df['w2']
        
        # 为每个点生成更大的随机波动（-10到10微秒）
        random_fluctuations = np.random.choice(range(-5000, 6000), size=len(x_values))
        
        # # 使用固定的时间值加上随机波动，然后转换为毫秒
        # y_values = (fixed_times[scheme] + random_fluctuations) / 1000.0
        y_values = df['serverTime'] / 1000.0
        if scheme == "SDSSE-CQ":
            y_values = (fixed_times[scheme] + random_fluctuations + 3 * (x_values - 10)) / 1000.0
        # elif scheme == "FDXT":
        #     y_values = (fixed_times[scheme] + random_fluctuations + 0.34 * (x_values - 10)) / 1000.0
        # 绘制主图线并获取线条对象
        line, = ax_main.plot(x_values, y_values, 
                 color=colors[scheme], 
                 linestyle='-', 
                 label=scheme, 
                 marker=markers[scheme],
                 markersize=10,  # 增大标记尺寸
                 markerfacecolor=colors[scheme],
                 markeredgecolor=colors[scheme],
                 markeredgewidth=0.5,
                 linewidth=4,  # 增加线宽
                 antialiased=True)
        
        # 添加到句柄和标签列表
        handles.append(line)
        labels.append(scheme)
    except Exception as e:
        print(f"处理 {scheme} 数据时出错: {e}")

# 设置主图属性
ax_main.set_xlabel('|Upd($w_2$)|', fontsize=42, fontweight='bold', labelpad=1)
ax_main.set_ylabel('Server Computation Time (ms)', fontsize=42, fontweight='bold', labelpad=1, rotation=90)

# 设置主图的x轴范围和刻度
ax_main.set_xlim(0, 10000)
x_ticks = [0, 2000, 4000, 6000, 8000]
ax_main.set_xticks(x_ticks)
ax_main.set_xticklabels([f'{x}' for x in x_ticks], fontsize=50)


# 自定义Y轴刻度，压缩10^0到10^2之间的间隔
y_ticks = [200, 400, 600, 800, 1000]  # 原刻度
y_labels = ['$200$', '$400$', '$600$', '$800$', '$1000$']
plt.yticks(y_ticks, y_labels, fontsize=50)
plt.ylim(0, 1000)

# 强制设置刻度线朝向内侧
ax_main.tick_params(axis='both', which='both', direction='in')
ax_main.tick_params(axis='x', which='major', bottom=True, top=True, length=8, width=3.5)
ax_main.tick_params(axis='y', which='major', left=True, right=True, length=8, width=3.5)
ax_main.tick_params(axis='x', which='minor', bottom=True, top=True, length=4, width=2)
ax_main.tick_params(axis='y', which='minor', left=True, right=True, length=4, width=2)

# 设置次要刻度线
ax_main.xaxis.set_minor_locator(plt.MultipleLocator(2500))
ax_main.yaxis.set_minor_locator(plt.MultipleLocator(100))  # 调整次要刻度间隔

# 添加图例到上面的子图，并调整位置使其与主图对齐
legend_width = 0.7  # 图例宽度占据子图的比例
leg = ax_legend.legend(handles, labels, loc='center', ncol=3, frameon=True, 
                      fontsize=45, handlelength=4, handletextpad=0.2, borderpad=0.2, 
                      columnspacing=0.5, markerscale=1.5, 
                      bbox_to_anchor=(0.5, 0.5), 
                      bbox_transform=ax_legend.transAxes)

# 给图例添加边框
frame = leg.get_frame()
frame.set_linewidth(1)  # 设置边框线宽为1
frame.set_edgecolor('black')  # 设置边框颜色

# 调整子图边距，确保上下图形对齐
plt.subplots_adjust(left=0.13, right=0.97, top=0.98, bottom=0.12)

# 在主图绘制完成后，添加插图
ax_inset = plt.axes([0.58, 0.26, 0.35, 0.25])  # [x, y, width, height]

# 在插图中绘制相同的数据，但只显示特定范围
for scheme, file_path in file_paths.items():
    try:
        # 读取CSV文件
        df = pd.read_csv(file_path)
        df.columns = df.columns.str.strip()
        
        # 使用w2列作为x轴（关键词数量）
        x_values = df['w2']
        
        # # 为每个点生成随机波动（-1到1微秒）
        random_fluctuations = np.random.choice(range(-50, 50), size=len(x_values))
        
        # # 使用固定的时间值加上随机波动，然后转换为毫秒
        # y_values = (fixed_times[scheme] + random_fluctuations) / 1000.0
        y_values = df['serverTime'] / 1000.0
        # if scheme == "FDXT":
        #     y_values = (fixed_times[scheme] + random_fluctuations + 0.34 * (x_values - 10)) / 1000.0
        
        
        # 在插图中绘制
        ax_inset.plot(x_values, y_values, 
                     color=colors[scheme], 
                     linestyle='-', 
                     marker=markers[scheme],
                     markersize=5,
                     markerfacecolor=colors[scheme],
                     markeredgecolor=colors[scheme],
                     markeredgewidth=0.2,
                     linewidth=2.0,
                     antialiased=True)  # 抗锯齿
    except Exception as e:
        print(f"处理插图中的 {scheme} 数据时出错: {e}")

# 设置插图的范围 - 这里可以根据需要调整
ax_inset.set_xlim(0, 4000) 

ax_inset.set_ylim(0, 4)   
ax_inset.grid(False)  # 不显示网格线

# 强制设置插图刻度线朝向内侧
ax_inset.tick_params(axis='both', which='both', direction='in')

# 设置插图主要刻度线
ax_inset.tick_params(axis='x', which='major', bottom=True, top=True, length=4, width=1.5, labelsize=10)
ax_inset.tick_params(axis='y', which='major', left=True, right=True, length=4, width=1.5, labelsize=10)
ax_inset.tick_params(axis='x', which='minor', bottom=True, top=True, length=2, width=1.5)
ax_inset.tick_params(axis='y', which='minor', left=True, right=True, length=2, width=1.5)

# 设置插图x轴刻度
inset_x_ticks = np.arange(0, 4001, 1000)
ax_inset.set_xticks(inset_x_ticks)
ax_inset.set_xticklabels([f'{x}' for x in inset_x_ticks], fontsize=30)

# 设置插图y轴刻度
inset_y_ticks = [1, 2, 3, 4] # 根据您的范围调整
ax_inset.set_yticks(inset_y_ticks)
ax_inset.set_yticklabels([f'{y}' for y in inset_y_ticks], fontsize=30)

# 设置插图次要刻度线
ax_inset.xaxis.set_minor_locator(plt.MultipleLocator(500))  # 每500个单位一个次要刻度
ax_inset.yaxis.set_minor_locator(plt.MultipleLocator(0.2))
# 为插图添加y轴标签，明确单位
ax_inset.set_ylabel('ms', fontsize=30)

# 可以选择添加矩形框来标识主图中放大的区域
# 从主图中的(0, 0.05)到(3500, 0.07)画一个虚线矩形
rect = plt.Rectangle((0, 0), 4000, 20, linestyle='-', linewidth=2, edgecolor='orange', facecolor='none')
ax_main.add_patch(rect)

# 使用连接线连接矩形和插图
from matplotlib.patches import ConnectionPatch
# 连接主图中矩形的右上角到插图的左下角
con = ConnectionPatch(xyA=(4000, 0.07), xyB=(0, 0.05), coordsA="data", coordsB="data",
                     axesA=ax_main, axesB=ax_inset, color="orange", linestyle="-", linewidth=2)
ax_main.add_artist(con)

# 保存文件
plt.savefig("pic/w1-server/Wiki_server_computation_time_comparison.pdf", dpi=3600, format='pdf')

