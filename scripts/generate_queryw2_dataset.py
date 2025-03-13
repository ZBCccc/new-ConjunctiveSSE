#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# 输入和输出文件路径
input_file = "cmd/ODXT/configs/filecnt_sorted.json"
output_file = "cmd/ODXT/configs/w2_keywords_2.txt"

# 确保输入文件存在
if not os.path.exists(input_file):
    print(f"错误：找不到输入文件 {input_file}")
    exit(1)

# 读取JSON文件
print(f"正在读取文件 {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# 找到出现次数最多的第一个关键词
max_count = 0
first_keyword = None
for keyword, count in data.items():
    if count > max_count:
        max_count = count
        first_keyword = keyword

if not first_keyword:
    print("错误：找不到关键词")
    exit(1)

print(f"第一个关键词（出现次数最多）: {first_keyword}，出现次数: {max_count}")

# 提取所有不同出现次数的关键词，每个出现次数只取第一个
second_keywords = []
current_count = None
for keyword, count in data.items():
    if count != current_count:
        second_keywords.append(keyword)
        current_count = count

print(f"找到 {len(second_keywords)} 个不同出现次数的关键词")

# 生成查询数据集
print(f"正在生成查询数据集到 {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    for second_keyword in second_keywords:
        query = f"{first_keyword}#{second_keyword}\n"
        f.write(query)

print(f"查询数据集生成完成！共生成 {len(second_keywords)} 条查询") 