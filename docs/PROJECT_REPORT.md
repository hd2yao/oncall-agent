# oncall-agent 项目报告

## 项目背景
oncall-agent 是一个面向运维/值班场景的智能助手服务，目标是在对话中结合内部知识库、日志平台、监控告警与数据库信息，输出可执行的排障建议与分析报告。项目包含后端服务与前端交互界面，支持同步/流式对话、知识库文件上传、AI Ops 诊断与对话记忆压缩。

## 技术选型
- 后端框架：GoFrame v2（HTTP 服务、路由、中间件、配置）
- Agent 编排：CloudWeGo Eino（DAG/Graph、ReAct Agent、工具调用、回调日志）
- 向量数据库：Milvus（Docker Compose 运行，Etcd + MinIO 依赖）
- 向量检索组件：eino-ext milvus retriever/indexer
- Embedding：DashScope 兼容接口（`text-embedding-v4`）
- LLM：OpenAI 兼容接口访问 DeepSeek（`deepseek-v3-1-terminus`）
- 工具集成：
  - MCP（腾讯云 CLS 日志）
  - Prometheus Alerts
  - MySQL CRUD（GORM）
  - 当前时间
  - 内部知识检索（RAG）
- 前端：纯静态页面（HTML/CSS/JS），支持 SSE 流式展示、Markdown 渲染与代码高亮

## 架构设计
### 系统分层
1) API 层（GoFrame）
   - 统一路由：`/api/chat`、`/api/chat_stream`、`/api/upload`、`/api/ai_ops`
   - 统一响应包装：`{ message, data }`（`utility/middleware`）
2) Agent 编排层（Eino DAG）
   - ChatAgent：检索 + 提示词 + ReAct Agent
   - KnowledgeIndexing：文件加载 → 文档分块 → Milvus 索引
   - SummaryAgent：对话记忆压缩
3) 工具能力层
   - MCP（日志查询）、Prometheus、MySQL、时间查询、内部知识检索
4) 存储层
   - Milvus（向量检索）
   - MySQL（工具查询）
   - 本地文件（知识库源文件）

### 关键目录
- `main.go`：HTTP 服务入口，绑定 `/api` 路由
- `internal/controller/chat`：业务接口实现
- `internal/ai/agent/chat_pipeline`：对话 DAG 编排
- `internal/ai/agent/knowledge_index_pipeline`：索引管道
- `internal/ai/agent/summary_agent`：记忆压缩
- `internal/ai/tools`：工具集成
- `utility/mem`：对话记忆
- `manifest/docker`：Milvus/Etcd/MinIO/Attu 编排
- `OncallAgentFrontend`：前端 UI

## Agent 流程解析
### Chat 流程（/api/chat）
1) 接收 `id + question`
2) 读取会话记忆（`utility/mem`），拼接历史与摘要
3) 进入 ChatAgent DAG：
   - InputToRag：提取 query，进入 MilvusRetriever
   - InputToChat：组装 `{content, history, date}`
   - ChatTemplate：系统提示词 + 历史 + 文档上下文
   - ReActAgent：LLM + 工具调用，生成最终回答
4) 写回记忆（用户与系统消息）
5) 超出窗口触发 SummaryAgent 进行历史压缩
6) 返回 `{ answer }`，由中间件包装为 `{ message, data }`

### 流式对话（/api/chat_stream）
与 `/api/chat` 相同的 DAG，但输出以 SSE 方式流式返回，并在结束时回写记忆。

### 文件上传与索引（/api/upload）
1) 保存文件至 `file_dir`
2) FileLoader 读取文件 → MarkdownSplitter 分块
3) MilvusIndexer 写入 `{id, vector, content, metadata}`
4) 若 `_source` 相同，先删除旧记录，避免重复索引

### AI Ops（/api/ai_ops）
基于 Plan-Execute-Replan 结构自动执行：
1) 获取 Prometheus 活跃告警
2) 根据告警名检索内部文档
3) 调用日志/时间工具补充上下文
4) 生成结构化告警分析报告

## 项目启动
### 后端
1) 修改配置：`manifest/config/config.yaml`
   - `ds_*`（LLM 模型）与 `doubao_embedding_model`（Embedding）
   - `file_dir`（上传文件目录）
   - `database`、`cls_msp_url`（如需）
2) 启动服务：`go run main.go`
3) 默认端口：`6871`

### Milvus
```bash
cd manifest/docker
docker compose up -d
```
Attu 管理台：`http://localhost:8001`

### 前端
```bash
cd OncallAgentFrontend
npm install
npm run dev
```

## 面试官视角问题与示例答案
1) **你们的对话链路是如何把 RAG 和工具调用组合起来的？**  
答：ChatAgent 用 DAG 编排，InputToRag 负责向量检索，InputToChat 负责组装上下文，两路合并到 ChatTemplate。ChatTemplate 的输出进入 ReAct Agent，后者可根据提示词决定是否调用工具（日志、告警、MySQL、内部文档等），最终输出答复。

2) **为什么要做 SummaryAgent？它在什么时候触发？**  
答：为了控制上下文窗口与成本，`utility/mem` 会维护 MaxWindowSize。当历史消息超过窗口时，`compressMemory` 会触发 SummaryAgent，总结旧对话并保留关键事实，避免信息丢失。

3) **Milvus 不可用会怎样？**  
答：Chat 侧对 Milvus 检索做了降级处理，失败时返回空文档，LLM 仍可回答。索引侧会直接报错提示，要求修复存储或清理旧 schema。

4) **你们如何保证工具的安全性与可控性？**  
答：工具调用由 ReAct Agent 驱动，且 tool schemas 明确限制输入格式与可调用范围。MySQL 工具要求显式 DSN，日志工具走 MCP 协议，Prometheus 只读 API。

5) **上传文件后如何避免重复索引？**  
答：对 `metadata._source` 相同的记录先查询并删除，再执行索引写入。

6) **为什么前端只收到 “OK”？**  
答：后端统一响应包装 `{ message, data }`，若前端只取 `message` 就只看到 “OK”。需要解析 `data.answer` 等字段。

7) **如果需要对不同模型切换，你们怎么支持？**  
答：通过配置 `manifest/config/config.yaml`，在 `internal/ai/models/open_ai.go` 中读取模型名称与 base_url，可扩展多个模型策略。

8) **如何排查链路中的节点错误？**  
答：Eino 在执行过程中会输出节点路径与错误来源，配合 `log_call_back` 可定位到具体节点（例如 MilvusRetriever 或 ChatTemplate）。

9) **RAG 检索为什么用 BinaryVector？**  
答：这是当前 Milvus schema 与 embedding 转换的实现方式，结合 Hamming 距离检索；若要切换到 float 向量检索，需要调整 embedding 与 schema 类型。

10) **生产部署方式是什么？**  
答：提供 Kustomize 配置与 Makefile 目标，可构建镜像并部署到 Kubernetes（`manifest/deploy`）。

## 后续可优化方向
- **配置治理**：将关键配置（Milvus 地址、模型 key、MCP URL）改为环境变量或 Secret 管理。
- **持久化记忆**：内存记忆改为 Redis/DB，支持多实例与重启恢复。
- **检索策略优化**：TopK 动态策略、混合检索（关键词 + 向量）。
- **RAG 质量提升**：增加 reranker、分块策略优化、去噪/去重处理。
- **工具安全**：对工具调用做权限/审计与速率限制。
- **可观测性**：加入 trace/span、Prometheus 指标，提供完整链路可视化。
- **前端体验**：对话级配置、消息结构化展示、工具调用可视化。
- **错误处理**：统一错误码与用户可读提示，避免模型错误透出。
