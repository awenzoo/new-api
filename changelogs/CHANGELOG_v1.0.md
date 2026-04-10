# v1.0 变更说明

## 新增功能

### 用户请求/响应报文记录

为系统添加了用户大模型请求和响应报文的完整记录功能，便于问题排查和审计追溯。

**核心变更：**

| 文件 | 变更类型 | 说明 |
|------|---------|------|
| `logger/request_logger.go` | 新增 | 请求/响应报文记录器，包含日志轮转机制 |
| `controller/relay.go` | 修改 | 在 Relay 函数中集成报文捕获和日志记录 |

**功能特性：**

- 按用户 ID 分目录存储日志文件
- 支持流式（SSE）和非流式响应的完整报文记录
- 异步写入，不影响请求响应性能
- 日志自动轮转：单文件 20MB 上限，每用户最多保留 10 个文件
- 超出数量限制时自动删除最旧的日志文件

**启用方式：**

通过环境变量 `REQUEST_LOG_ENABLED=true` 启用：

```bash
docker run -d --name new-api -p 3000:3000 \
  -e REQUEST_LOG_ENABLED=true \
  -v /data/new-api:/data \
  new-api:v1.0
```

**日志文件路径：**

```
{log-dir}/request_logs/user_{userId}/request_20260411_120000.log
```

默认 `log-dir` 为 `./logs`，容器内完整路径为 `/data/logs/request_logs/`。

**日志格式示例：**

```
================================================================================
[2026/04/11 - 12:00:00] userId=16 path=/pg/chat/completions status=200
================================================================================

>>> REQUEST BODY >>>
{
  "model": "glm-5",
  "messages": [...],
  "stream": true
}

<<< RESPONSE BODY <<>
data: {"choices":[{"delta":{"content":"你好"}}]}
data: [DONE]
```

**涉及提交：**

- `52dc73b8` feat: 添加用户请求/响应报文记录功能
- `f19ac2e1` fix: 流式响应也记录响应报文到日志文件
