# oncall-agent 用户手册

## 适用人群
- 需要使用对话助手进行排障、告警分析、知识查询的用户
- 需要上传内部文档进行 RAG 检索的用户

## 功能概览
- 普通对话（/api/chat）
- 流式对话（/api/chat_stream）
- 文件上传与知识库构建（/api/upload）
- AI Ops 告警分析（/api/ai_ops）

## Web 前端使用
前端位于 `OncallAgentFrontend`，提供对话、流式展示与文件上传入口。

常用流程：
1) 打开前端页面
2) 输入问题并发送
3) 需要知识库时点击“文件上传”上传文档
4) 选择“快速/流式”模式进行对话

## 接口使用（示例）
### 1) 对话（同步）
POST /api/chat
```json
{
  "id": "session-id",
  "question": "什么是 AGI"
}
```

响应：
```json
{
  "message": "OK",
  "data": {
    "answer": "..."
  }
}
```

### 2) 对话（流式）
POST /api/chat_stream
- SSE 返回 `event: message` + `data: ...` 形式

### 3) 文件上传
POST /api/upload (multipart/form-data)
字段：`file`

### 4) AI Ops
POST /api/ai_ops
```json
{}
```

## 常见问题
1) 返回只有“OK”
- 说明你只解析了 `message` 字段，真实内容在 `data.answer`。

2) 无法检索到文档
- 确保已上传文档并完成索引
- Milvus 必须正常启动

3) 检索报错但聊天仍可用
- 当 Milvus 不可用时系统会降级为空检索，属于正常行为

