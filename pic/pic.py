import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

# 预设文件地址（实际使用时替换）
file_paths = [
    "result/Search/ODXT/sse/2025-03-03_11-00-43.csv",
    "result/Search/HDXT/sse/2025-03-06_17-54-01.csv",
    "result/Search/FDXT/sse/2025-02-27_11-16-29.csv",
    "result/Search/SDSSE-CQ/sse/2025-02-27_17-55-40.csv"
]

# 设置不同文件的颜色方案
colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728']

def process_data(file_path):
    """
    处理CSV文件，将时间从微秒转换为毫秒
    """
    df = pd.read_csv(file_path)
    df.columns = df.columns.str.strip()  # 去掉列名前后空格
    
    # 转换时间单位（从微秒到毫秒）
    for col in ['clientTime', 'serverTime', 'totalTime']:
        df[col] = df[col].astype(float) / 1000.0  # 确保使用浮点数除法
    
    return df

def create_plot(file_paths):
    """
    创建三张独立的折线图
    """
    time_types = ['clientTime', 'serverTime', 'totalTime']
    titles = ['Client Time Comparison', 'Server Time Comparison', 'Total Time Comparison']
    
    # 修改y轴的刻度为完整的10的幂次序列
    y_ticks = [10**i for i in range(-1, 8)]  # 从10^-1到10^7，每个幂次都包含
    
    # 为每种时间类型创建单独的图
    for time_type, title in zip(time_types, titles):
        plt.figure(figsize=(12, 8))
        
        # 处理每个文件并绘图
        for idx, file_path in enumerate(file_paths):
            df = process_data(file_path)
            
            # 提取文件名中的算法名称作为图例
            algorithm_name = file_path.split('/')[2]
            
            # 绘制折线
            plt.plot(df.iloc[:, 0], df[time_type], 
                    color=colors[idx], 
                    linestyle='-', 
                    label=algorithm_name, 
                    alpha=0.8,
                    marker='o',  # 添加数据点标记
                    markersize=6)

        # 设置图表属性
        plt.yscale('log')  # 使用对数刻度
        plt.grid(True, which="both", ls="-", alpha=0.2)
        plt.yticks(y_ticks, [f'10^{int(np.log10(y))}' for y in y_ticks])
        
        # 设置标签
        plt.xlabel('Keywords')
        plt.ylabel('Time (ms)')
        plt.title(title)
        
        # 添加图例
        plt.legend(bbox_to_anchor=(1.05, 1), loc='upper left', borderaxespad=0.)
        
        # 调整布局以确保图例完全显示
        plt.tight_layout()
        
        # 保存图表
        output_path = f"pic/result/{time_type}_comparison.png"
        plt.savefig(output_path, dpi=300, bbox_inches='tight')
        print(f"图表已保存到: {output_path}")
        plt.close()  # 关闭图表释放内存

# 调用函数创建图表
create_plot(file_paths)
