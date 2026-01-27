package summary_agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/hd2yao/oncall-agent/internal/ai/models"
)

const summarySystemPrompt = `你是一个对话总结助手。
你的任务是把历史对话压缩成简洁、可延续的上下文记忆。
要求：
1. 保留关键事实、用户偏好、已确认的结论、未完成的事项。
2. 删除寒暄、重复内容和无关细节。
3. 输出使用中文纯文本，控制在 6-10 行以内。
4. 输出应可直接作为后续对话的背景记忆。`

// SummarizeHistory summarizes dropped conversation history and merges it with any existing summary.
func SummarizeHistory(ctx context.Context, existingSummary string, dropped []*schema.Message) (string, error) {
	if len(dropped) == 0 {
		return existingSummary, nil
	}

	chatModel, err := models.OpenAIForDeepSeekV3Quick(ctx)
	if err != nil {
		return "", err
	}

	historyText := formatMessages(dropped)
	userPrompt := buildSummaryPrompt(existingSummary, historyText)

	resp, err := chatModel.Generate(ctx, []*schema.Message{
		schema.SystemMessage(summarySystemPrompt),
		schema.UserMessage(userPrompt),
	})
	if err != nil {
		return "", err
	}

	summary := strings.TrimSpace(resp.Content)
	if summary == "" {
		return existingSummary, nil
	}
	return summary, nil
}

func buildSummaryPrompt(existingSummary, historyText string) string {
	if strings.TrimSpace(existingSummary) == "" {
		return fmt.Sprintf("请总结以下历史对话：\n\n%s", historyText)
	}
	return fmt.Sprintf(
		"已有总结：\n%s\n\n请将下面新增的对话历史融合进总结，输出新的完整总结：\n\n%s",
		existingSummary,
		historyText,
	)
}

func formatMessages(messages []*schema.Message) string {
	var b strings.Builder
	for i, msg := range messages {
		role := string(msg.Role)
		if role == "" {
			role = "unknown"
		}
		_, _ = fmt.Fprintf(&b, "%d. %s: %s\n", i+1, role, strings.TrimSpace(msg.Content))
	}
	return b.String()
}
