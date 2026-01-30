# oncall-agent 运维手册

## 运行依赖
- Go 1.24+
- Milvus（含 Etcd + MinIO）
- Node.js（仅前端）
- 可选：MySQL、Prometheus、MCP 日志服务

## 配置说明
配置文件：`manifest/config/config.yaml`
关键字段：
- `ds_think_chat_model` / `ds_quick_chat_model`：LLM 访问参数
- `doubao_embedding_model`：Embedding 参数
- `file_dir`：上传文件保存目录
- `database.default.link`：MySQL DSN
- `cls_msp_url`：MCP 日志服务地址

## 启动步骤
### 1) 启动 Milvus
```bash
cd manifest/docker
docker compose up -d
```
Attu 地址：`http://localhost:8001`

### 2) 启动后端
```bash
go run main.go
```
默认端口：6871

### 3) 启动前端
```bash
cd OncallAgentFrontend
npm install
npm run dev
```

## 运维检查
- 后端健康：请求 `/api/chat`（需要正常 model key）
- Milvus 健康：`curl http://localhost:9092/healthz`
- Attu UI：`http://localhost:8001`

## 故障排查
### 1) Milvus schema 不匹配
现象：`extra output fields [content metadata] ...`
解决：
- 通过 Attu 删除 `agent` 数据库中的 `biz` collection
- 或 `docker compose down -v` 清空卷后重启

### 2) Embedding/LLM 报错
- 检查 config 中的 `api_key` 与 `base_url`
- 确保模型名称有效

### 3) 文档检索无结果
- 确认已上传文件且索引成功
- 查看后端日志中索引构建输出

### 4) SSE 无响应
- 检查浏览器是否被代理或缓存
- 确认 `/api/chat_stream` 返回 headers 为 `text/event-stream`

