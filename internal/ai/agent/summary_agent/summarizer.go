package summary_agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// SummarizeHistory summarizes dropped conversation history and merges it with any existing summary.
func SummarizeHistory(ctx context.Context, existingSummary string, dropped []*schema.Message) (string, error) {
	if len(dropped) == 0 {
		return existingSummary, nil
	}

	runner, err := BuildSummaryAgent(ctx)
	if err != nil {
		return "", err
	}

	summary, err := runner.Invoke(ctx, &SummaryInput{
		ExistingSummary: existingSummary,
		Dropped:         dropped,
	})
	if err != nil {
		return "", err
	}

	summary = strings.TrimSpace(summary)
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
