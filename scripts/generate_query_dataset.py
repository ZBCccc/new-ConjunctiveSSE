#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import os

# 输入和输出文件路径
input_file = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"
output_file = "cmd/HDXT/configs/Enron_USENIX_keywords_w1_generated.txt"

# 确保输入文件存在
if not os.path.exists(input_file):
    print(f"错误：找不到输入文件 {input_file}")
    exit(1)

# 读取JSON文件
print(f"正在读取文件 {input_file}...")
with open(input_file, 'r', encoding='utf-8') as f:
    data = json.load(f)

# 找到出现次数为10的第一个关键词
first_keyword = None
for keyword, count in data.items():
    if count == 10:
        first_keyword = keyword
        break

if not first_keyword:
    print("错误：找不到出现次数为10的关键词")
    exit(1)

print(f"第一个关键词（出现次数为10）: {first_keyword}")

# 提取出现次数大于10的关键词，每个出现次数只取第一个
second_keywords = []
current_count = None
for keyword, count in data.items():
    if count > 10:
        if count != current_count:
            second_keywords.append(keyword)
            current_count = count

print(f"找到 {len(second_keywords)} 个不同出现次数（>10）的关键词")

# 生成查询数据集
print(f"正在生成查询数据集到 {output_file}...")
with open(output_file, 'w', encoding='utf-8') as f:
    for second_keyword in second_keywords:
        query = f"{first_keyword}#{second_keyword}\n"
        f.write(query)

print(f"查询数据集生成完成！共生成 {len(second_keywords)} 条查询") 