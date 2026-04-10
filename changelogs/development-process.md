# 用户请求/响应报文记录功能 - 开发过程文档

## 1. 需求背景

### 1.1 问题发现

用户提出疑问：当前 new-api 工程代码是否会记录用户请求大模型的报文。

### 1.2 调查结论

经过代码分析，确认当前系统**不会**记录完整的请求/响应报文：

- `model/log.go` 中的 `Log` 结构体有 `Content` 字段，但仅记录**计费摘要信息**（如模型倍率、分组倍率等）
- `RecordConsumeLog()` 函数记录的字段为：`UserId`、`ModelName`、`PromptTokens`、`CompletionTokens`、`Quota`、`UseTime`、`ChannelId` 等**元数据**
- **不记录**：用户的 prompt/messages 内容、大模型返回的 completion/response 内容、完整的 request body 或 response body

### 1.3 需求定义

添加用户大模型请求/响应报文的完整记录功能：
- 按用户编码记录到各自对应的日志文件中
- 超过一定大小后创建新的日志文件
- 每个用户最多记录10个文件

## 2. 代码分析

### 2.1 项目架构

项目为 Go 语言编写的 LLM API 代理网关（基于 OneAPI），主要结构：

```
new-api/
├── controller/     # 控制器层，请求入口
│   └── relay.go    # Relay() 函数，所有 API 代理请求的统一入口
├── relay/          # 代理层，转发请求到上游大模型
│   ├── claude_handler.go
│   ├── compatible_handler.go
│   ├── gemini_handler.go
│   └── ...
├── middleware/      # 中间件（认证、限流等）
│   └── auth.go     # TokenAuth() 认证中间件，设置 c.Set("id", token.UserId)
├── model/          # 数据模型
│   └── log.go      # 日志模型（仅计费元数据）
├── logger/         # 日志工具
│   └── logger.go   # 通用日志函数
├── common/         # 公共工具
│   ├── body_storage.go   # 请求体存储（支持内存/磁盘）
│   ├── gin.go           # BodyStorage 管理
│   └── constants.go     # 全局常量
└── router/         # 路由配置
    └── relay-router.go  # /v1/* 路由，使用 TokenAuth + Relay
```

### 2.2 请求处理流程

```
用户请求 → router (TokenAuth) → controller.Relay() → relay handler → 上游大模型
         ↓ 设置 userId                              ↑ 读取 BodyStorage
    middleware/auth.go                          ↓ 记录计费日志
                                         service.PostTextConsumeQuota()
```

### 2.3 关键发现

1. **请求体读取**：`common.GetBodyStorage(c)` 可获取完整的请求体（`storage.Bytes()` 返回 `[]byte`）
2. **用户ID获取**：通过 `c.GetInt("id")` 获取，在 `middleware/auth.go` 的 `TokenAuth()` 中设置
3. **异步写入**：项目使用 `github.com/bytedance/gopkg/util/gopool` 进行异步任务处理
4. **日志目录**：`*common.LogDir` 默认为 `./logs`，通过 `--log-dir` 启动参数配置

## 3. 设计方案

### 3.1 技术方案

- **新增文件**：`logger/request_logger.go` — 独立的报文记录模块
- **修改文件**：`controller/relay.go` — 在 Relay 函数中集成日志记录

### 3.2 核心设计

#### 响应体捕获

使用 `ResponseCaptureWriter` 包装 `gin.ResponseWriter`，在响应写入时同步缓存数据：

```go
type ResponseCaptureWriter struct {
    gin.ResponseWriter
    body   *bytes.Buffer
    status int
}
```

#### 日志轮转

```
{log-dir}/request_logs/user_{id}/
├── request_20260410_120000.log  (20MB)
├── request_20260411_080000.log  (20MB)
└── ...                          (最多10个文件)
```

- 单文件达到 20MB 后创建新文件
- 超过 10 个文件后删除最旧的
- 文件按修改时间排序确定新旧

#### 异步写入

```
Relay() 返回 → recordRequestLog() → logger.LogRequestResponse() → gopool.Go(writeRequestLog)
                     ↑ 同步读取请求/响应体                    ↑ 异步写文件
```

### 3.3 配置方式

通过环境变量 `REQUEST_LOG_ENABLED=true` 启用，避免修改代码重新编译。

## 4. 实现过程

### 4.1 第一版实现

1. 创建 `logger/request_logger.go`，包含：
   - `LogRequestResponse()` — 异步写日志入口
   - `ResponseCaptureWriter` — 响应体捕获包装器
   - `rotateIfNeeded()` — 日志轮转逻辑
   - `formatJSON()` — JSON 格式化

2. 修改 `controller/relay.go`：
   - 添加 `recordRequestLog()` 函数
   - 在 `Relay()` 中包装 `c.Writer`
   - 在请求成功时调用 `recordRequestLog()`

3. 编译验证：`go build ./logger/`、`go build ./controller/`、`go vet` 均通过

### 4.2 流式响应问题

**问题**：第一版对流式响应（SSE）不记录响应体，导致日志中响应为 `(empty)`。

**原因**：`recordRequestLog()` 中有 `!logger.IsStreamResponse(c)` 判断，跳过了流式响应的捕获。

**修复**：移除流式跳过逻辑，统一捕获所有响应。

**风险评估**：流式响应数据量较大，但因为是异步写入且请求结束后一次性读取，不影响请求性能。

### 4.3 配置优化

**问题**：初始版本 `RequestLogEnabled` 为硬编码 `const`，需改代码才能启用。

**修复**：改为通过环境变量控制：

```go
var RequestLogEnabled = os.Getenv("REQUEST_LOG_ENABLED") == "true"
```

### 4.4 参数调优

| 参数 | 初始值 | 最终值 | 原因 |
|------|--------|--------|------|
| `maxLogFileSize` | 50MB | 20MB | 60人场景下控制总占用约12GB |
| `maxLogFilesPerUser` | 10 | 10 | 保持不变 |

## 5. 构建与部署

### 5.1 Docker 镜像构建

```bash
# 构建镜像
docker build --network=host -t new-api:v1.0 .

# 导出镜像
docker save -o new-api-v1.0.tar new-api:v1.0
```

### 5.2 启动命令

```bash
docker run -d --name new-api -p 3000:3000 \
  -e REQUEST_LOG_ENABLED=true \
  -v /data/new-api:/data \
  new-api:v1.0
```

### 5.3 日志查看

```bash
# 进入容器查看日志
docker exec -it new-api ls /data/logs/request_logs/

# 查看某用户日志
docker exec -it new-api cat /data/logs/request_logs/user_16/request_20260411_120000.log
```

## 6. Git 提交记录

| 提交 | 说明 |
|------|------|
| `52dc73b8` | feat: 添加用户请求/响应报文记录功能 |
| `f19ac2e1` | fix: 流式响应也记录响应报文到日志文件 |
| `cb6c9ae7` | chore: 日志文件上限改为20MB，添加v1.0变更说明 |

## 7. 涉及文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `logger/request_logger.go` | 新增 | 报文记录器核心实现 |
| `controller/relay.go` | 修改 | Relay 函数集成日志记录 |
| `changelogs/CHANGELOG_v1.0.md` | 新增 | 版本变更说明 |
