# CDB_T Storage Overhead Analysis

## 1. 重要发现：Alpha是Zr元素，不是G1元素！

通过分析代码，我们发现：

### 代码分析

**ComputeAlpha函数** (`pkg/utils/cryptoUtil.go:54-72`):
```go
func ComputeAlpha(Ky, Kz, id []byte, op int, wWc []byte) (*pbc.Element, *pbc.Element, error) {
    alpha1, err := pbcUtil.PrfToZr(Ky, idOp)  // 返回 Zr 元素
    if err != nil {
        return nil, nil, err
    }
    alpha2, err := pbcUtil.PrfToZr(Kz, wWc)  // 返回 Zr 元素
    if err != nil {
        return nil, nil, err
    }
    alpha := pbcUtil.ZrDiv(alpha1, alpha2)   // Zr除法，返回 Zr 元素
    return alpha, alpha1, nil
}
```

- `alpha1` = `PrfToZr(...)` → **Zr元素**
- `alpha2` = `PrfToZr(...)` → **Zr元素**
- `alpha` = `ZrDiv(alpha1, alpha2)` → **Zr元素** (不是G1!)

### 各方案中的Alpha

| 方案 | Alpha变量 | 类型 | 存储位置 |
|------|----------|------|---------|
| FDXT | `alpha` | **Zr (20B)** | CDBTSet |
| ODXT | `alpha` | **Zr (20B)** | CDBTSet |
| SDSSE-CQ | `y` | **Zr (20B)** | TSet |

注意：XTag (G1元素, 128B) 存储在CDBXtag中，但Alpha (Zr元素, 20B) 存储在CDBTSet中。

## 2. PBC元素大小测量

### 标准Type A曲线 (q ~ 160 bits)

| 元素类型 | 理论值 | 实际测量值 |
|---------|--------|-----------|
| PBC G1 | ~128 bytes | 128 bytes |
| PBC Zr | ~32 bytes | **20 bytes** |

## 3. 各方案的CDB_T存储结构

### 3.1 FDXT / ODXT / SDSSE-CQ

```
每个 keyword-document 对的存储：
├── CDBTSet
│   ├── addr:  32 bytes  ← PRF输出 (HMAC-SHA256)
│   ├── val:   32 bytes  ← PRF输出 (HMAC-SHA256)
│   └── alpha: 20 bytes  ← PBC Zr 元素 (不是G1!)
│
└── CDBXtag
    ├── l:  32 bytes  ← PRF输出
    └── c:  32 bytes  ← XTag ⊕ T

总计: 32 + 32 + 20 + 32 + 32 = 148 bytes/条目
```

### 3.2 Mitra

```
每个 keyword-document 对的存储：
├── CDBTSet (MitraCipherList)
│   ├── addr: 32 bytes  ← PRF输出
│   └── val:  32 bytes  ← PRF输出
│
└── (无XTag部分)

总计: 32 + 32 = 64 bytes/条目
```

### 3.3 Bestie

```
每个 keyword-document 对的存储：
├── hash(L||D): 64 bytes  ← 2 × SHA-256
├── C: 32 bytes  ← AES-256-CBC加密(ID) + IV(16B)
└── 额外开销: 1 byte  ← 标志位或元数据

总计: 64 + 32 + 1 = 97 bytes/条目
```

## 4. 存储开销结果

### 各方案存储开销 (MB):

| 方案 | Crime (500 docs) | Enron (2000 docs) | Wiki (10000 docs) |
|------|------------------|-------------------|-------------------|
| FDXT/ODXT/SDSSE-CQ | 1.76 MB | 5.65 MB | 21.17 MB |
| Mitra | 0.76 MB | 2.44 MB | 9.16 MB |
| Bestie | 1.07 MB | 3.43 MB | 12.87 MB |

### 存储比例 (相对于FDXT):

| 方案 | 比例 |
|------|------|
| FDXT/ODXT/SDSSE-CQ | 1.00x (baseline) |
| Mitra | 0.76x |
| Bestie | 1.15x |

## 5. 总结

- **Alpha是Zr元素(20字节)，不是G1元素(128字节)**！
- 这是通过分析ComputeAlpha函数的代码发现的
- Alpha = ZrDiv(alpha1, alpha2)，其中alpha1和alpha2都是Zr元素
- XTag是G1元素，存储在CDBXtag中，大小为128字节
- 修正后的存储开销：FDXT/ODXT/SDSSE-CQ = 148 bytes/条目 (之前错误计算为256 bytes)
