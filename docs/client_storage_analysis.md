现在进行Server端的存储分析，我们需要生成ODXT，FDXT和SDSSE-CQ三个方案在三个数据集下的开销。计算方式基于每个keyword-document对的存储大小，需要计算整数的二进制表示占用的空间，例如某个关键词对应的文档对数为10，计算思路是
1. 将整数转换为二进制字符串（去除编程语言自带的前缀，如 Python 中bin()返回的0b）；
2. 统计二进制字符串的长度，得到有效二进制位数；
3. 计算占用的比特数
这是对于对数存储大小的计算方式，我们假设这个函数名称为calculate_binary_storage。三个方案的存储开销定义为：
1. ODXT：storage = calculate_binary_storage(KeywordCount) + 9 * 8
2. FDXT：storage = calculate_binary_storage(KeywordCount) * 1.02 + 9 * 8
3. SDSSE-CQ：定义 d_max = 1644，bf_storage = d_max * 0.0024 * 1024 * 8 * 2。storage = calculate_binary_storage(KeywordCount) * 3 + bf_storage + 9 * 8
同时，存储也是累积的，即每个keyword-document对的存储开销加上前面所有keyword-document对的存储开销。需要生成15个csv文件，分别是：ODXT_Enron.csv, ODXT_Crime.csv, ODXT_Wikipedia.csv, FDXT_Enron.csv, FDXT_Crime.csv, FDXT_Wikipedia.csv, SDSSE-CQ_Enron.csv, SDSSE-CQ_Crime.csv, SDSSE-CQ_Wikipedia.csv。每个csv文件包含对应数据集的行数，例如对于ODXT_Enron.csv，一共有16241行。每个csv文件的第一行是表头，包含两列：分别是`KeywordCount`, `Storage(Bits)`。/raw_data目录下的csv文件包含了三个数据集每个关键词含有的文档对数，在计算存储开销时需要用到这些数据。KeywordCount列的值从1递增至数据集的总keyword-document对数，Storage(Bits)列的值为累积值，即该keyword-document对的存储开销，再加上前面所有keyword-document对的存储开销。例如对于ODXT的Enron数据集，首先查看pic/raw_data/Enron_filecnt_sorted.json，Enron_filecnt_sorted.json的第一行数据为"F19998": 10,那么存储开销计算为calculate_binary_storage(10) + 9 * 8，Enron_filecnt_sorted.json的第二行数据为F16238": 37，那么存储开销为calculate_binary_storage(37) + 9 * 8 + 前面所有keyword-document对的存储开销。
其他csv文件同理。