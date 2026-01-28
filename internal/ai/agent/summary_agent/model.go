package summary_agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"

	"github.com/hd2yao/oncall-agent/internal/ai/models"
)

func newSummaryChatModel(ctx context.Context) (model.BaseChatModel, error) {
	cm, err := models.OpenAIForDeepSeekV3Quick(ctx)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
