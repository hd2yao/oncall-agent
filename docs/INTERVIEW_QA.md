# oncall-agent 面试 Q&A

## 1. Chat 的主链路是如何把 RAG 和工具调用结合的？
答：ChatAgent 通过 DAG 编排。InputToRag 提取 query 进入 MilvusRetriever；InputToChat 组装 `{content, history, date}`；两路在 ChatTemplate 合并，再交给 ReAct Agent 决定是否调用工具并产出回答。

## 2. 为什么需要 SummaryAgent？触发条件是什么？
答：为控制上下文长度与成本，`utility/mem` 维护窗口大小。超过阈值时触发 SummaryAgent，把旧对话压缩为摘要，避免信息丢失。

## 3. Milvus 不可用会怎样？
答：检索层降级为空文档，聊天链路仍可用；索引链路仍会提示错误，避免无声失败。

## 4. 为什么用 BinaryVector + HAMMING？
答：当前 embedding 输出与 BinaryVector 适配实现路径简单；若需要 float 向量检索，可替换 schema 与检索参数。

## 5. AI Ops 报告如何生成？
答：使用 Plan-Execute-Replan：先取 Prometheus 活跃告警，再按告警名检索内部文档，必要时调用日志与时间工具补全上下文，最后生成报告。

## 6. 为什么前端只显示 “OK”？
答：后端响应统一包装 `{message, data}`，如果前端只解析 `message` 就会看到 “OK”，实际答案在 `data.answer`。

## 7. 工具调用如何保证安全与可控？
答：工具具备明确的输入 schema；后续可加权限校验、审计与速率限制，控制高风险操作。

## 8. 上传文档如何避免重复索引？
答：按 `metadata._source` 进行去重，存在则先删除旧记录再重建索引。

## 9. 如何定位链路错误？
答：Eino DAG 会输出 node path，结合回调日志能快速定位到检索、模板或模型节点。

