package plan_execute_replan

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/compose"

	"github.com/hd2yao/oncall-agent/internal/ai/models"
	"github.com/hd2yao/oncall-agent/internal/ai/tools"
)

func NewExecutor(ctx context.Context) (adk.Agent, error) {
	// 腾讯日志文件
	mcpTool, err := tools.GetLogMcpTool()
	if err != nil {
		return nil, err
	}
	toolList := mcpTool
	// 告警查询
	toolList = append(toolList, tools.NewPrometheusAlertsQueryTool())
	// 文档检索
	toolList = append(toolList, tools.NewQueryInternalDocsTool())
	// 时间获取
	toolList = append(toolList, tools.NewGetCurrentTimeTool())

	// 使用 DeepSeek-V3-Quick
	execModel, err := models.OpenAIForDeepSeekV3Quick(ctx)
	if err != nil {
		return nil, err
	}
	return planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model: execModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: toolList,
			},
		},
		MaxIterations: 999999,
	})
}
