---
name: docker-build-export
description: 构建 Docker 镜像并导出为 tar 文件。当用户提到"构建镜像"、"导出镜像"、"打包镜像"、"build image"、"export image"、"docker save" 等关键词时触发此技能。用户通常会指定版本号，如"构建镜像v1.3"。此技能适用于 new-api-ssc 项目的镜像构建与导出流程。
---

# Docker 镜像构建与导出

## 适用场景

用户需要将当前项目构建为 Docker 镜像，并导出为 tar 文件用于部署或分发。

## 参数提取

从用户输入中提取版本号，格式为 `vX.Y`（如 v1.0、v1.2、v2.1）。如果用户未指定版本号，询问用户要使用的版本号。

## 执行步骤

### 1. 确认版本号

提取用户指定的版本号（如 `v1.2`）。如果用户没有指定，主动询问。

### 2. 构建 Docker 镜像

在项目根目录执行构建命令：

```bash
docker build -t new-api-ssc:{version} .
```

- 构建超时设置为 10 分钟（600000ms）
- 构建成功后确认无错误输出

### 3. 导出镜像为 tar 文件

构建成功后，导出镜像：

```bash
docker save new-api-ssc:{version} -o new-api-ssc-{version}.tar
```

- 导出超时设置为 2 分钟（120000ms）

### 4. 验证并报告结果

检查导出的 tar 文件：

```bash
ls -lh new-api-ssc-{version}.tar
```

向用户报告：
- 镜像名称：`new-api-ssc:{version}`
- 导出文件：`new-api-ssc-{version}.tar`
- 文件大小

## 命名规范

| 项目 | 格式 | 示例 |
|------|------|------|
| Docker 镜像标签 | `new-api-ssc:vX.Y` | `new-api-ssc:v1.2` |
| 导出 tar 文件 | `new-api-ssc-vX.Y.tar` | `new-api-ssc-v1.2.tar` |

## 注意事项

- 构建前不需要执行额外操作（Dockerfile 已包含前端构建和后端编译）
- 项目根目录必须包含 `Dockerfile`
- 确保 Docker daemon 正在运行
- 如果构建失败，将错误信息展示给用户并停止
