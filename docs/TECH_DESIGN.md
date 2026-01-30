# oncall-agent 技术设计

## 总体结构
系统分为 API 层、Agent 编排层、工具层与存储层。

### API 层
入口：`main.go`
- `/api/chat`：同步对话
- `/api/chat_stream`：SSE 流式对话
- `/api/upload`：文件上传 + 索引
- `/api/ai_ops`：AI Ops 告警分析

中间件：
- CORS 处理
- 统一响应包装 `{ message, data }`

### Agent 编排层
使用 CloudWeGo Eino 进行 DAG 编排：
- ChatAgent（对话）
- KnowledgeIndexing（索引）
- SummaryAgent（对话记忆压缩）

### 工具层
集成多工具：
- MCP 日志（腾讯云 CLS）
- Prometheus Alerts
- MySQL CRUD
- 当前时间
- 内部知识检索（RAG）

### 存储层
- Milvus：向量检索（collection: `biz`，db: `agent`）
- 本地文件：知识库源文件
- 内存：对话记忆（可扩展为 Redis/DB）

## ChatAgent 流程
1) 输入：`UserMessage { id, query, history }`
2) InputToRag：提取 query → MilvusRetriever
3) InputToChat：组装 `{content, history, date}`
4) ChatTemplate：系统提示词 + 历史 + 文档上下文
5) ReAct Agent：LLM + 工具调用
6) 输出：`schema.Message`

## KnowledgeIndexing 流程
1) FileLoader 读取文件
2) MarkdownSplitter 按标题分块
3) MilvusIndexer 写入向量与字段

## SummaryAgent 流程
1) 判断对话窗口超限
2) 取出过旧消息
3) 调用摘要 Agent 生成总结
4) 总结作为系统消息前置

## 关键实现要点
- ChatTemplate 使用 `{documents}` 注入 RAG 结果
- `utility/mem` 管理会话窗口与摘要
- Milvus 不可用时降级为空检索（不影响对话主链路）

