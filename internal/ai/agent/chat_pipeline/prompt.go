package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate' in graph 'EinoAgent'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// 1. 配置模板参数
	config := &ChatTemplateConfig{
		FormatType: schema.FString, // 占位符格式，schema.FString 代表使用 {} 作为占位符，本项目中为 {date}、{documents}、{content}
		Templates: []schema.MessagesTemplate{ // 模板列表
			// 模板1: 系统提示词
			schema.SystemMessage(systemPrompt),
			// 模板2: 历史消息占位符
			// MessagesPlaceholder 的意思就是占位，如果 map 里有 "history" 直接将其 value append 进行即可
			// history 是 []*schema.Message 类型实现了 schema.MessagesTemplate 接口类型
			// Templates 本质也是自定义结构体切片，所以就是把两个切片合并成一个大切片
			schema.MessagesPlaceholder("history", false),
			// 模板3: 用户消息模板
			schema.UserMessage("{content}"),
		},
	}

	// 2. 创建 ChatTemplate
	ctp = prompt.FromMessages(config.FormatType, config.Templates...)
	return ctp, nil
}

var systemPrompt = `
# 角色：对话小助手
## 核心能力
- 上下文理解与对话
- 搜索网络获得信息
## 互动指南
- 在回复前，请确保你：
  • 完全理解用户的需求和问题，如果有不清楚的地方，要向用户确认
  • 考虑最合适的解决方案方法
  • 日志主题地域：ap-guangzhou；日志主题id：869830db-a055-4479-963b-3c898d27e755
- 提供帮助时：
  • 语言清晰简洁
  • 适当的时候提供实际例子
  • 有帮助时参考文档
  • 适用时建议改进或下一步操作
- 如果请求超出了你的能力范围：
  • 清晰地说明你的局限性，如果可能的话，建议其他方法
- 如果问题是复合或复杂的，你需要一步步思考，避免直接给出质量不高的回答。
## 输出要求：
  • 易读，结构良好，必要时换行
  • 输出不能包含markdown的语法，输出需要纯文本
## 上下文信息
- 当前日期：{date}
- 相关文档：|-
==== 文档开始 ====
  {documents}
==== 文档结束 ====
`
