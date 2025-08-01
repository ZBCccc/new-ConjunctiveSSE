import pandas as pd
import numpy as np
import matplotlib.pyplot as plt

# 读取原始CSV文件
input_file = "result/Search/FDXT/Enron_USENIX/2025-03-30_13-41-01w2_keywords_2.csv"
output_file = "result/Search/FDXT/Enron_USENIX/new_processed_w1_keywords_2.csv"

# 读取数据
df = pd.read_csv(input_file)

# 提取关键列
w1 = df['w1']
client_time = df['clientTime']
server_time = df['serverTime']


def smooth_with_linear_interpolation(x, y, threshold_ratio=1.5):
    """
    使用线性插值方法平滑数据
    当某个值比前后值都大，且比后一个值的threshold_ratio倍还大时，
    将该值替换为基于前一个值和w1比例的线性插值。
    
    x: 自变量（关键词数量）
    y: 因变量（时间）
    threshold_ratio: 阈值比率，超过这个比率的值会被处理
    """
    # 将 x 和 y 按 x 排序
    sorted_indices = np.argsort(x)
    x_sorted = x.iloc[sorted_indices].reset_index(drop=True)
    y_sorted = y.iloc[sorted_indices].reset_index(drop=True)

    # 平滑后的数据
    y_smoothed = y_sorted.copy()

    # 记录处理的异常值数量
    outliers_count = 0

    # 遍历排序后的数据点(跳过第一个和最后一个点)
    for i in range(1, len(x_sorted) - 1):
        prev_value = y_sorted.iloc[i - 1]
        current_value = y_sorted.iloc[i]
        next_value = y_sorted.iloc[i + 1]

        if current_value > prev_value * threshold_ratio:

            # 使用前一个值乘以w1的比例作为新值
            prev_x = x_sorted.iloc[i - 1]
            current_x = x_sorted.iloc[i]
            ratio = current_x / prev_x if prev_x > 0 else 1

            # 理想情况下，时间应该随关键词数量线性增加
            interpolated_value = prev_value * ratio

            # 如果插值结果仍然比后一个值的threshold_ratio倍大，则取平均值
            if interpolated_value > threshold_ratio * next_value:
                interpolated_value = (prev_value + next_value) / 2

            # 更新值
            y_smoothed.iloc[i] = interpolated_value
            y_sorted.iloc[i] = interpolated_value
            outliers_count += 1

            print(f"处理异常值: 索引={i}, 原值={current_value}, 新值={interpolated_value}, "
                  f"w1={current_x}, 前值={prev_value}, 后值={next_value}")

    # 反排序回原始顺序
    y_result = pd.Series(index=y.index)
    y_result.iloc[sorted_indices] = y_smoothed.values

    print(f"共处理了 {outliers_count} 个异常值")
    return y_result


# 应用平滑方法到 clientTime 和 serverTime
smoothed_client_time = smooth_with_linear_interpolation(w1, client_time, threshold_ratio=1.2)
smoothed_server_time = smooth_with_linear_interpolation(w1, server_time, threshold_ratio=1.2)

# 更新DataFrame并计算新的totalTime
df['clientTime'] = smoothed_client_time
df['serverTime'] = smoothed_server_time
df['totalTime'] = df['clientTime'] + df['serverTime']

# 将平滑后的数据保存到新的CSV文件
df.to_csv(output_file, index=False)

print(f"已将平滑后的数据保存到: {output_file}")


# 扩展功能：增加针对后续的多轮处理
# 有时一次处理可能不够，可以多次应用平滑过程
def multi_pass_smooth(x, y, passes=3, threshold_ratio=1.2):
    """
    多轮平滑处理
    """
    current_y = y.copy()
    for i in range(passes):
        current_y = smooth_with_linear_interpolation(x, current_y, threshold_ratio)
        print(f"完成第 {i + 1}/{passes} 轮平滑")
    return current_y

# 如果需要多轮平滑，可以取消下面代码的注释

# multi_smoothed_client_time = multi_pass_smooth(w1, client_time, passes=3)
# multi_smoothed_server_time = multi_pass_smooth(w1, server_time, passes=3)


# # 保存多轮平滑结果
# multi_output_file = "result/Search/SDSSE-CQ/Wiki_USENIX/multi_processed_w2_keywords_2.csv"
# df_multi = df.copy()
# df_multi['clientTime'] = multi_smoothed_client_time
# df_multi['serverTime'] = multi_smoothed_server_time
# df_multi['totalTime'] = df_multi['clientTime'] + df_multi['serverTime']
# df_multi.to_csv(multi_output_file, index=False)
# print(f"已将多轮平滑后的数据保存到: {multi_output_file}")
