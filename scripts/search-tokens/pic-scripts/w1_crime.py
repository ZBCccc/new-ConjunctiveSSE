import pandas as pd
import matplotlib.pyplot as plt
import matplotlib as mpl
from matplotlib import gridspec

# 设置全局刻度线朝向内侧
mpl.rcParams['xtick.direction'] = 'in'
mpl.rcParams['ytick.direction'] = 'in'
# 设置字体为Times New Roman
plt.rcParams['font.family'] = 'Times New Roman'
plt.rcParams['mathtext.fontset'] = 'stix'  # 使数学文本也使用相似的字体
plt.rcParams['font.size'] = 20  # 增加基础字体大小

file_paths = {
    "FDXT": "scripts/search-tokens/data/w1_fdxt_crime_storage_data.csv",
    "ODXT": "scripts/search-tokens/data/w1_odxt_crime_storage_data.csv",
    "SDSSE-CQ": "scripts/search-tokens/data/w1_sdssecq_crime_storage_data.csv"
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
        # Define the interval for sampling
        interval = 1000
        
        # Generate indices for sampling, starting from the first point
        indices = list(range(0, len(df), interval))
        
        # Ensure the last data point is always included
        last_index = len(df) - 1
        if last_index not in indices:
            indices.append(last_index)
            
        # Sort indices to maintain order (optional but good practice)
        indices.sort()

        # Select data points based on the calculated indices
        sampled_df = df.iloc[indices]

        # Use sampled 'updtw2' column for x-axis
        x_values = sampled_df['updtw2']
        
        # Use sampled 'storage' column divided by 1024 for y-axis
        y_values = sampled_df['storage'] / 1024
        
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
ax_main.set_ylabel('Communication Cost (KB)', fontsize=42, fontweight='bold', labelpad=1, rotation=90)

# 设置主图的x轴范围和刻度
ax_main.set_xlim(0, 17000)
x_ticks = [0, 4000, 8000, 12000, 16000]
ax_main.set_xticks(x_ticks)
ax_main.set_xticklabels([f'{x}' for x in x_ticks], fontsize=50)


# 创建带有断点的y轴
ax_main.set_ylim(0, 600)
ax_main.set_yticks([100, 200, 300, 400, 500, 600])
ax_main.set_yticklabels(['100', '200', '300', '400', '500', '600'], fontsize=50)


# 强制设置刻度线朝向内侧
ax_main.tick_params(axis='both', which='both', direction='in')
ax_main.tick_params(axis='x', which='major', bottom=True, top=True, length=8, width=3.5)
ax_main.tick_params(axis='y', which='major', left=True, right=True, length=8, width=3.5)
ax_main.tick_params(axis='x', which='minor', bottom=True, top=True, length=4, width=2)
ax_main.tick_params(axis='y', which='minor', left=True, right=True, length=4, width=2)

# 设置次要刻度线
ax_main.xaxis.set_minor_locator(plt.MultipleLocator(2000))
ax_main.yaxis.set_minor_locator(plt.MultipleLocator(50))  # 调整次要刻度间隔

# 添加图例到上面的子图，并调整位置使其与主图对齐
legend_width = 0.7  # 图例宽度占据子图的比例
leg = ax_legend.legend(handles, labels, loc='center', ncol=3, frameon=True, 
                      fontsize=45, handlelength=4.2, handletextpad=0.2, borderpad=0.2, 
                      columnspacing=0.5, markerscale=1.5, 
                      bbox_to_anchor=(0.5, 0.5), 
                      bbox_transform=ax_legend.transAxes)

# 给图例添加边框
frame = leg.get_frame()
frame.set_linewidth(1)  # 设置边框线宽为1
frame.set_edgecolor('black')  # 设置边框颜色

# 调整子图边距，确保上下图形对齐
plt.subplots_adjust(left=0.12, right=0.97, top=0.98, bottom=0.12)

# 保存文件
plt.savefig("scripts/search-tokens/pic/Crime_w1_comm_cost_comparison.pdf", dpi=3600, format='pdf')

