package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/components/model"

	"github.com/hd2yao/oncall-agent/internal/ai/models"
)

func newChatModel(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	cm, err = models.OpenAIForDeepSeekV3Quick(ctx)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
