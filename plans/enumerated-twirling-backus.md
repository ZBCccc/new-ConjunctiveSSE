# SDSSE-CQ gRPC Client-Server 分离改造计划

## 背景

当前 FDXT、ODXT、HDXT 已完成 gRPC 客户端-服务器分离模式改造，但 SDSSE-CQ 仍是单体实现，需要进行相同模式的改造。

## 现状分析

### SDSSE-CQ 当前结构
- **入口点**: `cmd/SDSSE-CQ/main.go` (单体)
- **客户端逻辑**: `pkg/SDSSE-CQ/Client/client.go` (嵌入在客户端)
- **主逻辑**: `pkg/SDSSE-CQ/SDSSE-CQ.go`
- **Proto**: `pkg/SDSSE-CQ/proto/sdssecq.proto` (服务名错误为 ODXTService，未生成代码)

### 需要新增的组件 (参考 FDXT/ODXT 模式)
1. 修正并生成 proto 文件
2. 创建 `cmd/SDSSE-CQ/server/main.go`
3. 创建 `cmd/SDSSE-CQ/client/main.go`
4. 创建 `pkg/SDSSE-CQ/server/server.go`
5. 重构 `pkg/SDSSE-CQ/Client/client.go` 为 gRPC 客户端包装

## 改造步骤

### 1. 修正并生成 Proto 文件
**文件**: `pkg/SDSSE-CQ/proto/sdssecq.proto`

修改服务名为 `SDSSEcqService`，添加 Setup (streaming) 和 Search RPC：
```protobuf
service SDSSEcqService {
  rpc Setup(stream SetupRequest) returns (SetupResponse) {}
  rpc Update(UpdateRequest) returns (UpdateResponse) {}
  rpc Search(SearchRequest) returns (SearchResponse) {}
}
```

生成代码:
```bash
cd /Users/bytedance/Code/Personal/new-ConjunctiveSSE
./scripts/gen-proto.sh
```

### 2. 创建 Server 实现
**文件**: `pkg/SDSSE-CQ/server/server.go`

实现 gRPC server 接口：
- 存储 TSet 和 XSet (使用 Aura 的 SSEClient)
- `Setup()` - 接收客户端流式发送的加密数据
- `Update()` - 接收更新操作
- `Search()` - 执行搜索（当前单体代码中的 Server side 部分）

### 3. 创建 Server 入口点
**文件**: `cmd/SDSSE-CQ/server/main.go`

参考 `cmd/FDXT/server/main.go`:
- 监听端口 **50051** (与 ODXT/FDXT/HDXT 一致)
- 设置大消息尺寸限制
- 注册 SDSSEcqServiceServer

### 4. 重构 Client 包装
**文件**: `pkg/SDSSE-CQ/Client/client.go`

重构为 gRPC 客户端：
- 包装 gRPC 连接
- 实现 Setup(), Update(), Search() 方法调用 gRPC 服务
- 保留本地密码学操作（密钥生成、trapdoor 生成、解密）

### 5. 创建 Client 入口点
**文件**: `cmd/SDSSE-CQ/client/main.go`

参考 `cmd/ODXT/client/main.go`:
- 连接 MongoDB 获取数据
- 连接 gRPC 服务器
- 实现 Setup Phase (CiphertextGen) 和 Search Phase

### 6. 创建 Convert 工具
**文件**:
- `pkg/SDSSE-CQ/client/convert.go`
- `pkg/SDSSE-CQ/server/convert.go`

处理 protobuf 类型与内部类型的转换。

### 6. 删除旧的单体入口点
**文件**: `cmd/SDSSE-CQ/main.go`

在验证新架构正常工作后，删除旧的单体入口点。

## 关键文件路径

| 组件 | 路径 |
|------|------|
| Proto 定义 | `pkg/SDSSE-CQ/proto/sdssecq.proto` |
| Server 实现 | `pkg/SDSSE-CQ/server/server.go` |
| Server 入口 | `cmd/SDSSE-CQ/server/main.go` |
| Client 包装 | `pkg/SDSSE-CQ/Client/client.go` |
| Client 入口 | `cmd/SDSSE-CQ/client/main.go` |

## 验证方法

1. **编译验证**: `go build ./cmd/SDSSE-CQ/server` 和 `go build ./cmd/SDSSE-CQ/client`
2. **运行测试**:
   - 启动 server: `go run cmd/SDSSE-CQ/server/main.go`
   - 运行 client: `go run cmd/SDSSE-CQ/client/main.go -db=Crime_USENIX_REV -phase=cs`
3. **对比结果**: 与单体版本输出对比，确保功能一致

## HDXT 状态说明

HDXT 已完成 gRPC 分离，无需进一步改造。
