# OncallGPT Frontend

深色主题的 ChatGPT 风格前端界面，支持普通对话、SSE 流式对话、文件上传与 AI Ops。

## 功能
- 深色主题 + 紫色渐变品牌色
- 消息气泡 + 头像 + 打字指示动画
- Markdown 渲染 + 代码高亮 + 代码块复制按钮 + 行号
- 自动滚动与自适应高度输入框
- 可折叠侧边栏 + 对话历史（localStorage）
- 模式选择器（快速/流式）
- 文件上传按钮
- AI Ops 快捷按钮
- Toast 通知

## 运行
```bash
npm install
npm run dev
```

## 后端接口
```text
POST http://localhost:6871/api/chat
POST http://localhost:6871/api/chat_stream
POST http://localhost:6871/api/upload
POST http://localhost:6871/api/ai_ops
```

## 目录结构
```text
.
├── index.html
├── styles.css
├── app.js
├── package.json
└── README.md
```
