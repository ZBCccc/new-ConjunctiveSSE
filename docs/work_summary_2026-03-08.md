# 工作总结 - 2026年3月8日

## 一、存储开销数据生成

### 1.1 Server 端存储数据（5个方案）

**数据位置**: `pic/server_data/`

**方案**: FDXT, ODXT, SDSSE-CQ, Mitra, Bestie

**存储计算**:
- 基于 keyword-document 对的累积存储
- 每个方案的存储大小（bytes/pair）:
  - FDXT: 84
  - ODXT: 84
  - SDSSE-CQ: 84
  - Mitra: 64
  - Bestie: 97 (从90更新为97，包含64B hash + 32B密文 + 1B额外开销)

**生成脚本**: `gen_storage_data_v2.go`

**数据集**:
- Enron: 16,242 keywords
- Crime: 63,659 keywords
- Wikipedia: 10,000 keywords

**生成文件**: 15个CSV文件（5方案 × 3数据集）

### 1.2 Client 端存储数据（3个方案）

**数据位置**: `pic/client_data/`

**方案**: ODXT, FDXT, SDSSE-CQ

**存储计算公式**:
```
calculate_binary_storage(n) = bits.Len(n)  // 二进制位数

1. ODXT: storage = calculate_binary_storage(docCount) + 9 * 8
2. FDXT: storage = calculate_binary_storage(docCount) * 2.02 + 9 * 8
   (经过多次修正: 2 → 1.02 → 2.02)
3. SDSSE-CQ: storage = calculate_binary_storage(docCount) * 3 + bf_storage + 9 * 8
   其中 bf_storage = 1644 * 0.0024 * 1024 * 8 * 2 = 64,644 bits
```

**生成脚本**: `gen_client_storage_data.go`

**生成文件**: 9个CSV文件（3方案 × 3数据集）

## 二、图表生成

### 2.1 Server 端存储图表

**输出位置**: `pic/server_storage/`

**特点**:
- 展示5个方案（FDXT, ODXT, SDSSE-CQ, Mitra, Bestie）
- y轴单位: MB (Server Storage Cost)
- 对数刻度
- 3个数据集各一张图

**绘图脚本**:
- `pic/pic-scripts/enron_storage.py`
- `pic/pic-scripts/crime_storage.py`
- `pic/pic-scripts/wikipedia_storage.py`

**图表文件**:
- Enron_storage_comparison.pdf
- Crime_storage_comparison.pdf
- Wikipedia_storage_comparison.pdf

### 2.2 Client 端存储图表

**输出位置**: `pic/client_storage/`

**特点**:
- 展示3个方案（FDXT, ODXT, SDSSE-CQ）
- y轴单位: KB (Client Storage Cost)
- 对数刻度
- 3个数据集各一张图

**绘图脚本**:
- `pic/pic-scripts/enron_client_storage.py`
- `pic/pic-scripts/crime_client_storage.py`
- `pic/pic-scripts/wikipedia_client_storage.py`

**图表文件**:
- Enron_client_storage_comparison.pdf
- Crime_client_storage_comparison.pdf
- Wikipedia_client_storage_comparison.pdf

## 三、图表优化调整

### 3.1 y轴单位调整
- 初始: MB → KB → MB → KB
- 最终确定: Server端使用MB，Client端使用KB

### 3.2 y轴范围调整
经过多次调整，最终确定：

**Server端 (MB)**:
- Enron: $10^{-1}$ 到 $10^{3}$
- Crime: $10^{-2}$ 到 $10^{3}$
- Wikipedia: $10^{-1}$ 到 $10^{3}$

**Client端 (KB)**:
- Enron: $10^{0}$ 到 $10^{5}$
- Crime: $10^{0}$ 到 $10^{6}$
- Wikipedia: $10^{0}$ 到 $10^{5}$

### 3.3 图例优化
- 从5列改为3列（2行布局）
- 后又改回5列（单行布局）
- 最终参数:
  - fontsize: 36 → 45
  - handlelength: 2.5 → 4.2
  - columnspacing: 0.6 → 0.5
  - markerscale: 1.2 → 1.5

### 3.4 刻度线优化
- 添加 y轴次要刻度线（对数刻度自动生成）
- Wikipedia 数据集 x轴次要刻度间隔: 2000 → 1000
- 确保所有数据集刻度线一致

### 3.5 数据点采样优化
尝试了多种采样策略：
1. 统一采样（所有方案相同位置）
2. 错开采样（不同偏移量）
3. 随机采样（不同随机种子）
4. 起点一致+随机采样
5. 最终回退到统一采样

目的：解决 FDXT/ODXT/SDSSE-CQ 数据点重合问题

## 四、关键修正

### 4.1 Bestie 存储开销修正
- 原值: 90 bytes
- 修正为: 97 bytes
- 组成:
  - hash(L||D): 64 bytes (2 × SHA-256)
  - C (密文): 32 bytes (IV 16B + 密文 16B)
  - 额外开销: 1 byte

### 4.2 FDXT Client端公式修正
- 第一版: `* 2`
- 第二版: `* 1.02`
- 第三版: `* 2.02` (最终版本)

### 4.3 SDSSE-CQ Client端公式修正
- 原公式: `calculate_binary_storage(docCount) + bf_storage + 9 * 8`
- 修正为: `calculate_binary_storage(docCount) * 3 + bf_storage + 9 * 8`

## 五、目录结构

```
pic/
├── raw_data/              # 原始JSON数据
│   ├── Enron_filecnt_sorted.json
│   ├── Crime_filecnt_sorted.json
│   └── Wiki_filecnt_sorted.json
├── server_data/           # Server端存储数据 (5方案×3数据集=15个CSV)
├── client_data/           # Client端存储数据 (3方案×3数据集=9个CSV)
├── server_storage/        # Server端存储图表 (3个PDF)
├── client_storage/        # Client端存储图表 (3个PDF)
└── pic-scripts/           # 所有绘图脚本
```

## 六、文档更新

### 6.1 创建的文档
- `docs/storage_data.md` - Server端存储数据生成说明
- `docs/storage_analysis.md` - 存储开销分析
- `docs/bestie_storage_analysis.md` - Bestie方案详细分析
- `docs/client_storage_analysis.md` - Client端存储分析

### 6.2 更新的文档
- 多次更新存储开销定义
- 修正计算公式
- 更新存储比例

## 七、技术要点

### 7.1 数据生成
- 使用Go语言编写数据生成脚本
- 逐行解析JSON保持顺序
- 累积计算存储开销
- 二进制位数计算: `bits.Len(uint(n))`

### 7.2 图表绘制
- 使用Python + matplotlib
- 对数刻度 y轴
- 自定义颜色和标记
- 图例分离布局（顶部图例+底部主图）
- 高分辨率输出 (dpi=3600)

### 7.3 环境配置
- Python环境: conda JXT环境
- 字体: Times New Roman
- 图表尺寸: 18×12 inches

## 八、最终成果

### 8.1 数据文件
- Server端: 15个CSV文件
- Client端: 9个CSV文件
- 总计: 24个数据文件

### 8.2 图表文件
- Server端: 3个PDF
- Client端: 3个PDF
- 总计: 6个图表文件

### 8.3 代码文件
- 数据生成脚本: 2个Go文件
- 绘图脚本: 6个Python文件
- 总计: 8个脚本文件

## 九、存储比例对比

### Server端（5方案）
| 方案 | 存储开销 (bytes/pair) | 相对比例 |
|------|----------------------|---------|
| FDXT/ODXT/SDSSE-CQ | 84 | 1.00x |
| Mitra | 64 | 0.76x |
| Bestie | 97 | 1.15x |

### Client端（3方案）
基于二进制存储计算，存储开销随文档数动态变化。

## 十、遗留问题和注意事项

1. **数据点重合**: FDXT/ODXT/SDSSE-CQ在Server端存储相同，标记会重合
2. **目录命名**: 部分脚本名称包含"server_storage"但实际处理client数据，需注意区分
3. **公式版本**: 确保使用最新的FDXT公式 (* 2.02)
4. **Bestie开销**: 已确认为97 bytes

## 十一、命令记录

### 数据生成
```bash
go run gen_storage_data_v2.go      # Server端数据
go run gen_client_storage_data.go  # Client端数据
```

### 图表生成
```bash
conda activate JXT
python pic/pic-scripts/enron_storage.py
python pic/pic-scripts/crime_storage.py
python pic/pic-scripts/wikipedia_storage.py
python pic/pic-scripts/enron_client_storage.py
python pic/pic-scripts/crime_client_storage.py
python pic/pic-scripts/wikipedia_client_storage.py
```

## 十二、时间线

1. 生成Server端存储数据（5方案）
2. 创建绘图脚本并生成图表
3. 多次调整y轴单位和范围
4. 优化图例大小和布局
5. 添加和调整刻度线
6. 尝试多种数据点采样策略
7. 修正Bestie存储开销（90→97）
8. 生成Client端存储数据（3方案）
9. 修正FDXT公式（2→1.02→2.02）
10. 修正SDSSE-CQ公式（添加*3系数）
11. 最终验证和文档整理

---

**总结**: 今天完成了完整的存储开销数据生成和可视化系统，包括Server端和Client端两套数据，共24个数据文件和6个高质量图表，经过多次优化和修正，确保了数据准确性和图表美观性。
