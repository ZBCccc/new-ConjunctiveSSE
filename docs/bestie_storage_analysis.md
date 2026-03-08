# Bestie 方案存储开销详细分析

## 1. Bestie 方案概述

Bestie 是一个支持布尔查询的可搜索加密方案。对于每个 keyword-document 对，需要存储：

## 2. 存储结构分析（已确认）

### 2.1 Hash(L||D) - 64 bytes

- **第一个 SHA-256**: 32 bytes
- **第二个 SHA-256**: 32 bytes
- **总计**: 64 bytes

### 2.2 C (密文) - 32 bytes

包含：
- **IV (初始化向量)**: 16 bytes (AES-256-CBC 标准)
- **AES-256-CBC 加密的文档 ID**: 16 bytes (一个 AES 块)
- **总计**: 32 bytes

### 2.3 额外开销 - 1 byte

- **标志位或元数据**: 1 byte

## 3. 最终存储开销

```
Bestie 存储开销：
├── hash(L||D): 64 bytes
├── C: 32 bytes (IV 16B + 密文 16B)
└── 额外开销: 1 byte
总计: 97 bytes/条目
```

## 4. 与其他方案的比较

| 方案 | 存储开销 (bytes/pair) | 相对比例 |
|------|----------------------|---------|
| FDXT/ODXT/SDSSE-CQ | 84 | 1.00x |
| Mitra | 64 | 0.76x |
| Bestie | 97 | 1.15x |

注意：Bestie 的存储开销略高于 FDXT/ODXT/SDSSE-CQ，但提供了不同的安全特性和查询能力。
