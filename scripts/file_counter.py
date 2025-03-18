#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
import pymongo
import argparse
import sys
from collections import Counter
from tqdm import tqdm

def main():
    # 解析命令行参数
    parser = argparse.ArgumentParser(description='统计MongoDB中每个关键词拥有的文件ID数量')
    parser.add_argument('--host', default='localhost', help='MongoDB主机地址')
    parser.add_argument('--port', type=int, default=27017, help='MongoDB端口')
    parser.add_argument('--db', default='Crime_USENIX_REV_toy', help='MongoDB数据库名称')
    parser.add_argument('--collection', default='id_keywords', help='MongoDB集合名称')
    parser.add_argument('--output', default='cmd/HDXT/configs/filecnt_toy.json', help='输出JSON文件路径')
    parser.add_argument('--batch-size', type=int, default=1000, help='批处理大小')
    args = parser.parse_args()

    try:
        # 连接到MongoDB
        client = pymongo.MongoClient(f"mongodb://{args.host}:{args.port}/", 
                                    serverSelectionTimeoutMS=5000)
        
        # 检查连接是否成功
        client.server_info()
        print(f"成功连接到MongoDB: {args.host}:{args.port}")
        
        # 选择数据库和集合
        db = client[args.db]
        collection = db[args.collection]
        
        # 获取文档总数
        total_docs = collection.count_documents({})
        print(f"集合 '{args.collection}' 中共有 {total_docs} 个文档")
        
        if total_docs == 0:
            print("警告: 集合为空，没有数据可处理")
            return
        
        # 初始化关键词计数器
        keyword_counter = Counter()
        
        # 使用批处理和进度条处理大量数据
        print("正在统计关键词文件数量...")
        cursor = collection.find({}, batch_size=args.batch_size)
        
        for doc in tqdm(cursor, total=total_docs, desc="处理进度"):
            # 确保文档包含所需字段
            if "id" in doc and "val_st" in doc:
                # 获取关键词列表
                keywords = doc["val_st"]
                
                # 对每个关键词计数（去重，每个ID只计算一次）
                unique_keywords = set(keywords)
                for keyword in unique_keywords:
                    keyword_counter[keyword] += 1
        
        # 将计数器转换为字典
        keyword_count_dict = dict(keyword_counter)
        
        # 保存结果到JSON文件
        with open(args.output, "w", encoding="utf-8") as f:
            json.dump(keyword_count_dict, f, ensure_ascii=False, indent=2)
        
        print(f"关键词文件计数已保存到 {args.output}")
        print(f"共统计了 {len(keyword_count_dict)} 个关键词")
        
    except pymongo.errors.ServerSelectionTimeoutError:
        print(f"错误: 无法连接到MongoDB服务器 {args.host}:{args.port}")
        sys.exit(1)
    except pymongo.errors.OperationFailure as e:
        print(f"MongoDB操作失败: {e}")
        sys.exit(1)
    except Exception as e:
        print(f"发生错误: {e}")
        sys.exit(1)
    finally:
        if 'client' in locals():
            client.close()
            print("MongoDB连接已关闭")

if __name__ == "__main__":
    main()
