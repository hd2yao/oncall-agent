package plan_execute_replan

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"

	"github.com/hd2yao/oncall-agent/internal/ai/models"
)

func NewPlanner(ctx context.Context) (adk.Agent, error) {
	planModel, err := models.OpenAIForDeepSeekV31Think(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: planModel,
	})
}
