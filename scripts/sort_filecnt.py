#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# 输入和输出文件路径
input_file = "cmd/HDXT/configs/Enron_USENIX_filecnt.json"
output_file = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"

# 确保输入文件存在
if not os.path.exists(input_file):
    print(f"错误：找不到输入文件 {input_file}")
    exit(1)

# 读取JSON文件
print(f"正在读取文件 {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# 按值（关键词出现次数）从小到大排序
print("正在按关键词出现次数从小到大排序...")
sorted_data = {k: v for k, v in sorted(data.items(), key=lambda item: item[1])}

# 将排序后的数据写入新文件
print(f"正在写入排序后的数据到 {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    json.dump(sorted_data, f, indent=2, ensure_ascii=False)

print(f"排序完成！结果已保存到 {output_file}")

# 输出一些统计信息
print(f"总关键词数量: {len(data)}")
min_key = min(sorted_data.items(), key=lambda x: x[1])
max_key = max(sorted_data.items(), key=lambda x: x[1])
print(f"出现次数最少的关键词: {min_key[0]} ({min_key[1]}次)")
print(f"出现次数最多的关键词: {max_key[0]} ({max_key[1]}次)") 