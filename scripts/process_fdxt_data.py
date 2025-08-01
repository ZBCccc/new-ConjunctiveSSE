#!/usr/bin/env python
# -*- coding: utf-8 -*-

import csv
import argparse
import os
from datetime import datetime

def process_csv_files(file1_path, file2_path, output_path=None):
    """
    处理两个CSV文件，生成新的CSV文件，其中clientTime和serverTime的值为
    file2的值加上(file2的值-file1的值)的10%
    """
    # 如果未指定输出路径，创建带时间戳的默认文件名
    if not output_path:
        timestamp = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
        output_path = f"result/processed_{timestamp}.csv"
    
    # 确保输出目录存在
    os.makedirs(os.path.dirname(output_path), exist_ok=True)
    
    # 读取第一个文件
    file1_data = []
    with open(file1_path, 'r', newline='') as f1:
        reader = csv.DictReader(f1)
        # 处理列名可能存在的空格
        clean_fieldnames = [field.strip() for field in reader.fieldnames]
        for row in reader:
            # 创建一个新的清理后的行
            clean_row = {}
            for key, value in row.items():
                clean_row[key.strip()] = value.strip() if isinstance(value, str) else value
            file1_data.append(clean_row)
    
    # 读取第二个文件
    file2_data = []
    with open(file2_path, 'r', newline='') as f2:
        reader = csv.DictReader(f2)
        # 处理列名可能存在的空格
        clean_fieldnames = [field.strip() for field in reader.fieldnames]
        for row in reader:
            # 创建一个新的清理后的行
            clean_row = {}
            for key, value in row.items():
                clean_row[key.strip()] = value.strip() if isinstance(value, str) else value
            file2_data.append(clean_row)
    
    # 检查两个文件行数是否一致
    if len(file1_data) != len(file2_data):
        raise ValueError(f"文件行数不一致：文件1有{len(file1_data)}行，文件2有{len(file2_data)}行")
    
    # 处理数据并写入新文件
    with open(output_path, 'w', newline='') as out_file:
        # 获取表头（使用第一个文件的表头，去除空格）
        fieldnames = [key.strip() for key in file1_data[0].keys()]
        writer = csv.DictWriter(out_file, fieldnames=fieldnames)
        writer.writeheader()
        
        # 对每一行进行处理
        for i in range(len(file1_data)):
            row1 = file1_data[i]
            row2 = file2_data[i]
            new_row = {}
            
            # 复制其他字段（从第二个文件复制，因为我们以第二个文件为基准）
            for key in row2.keys():
                clean_key = key.strip()
                new_row[clean_key] = row2[key]
            
            # 处理clientTime：file2值 + (file2值 - file1值) * 10%
            client_key = 'clientTime'
            client_time1 = int(row1[client_key])
            client_time2 = int(row2[client_key])
            new_client_time = client_time2 + (client_time1 - client_time2) * 0.1
            new_row[client_key] = str(int(round(new_client_time)))  # 四舍五入并转为整数
            
            # 处理serverTime：file2值 + (file2值 - file1值) * 10%
            server_key = 'serverTime'
            server_time1 = int(row1[server_key])
            server_time2 = int(row2[server_key])
            new_server_time = server_time2 + (server_time1 - server_time2) * 0.1
            new_row[server_key] = str(int(round(new_server_time)))  # 四舍五入并转为整数
            
            # 计算新的totalTime = clientTime + serverTime
            new_row['totalTime'] = str(int(new_row[client_key]) + int(new_row[server_key]))
            
            writer.writerow(new_row)
    
    print(f"处理完成，结果已保存至：{output_path}")
    return output_path

def main():
    file1 = "result/Search/FDXT/Crime_USENIX_REV/2025-03-30_22-27-01_w1_keywords_2.csv"
    file2 = "result/Search/FDXT/Crime_USENIX_REV/2025-03-30_22-27-11_w1_keywords_2.csv"
    
    output = "result/Search/FDXT/Crime_USENIX_REV/final_w2_keywords_22.csv"
    process_csv_files(file1, file2, output)

if __name__ == "__main__":
    main()
