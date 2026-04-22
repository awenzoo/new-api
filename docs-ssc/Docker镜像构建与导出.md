# Docker 镜像构建与导出

## 构建镜像

在项目根目录执行：

```bash
# 默认构建（linux/amd64）
docker build -t new-api-ssc:v1.1 .

# 指定平台构建（如 arm64）
docker build --platform linux/arm64 -t new-api-ssc:v1.1 .
```

## 导出镜像

```bash
# 导出为 tar 文件
docker save -o new-api-ssc-v1.1.tar new-api-ssc:v1.1

# 压缩导出（推荐，体积更小）
docker save new-api-ssc:v1.1 | gzip > new-api-ssc-v1.1.tar.gz
```

## 导入镜像

在目标机器上执行：

```bash
# 导入 tar 文件
docker load -i new-api-ssc-v1.1.tar

# 导入压缩文件
gunzip -c new-api-ssc-v1.1.tar.gz | docker load
```

## 使用 docker-compose 启动

修改 `docker-compose.yml` 中的镜像名为本地构建的镜像：

```yaml
services:
  new-api:
    image: new-api-ssc:v1.1
    # ...
```

然后启动：

```bash
docker-compose up -d
```
