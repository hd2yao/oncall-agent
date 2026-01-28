package summary_agent

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// newInputToTemplateLambda converts summary input to prompt variables for ChatTemplate.
func newInputToTemplateLambda(ctx context.Context, input *SummaryInput, opts ...any) (output map[string]any, err error) {
	historyText := formatMessages(input.Dropped)
	promptText := buildSummaryPrompt(input.ExistingSummary, historyText)
	return map[string]any{
		"prompt": promptText,
	}, nil
}

// newOutputToSummaryLambda extracts summary text from model output.
func newOutputToSummaryLambda(ctx context.Context, input *schema.Message, opts ...any) (output string, err error) {
	if input == nil {
		return "", nil
	}
	return strings.TrimSpace(input.Content), nil
}
