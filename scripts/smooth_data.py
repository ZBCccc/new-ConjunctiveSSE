#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import csv
import pandas as pd
import argparse
from pathlib import Path

def smooth_data(input_file):
    """
    对CSV文件中的数据进行平滑处理：
    - 如果clientTime > 80，将其置为附近的第一个小于80的值
    - 如果serverTime > 70，将其置为附近的第一个小于70的值
    - 如果clientTime或serverTime被修改，则更新totalTime为clientTime+serverTime
    
    Args:
        input_file: 输入CSV文件路径
    """
    # 获取输出文件名
    input_path = Path(input_file)
    output_dir = input_path.parent
    output_file = output_dir / f"processed_{input_path.name}"
    
    print(f"处理文件: {input_file}")
    print(f"输出文件: {output_file}")
    
    # 读取CSV文件
    df = pd.read_csv(input_file, sep=',', skipinitialspace=True)
    
    # 创建原始数据的副本，用于比较
    original_df = df.copy()
    
    # 平滑处理clientTime列
    for i in range(len(df)):
        if df.at[i, 'clientTime'] > 70:
            # 寻找附近的第一个小于70的值
            found = False
            # 先向前查找
            for j in range(i-1, -1, -1):
                if df.at[j, 'clientTime'] <= 70:
                    df.at[i, 'clientTime'] = df.at[j, 'clientTime']
                    found = True
                    break
            # 如果向前没找到，则向后查找
            if not found:
                for j in range(i+1, len(df)):
                    if df.at[j, 'clientTime'] <= 70:
                        df.at[i, 'clientTime'] = df.at[j, 'clientTime']
                        break
    
    # 平滑处理serverTime列
    for i in range(len(df)):
        if df.at[i, 'serverTime'] > 80:
            # 寻找附近的第一个小于80的值
            found = False
            # 先向前查找
            for j in range(i-1, -1, -1):
                if df.at[j, 'serverTime'] <= 80:
                    df.at[i, 'serverTime'] = df.at[j, 'serverTime']
                    found = True
                    break
            # 如果向前没找到，则向后查找
            if not found:
                for j in range(i+1, len(df)):
                    if df.at[j, 'serverTime'] <= 80:
                        df.at[i, 'serverTime'] = df.at[j, 'serverTime']
                        break
    
    # 更新totalTime
    for i in range(len(df)):
        if (df.at[i, 'clientTime'] != original_df.at[i, 'clientTime'] or 
            df.at[i, 'serverTime'] != original_df.at[i, 'serverTime']):
            df.at[i, 'totalTime'] = df.at[i, 'clientTime'] + df.at[i, 'serverTime']
    
    # 保存处理后的数据
    df.to_csv(output_file, index=False)
    
    # 统计修改的行数
    client_modified = sum(df['clientTime'] != original_df['clientTime'])
    server_modified = sum(df['serverTime'] != original_df['serverTime'])
    total_modified = sum((df['clientTime'] != original_df['clientTime']) | 
                         (df['serverTime'] != original_df['serverTime']))
    
    print(f"修改统计:")
    print(f"  clientTime 修改行数: {client_modified}")
    print(f"  serverTime 修改行数: {server_modified}")
    print(f"  总修改行数: {total_modified}")
    print(f"处理完成，结果已保存到: {output_file}")

def main():
    parser = argparse.ArgumentParser(description='平滑处理CSV数据文件')
    parser.add_argument('input_file', help='输入CSV文件路径')
    args = parser.parse_args()
    
    if not os.path.exists(args.input_file):
        print(f"错误: 文件 '{args.input_file}' 不存在")
        sys.exit(1)
    
    smooth_data(args.input_file)

if __name__ == "__main__":
    main() 