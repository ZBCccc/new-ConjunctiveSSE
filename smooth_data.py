import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# 读取原始CSV文件
input_file = "result/Search/FDXT/Wiki_USENIX/2025-03-25_12-05-29w2_keywords_2.csv"
output_file = "result/Search/FDXT/Wiki_USENIX/2025-03-25_12-05-29w2_keywords_2_smoothed.csv"

# 读取数据
df = pd.read_csv(input_file)

# 提取关键列：w1 作为关键词数量，clientTime 和 serverTime 是需要平滑的指标
w1 = df['w1']
client_time = df['clientTime']
server_time = df['serverTime']


# 平滑数据的函数
def smooth_outliers(x, y, window_size=20, threshold=2.0):
    """
    使用移动窗口平滑异常值
    x: 自变量（关键词数量）
    y: 因变量（时间）
    window_size: 移动窗口大小
    threshold: 异常值阈值（标准差的倍数）
    """
    # 将 x 和 y 按 x 排序
    sorted_indices = np.argsort(x)
    x_sorted = x.iloc[sorted_indices].reset_index(drop=True)
    y_sorted = y.iloc[sorted_indices].reset_index(drop=True)
    
    # 平滑后的数据
    y_smoothed = y_sorted.copy()
    
    # 使用移动窗口检测和平滑异常值
    for i in range(len(x_sorted)):
        # 找到窗口范围内的数据点
        lower_bound = max(0, i - window_size // 2)
        upper_bound = min(len(x_sorted) - 1, i + window_size // 2)
        window_indices = range(lower_bound, upper_bound + 1)
        
        # 排除当前点计算窗口统计值
        window_values = [y_sorted.iloc[j] for j in window_indices if j != i]
        
        if len(window_values) > 0:
            window_mean = np.mean(window_values)
            window_std = np.std(window_values)
            
            # 如果当前值偏离平均值太远，则认为是异常值
            if window_std > 0:  # 避免除零错误
                z_score = abs(y_sorted.iloc[i] - window_mean) / window_std
                
                if z_score > threshold:
                    # 使用线性回归预测值替代异常值
                    if i > 0 and i < len(x_sorted) - 1:
                        # 使用前后点的线性插值
                        prev_x, prev_y = x_sorted.iloc[i-1], y_sorted.iloc[i-1]
                        next_x, next_y = x_sorted.iloc[i+1], y_sorted.iloc[i+1]
                        
                        # 如果前一个或后一个点也是异常值，则使用窗口平均值
                        if (abs(prev_y - window_mean) / window_std > threshold or 
                            abs(next_y - window_mean) / window_std > threshold):
                            y_smoothed.iloc[i] = window_mean
                        else:
                            # 线性插值
                            slope = (next_y - prev_y) / (next_x - prev_x)
                            current_x = x_sorted.iloc[i]
                            y_smoothed.iloc[i] = prev_y + slope * (current_x - prev_x)
                    else:
                        # 边界点使用窗口平均值
                        y_smoothed.iloc[i] = window_mean
    
    # 反排序回原始顺序
    y_result = pd.Series(index=y.index)
    y_result.iloc[sorted_indices] = y_smoothed.values
    
    return y_result

# 另一种更简单的方法：使用多项式拟合来预测趋势线
def polynomial_smooth(x, y, degree=3, threshold=2.0):
    """
    使用多项式拟合来平滑异常值
    x: 自变量（关键词数量）
    y: 因变量（时间）
    degree: 多项式次数
    threshold: 异常值阈值（标准差的倍数）
    """
    # 将 NaN 和无穷大值替换为合适的值
    y_clean = y.copy()
    mask = ~np.isfinite(y_clean)
    y_clean[mask] = np.nanmean(y_clean)  # 使用平均值替换非有限值
    
    # 拟合多项式模型
    coeffs = np.polyfit(x, y_clean, degree)
    poly_model = np.poly1d(coeffs)
    
    # 计算预测值
    y_pred = poly_model(x)
    
    # 计算残差
    residuals = y_clean - y_pred
    
    # 计算残差的标准差
    res_std = np.std(residuals)
    
    # 找出异常值
    outliers = (np.abs(residuals) > threshold * res_std)
    
    # 创建平滑后的数据，异常值采用模型预测值，其他保持原值
    y_smoothed = y.copy()
    y_smoothed[outliers] = y_pred[outliers]
    
    return y_smoothed

# 分段平滑，处理不同区间的数据
def segment_smooth(df, column, window_sizes=[20, 100, 200], thresholds=[2.0, 2.5, 3.0], segment_points=[400, 1500]):
    """
    对不同数据段使用不同参数进行平滑
    df: 数据框
    column: 要平滑的列名
    window_sizes: 每个段的窗口大小
    thresholds: 每个段的阈值
    segment_points: 分段点列表
    """
    result = df[column].copy()
    
    # 定义段
    segments = [(0, segment_points[0])]
    for i in range(len(segment_points) - 1):
        segments.append((segment_points[i], segment_points[i+1]))
    segments.append((segment_points[-1], df.shape[0]))
    
    for i, ((start, end), window_size, threshold) in enumerate(zip(segments, window_sizes, thresholds)):
        segment_df = df.iloc[start:end]
        
        print(f"处理段 {i+1}/{len(segments)}: 行 {start}-{end}, 窗口大小={window_size}, 阈值={threshold}")
        
        # 对该段应用平滑
        smoothed_segment = smooth_outliers(segment_df['w1'], segment_df[column], window_size=window_size, threshold=threshold)
        
        # 将平滑后的值添加到结果中
        result.iloc[start:end] = smoothed_segment.values
    
    return result

# 分别对 clientTime 和 serverTime 应用平滑处理
smoothed_client_time = segment_smooth(df, 'clientTime', 
                                      window_sizes=[20, 50, 100], 
                                      thresholds=[2.0, 2.5, 3.0], 
                                      segment_points=[400, 1500])

smoothed_server_time = segment_smooth(df, 'serverTime', 
                                      window_sizes=[20, 50, 100], 
                                      thresholds=[2.0, 2.5, 3.0], 
                                      segment_points=[400, 1500])

# 创建平滑后的散点图
plt.figure(figsize=(12, 6))
plt.scatter(w1, smoothed_client_time, alpha=0.5, label='平滑后 ClientTime')
plt.title('平滑后数据 - ClientTime vs 关键词数量')
plt.xlabel('关键词数量 (w1)')
plt.ylabel('ClientTime (ms)')
plt.xlim(0, 3000)
plt.ylim(0, 3000)
plt.savefig("smoothed_client_time.png")

# 创建原始和平滑数据对比图
plt.figure(figsize=(12, 8))
plt.scatter(w1, client_time, alpha=0.3, label='原始 ClientTime', color='blue')
plt.scatter(w1, smoothed_client_time, alpha=0.3, label='平滑后 ClientTime', color='red')
plt.title('原始 vs 平滑后数据 - ClientTime')
plt.xlabel('关键词数量 (w1)')
plt.ylabel('ClientTime (ms)')
plt.legend()
plt.xlim(0, 3000)
plt.ylim(0, 3000)
plt.savefig("comparison_client_time.png")

# 更新DataFrame并计算新的totalTime
df['clientTime'] = smoothed_client_time
df['serverTime'] = smoothed_server_time
df['totalTime'] = df['clientTime'] + df['serverTime']

# 将平滑后的数据保存到新的CSV文件
df.to_csv(output_file, index=False)

print(f"已将平滑后的数据保存到: {output_file}") 