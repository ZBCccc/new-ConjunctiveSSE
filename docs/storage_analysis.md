# CDB_T Storage Overhead Analysis

## 1. PBC元素大小测量

我们进行了两组PBC元素大小测量，使用不同的曲线参数：

### 测试1: PBC GenerateA(32, 64) - 压缩表示

| 元素类型 | 理论值 | 实际测量值 | 测试样本数 |
|---------|--------|-----------|-----------|
| PBC G1 | ~128 bytes | **16 bytes** | 100, 1000, 10000 |
| PBC Zr | ~32 bytes | **4 bytes** | 100, 1000, 10000 |

### 测试2: 标准Type A曲线 (q ~ 160 bits)

| 元素类型 | 理论值 | 实际测量值 | 测试样本数 |
|---------|--------|-----------|-----------|
| PBC G1 | ~128 bytes | **128 bytes** | 100, 1000, 10000 |
| PBC Zr | ~32 bytes | **20 bytes** | 100, 1000, 10000 |

**重要发现**: PBC库对GenerateA(32, 64)使用压缩表示(16字节)，而标准Type A曲线使用完整的128字节表示。

## 2. 各方案的CDB_T存储结构

基于标准Type A曲线的测量结果，以下是各方案的存储结构：

### 2.1 FDXT / ODXT / SDSSE-CQ (三者理论相同)

```
每个 keyword-document 对的存储：
├── CDBTSet (核心加密结构)
│   ├── addr:  32 bytes  ← PRF输出 (HMAC-SHA256)
│   ├── val:   32 bytes  ← PRF输出 (HMAC-SHA256)
│   └── alpha: 128 bytes ← PBC G1 群元素 (标准Type A)
│
└── CDBXtag (XSet映射)
    ├── l:  32 bytes  ← PRF输出 (HMAC-SHA256)
    └── c:  32 bytes  ← XTag ⊕ T 的结果

总计: 32 + 32 + 128 + 32 + 32 = 256 bytes/条目
```

### 2.2 Mitra (HDXT中的单关键词搜索部分)

```
每个 keyword-document 对的存储：
├── CDBTSet (MitraCipherList)
│   ├── addr: 32 bytes  ← PRF输出
│   └── val:  32 bytes  ← PRF输出
│
└── (无XTag部分)

总计: 32 + 32 = 64 bytes/条目
```

### 2.3 Bestie

```
每个 keyword-document 对的存储：
├── hash(L||D): 64 bytes  ← 2 × SHA-256
└── C: ~26 bytes  ← AES-256-CBC加密(ID) + IV(16B)

总计: 64 + 26 = 90 bytes/条目
```

## 3. 密码学原语大小

### 标准Type A曲线（当前使用）

| 原语 | 算法 | 理论大小 | 实际测量大小 |
|------|------|---------|-------------|
| PRF | HMAC-SHA256 | 32 bytes | 32 bytes |
| Hash | SHA-256 | 32 bytes | 32 bytes |
| PBC G1 | Type A | ~128 bytes | **128 bytes** |
| PBC Zr | Type A | ~32 bytes | **20 bytes** |
| AES | AES-256-CBC | 32+16 bytes | 32+16 bytes |

## 4. 计算公式

```
总存储 = Σ(每个keyword-document对的存储)

对于数据库 D:
- 文档数: N
- 平均每文档关键词数: K
- 总键值对数: N × K

FDXT/ODXT/SDSSE-CQ:  N × K × 256 bytes
Mitra:               N × K × 64  bytes
Bestie:             N × K × 90  bytes
```

## 5. 实际数据库参数

| 数据库 | 文档数 N | 平均关键词数 K | 总键值对 |
|--------|---------|---------------|---------|
| Crime_USENIX_REV | 500 | 25 | 12,500 |
| Enron_USENIX | 2,000 | 20 | 40,000 |
| Wiki_USENIX | 10,000 | 15 | 150,000 |

## 6. 存储开销结果（标准Type A曲线）

### 各方案存储开销 (MB):

| 方案 | Crime (500 docs) | Enron (2000 docs) | Wiki (10000 docs) |
|------|------------------|-------------------|-------------------|
| FDXT/ODXT/SDSSE-CQ | 3.05 MB | 9.77 MB | 36.62 MB |
| Mitra | 0.76 MB | 2.44 MB | 9.16 MB |
| Bestie | 1.07 MB | 3.43 MB | 12.87 MB |

### 存储比例 (相对于FDXT):

| 方案 | 比例 |
|------|------|
| FDXT/ODXT/SDSSE-CQ | 1.00x (baseline) |
| Mitra | 0.25x |
| Bestie | 0.35x |

## 7. 代码实现

- **`pkg/utils/pbc/pbc.go`** - PBC库配置（使用标准Type A曲线）
- **`pkg/utils/pbc/pbc_size_test.go`** - PBC大小测量测试
- **`pkg/utils/storage.go`** - 存储开销计算
- **`cmd/storage/main.go`** - 命令行工具

### 使用方法

```bash
# 运行PBC大小测量测试
go test -v ./pkg/utils/pbc/... -run TestPBCG1Size
go test -v ./pkg/utils/pbc/... -run TestPBCZrSize

# 运行存储计算工具
go run ./cmd/storage/main.go -db Crime_USENIX_REV
go run ./cmd/storage/main.go -db Enron_USENIX
go run ./cmd/storage/main.go -db Wiki_USENIX
```

## 8. 数据文件

- `result/Storage/storage_comparison_typeA_32_64.csv` - GenerateA(32,64) 压缩表示的结果
- `result/Storage/storage_comparison_typeA_standard.csv` - 标准Type A曲线的结果
