import hashlib

# 定义一个字典，存储单词及其对应的频率
word_freq = {
    "2001": 246613,
    "pst": 218860,
    "2000": 208551,
    "call": 102807,
    "thu": 93835,
    "question": 83882,
    "follow": 75409,
    "regard": 68923,
    "contact": 60270,
    "energi": 54090,
    "current": 47707,
    "legal": 39923,
    "problem": 31282,
    "industri": 21472,
    "transport": 12879,
    "target": 7311,
    "exactli": 4644,
    "enterpris": 3130
}

# 打开一个文件 "sse_data_lite" 以写入模式
with open("sse_data_lite", "w") as f:
    # 写入字典中单词的数量
    f.write(str(len(word_freq.keys())) + "\n")
    # 遍历字典中的每个单词
    for _k in word_freq.keys():
        # 写入单词
        f.write(_k + "\n")
        # 写入单词的频率
        f.write(str(word_freq[_k]) + "\n")
        # 根据单词的频率生成相应数量的哈希值
        for i in range(word_freq[_k]):
            f.write(hashlib.sha256((_k+str(i)).encode()).hexdigest()[:10] + "\n")
# 打印 "ok" 表示操作完成
print("ok")