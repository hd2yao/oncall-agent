package summary_agent

import (
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

const summarySystemPrompt = `你是一个对话总结助手。
你的任务是把历史对话压缩成简洁、可延续的上下文记忆。
要求：
1. 保留关键事实、用户偏好、已确认的结论、未完成的事项。
2. 删除寒暄、重复内容和无关细节。
3. 输出使用中文纯文本，控制在 6-10 行以内。
4. 输出应可直接作为后续对话的背景记忆。`

func newSummaryChatTemplate(ctx context.Context) (prompt.ChatTemplate, error) {
	template := prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(summarySystemPrompt),
		schema.UserMessage("{prompt}"),
	)
	return template, nil
}
