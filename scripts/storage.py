import json
import os

def calculate_storage(json_file_path):
    """
    计算JSON文件中每个关键词的空间占用
    计算公式：(行号+1)*文件数
    """
    try:
        # 检查文件是否存在
        if not os.path.exists(json_file_path):
            print(f"错误：文件 {json_file_path} 不存在")
            return
        
        # 读取JSON文件
        with open(json_file_path, 'r', encoding='utf-8') as f:
            data = json.load(f)
        
        # 计算总空间占用
        csv_file_path = json_file_path.replace('.json', '_storage_1.csv')
        with open(csv_file_path, 'w', encoding='utf-8') as csv_file:
            csv_file.write("KeywordCount, Counter, Storage(Bits)\n")
            total_counter = 0
            for i, (keyword, file_count) in enumerate(data.items()):
                total_counter += file_count
                storage = (i + 1) * calculate_binary_storage(file_count)
                csv_file.write(f"{i + 1},{file_count},{storage}\n")
        print(f"存储计算结果已保存到 {csv_file_path}")
            
    except json.JSONDecodeError:
        print("错误：JSON格式无效")
    except Exception as e:
        print(f"发生错误: {str(e)}")

def calculate_binary_storage(number):
    """
    计算整数的二进制表示占用的空间
    
    参数:
        number (int): 要计算的整数
    
    返回:
        int: 二进制表示占用的空间（比特数）
        str: 二进制表示
        int: 字节数（向上取整到最近的字节）
    """
    # 确保输入是正整数
    if not isinstance(number, int):
        raise TypeError("输入必须是整数")
    
    # 处理负数
    if number < 0:
        # 对于负数，Python的bin会输出'-0b...'格式
        # 我们先取绝对值，之后再添加一位用于符号位
        abs_number = abs(number)
        binary = bin(abs_number)[2:]  # 去掉'0b'前缀
        # 对于负数需要额外的一位来表示符号
        bits = len(binary) + 1
    else:
        # 对于0和正数
        if number == 0:
            binary = '0'
            bits = 1
        else:
            binary = bin(number)[2:]  # 去掉'0b'前缀
            bits = len(binary)
    
    # 计算占用的字节数（向上取整）
    # bytes_count = (bits + 7) // 8  # 每8位为一个字节，不足8位的按一个字节计算
    
    return bits

if __name__ == "__main__":
    # 文件路径
    json_file_path = "cmd/HDXT/configs/Enron_USENIX_filecnt_sorted.json"
    
    # 计算空间占用
    calculate_storage(json_file_path)
