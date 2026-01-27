package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"

	"github.com/hd2yao/oncall-agent/internal/ai/tools"
)

// newReactAgentLambda component initialization function of node 'ReactAgent' in graph 'EinoAgent'
func newReactAgentLambda(ctx context.Context) (lba *compose.Lambda, err error) {
	config := &react.AgentConfig{
		MaxStep:            25, // 最多思考 25 步
		ToolReturnDirectly: map[string]struct{}{},
	}

	// 配置 LLM 模型
	chatModelIns11, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	config.ToolCallingModel = chatModelIns11

	// 配置 MCP 工具
	// 这些工具会在 Agent 思考时被调用
	mcpTool, err := tools.GetLogMcpTool() // 获取日志 MCP 工具
	if err != nil {
		return nil, err
	}
	config.ToolsConfig.Tools = mcpTool
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewPrometheusAlertsQueryTool()) // 告警 Prometheus 查询
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewMysqlCrudTool())             // Mysql 操作
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewGetCurrentTimeTool())        // 获取当前时间
	config.ToolsConfig.Tools = append(config.ToolsConfig.Tools, tools.NewQueryInternalDocsTool())     // RAG 文档检索

	// 创建 ReAct Agent
	ins, err := react.NewAgent(ctx, config)
	if err != nil {
		return nil, err
	}

	// 包装成 Lambda
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		return nil, err
	}
	return lba, nil
}
