package summary_agent

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

// BuildSummaryAgent builds a summary agent graph that compresses dialogue history.
func BuildSummaryAgent(ctx context.Context) (compose.Runnable[*SummaryInput, string], error) {
	const (
		InputToTemplate = "InputToTemplate"
		ChatTemplate    = "ChatTemplate"
		ChatModel       = "ChatModel"
		OutputToSummary = "OutputToSummary"
	)

	g := compose.NewGraph[*SummaryInput, string]()

	_ = g.AddLambdaNode(InputToTemplate, compose.InvokableLambdaWithOption(newInputToTemplateLambda), compose.WithNodeName("SummaryInputToTemplate"))

	chatTemplate, err := newSummaryChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplate)

	chatModel, err := newSummaryChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel, chatModel)

	_ = g.AddLambdaNode(OutputToSummary, compose.InvokableLambdaWithOption(newOutputToSummaryLambda), compose.WithNodeName("OutputToSummary"))

	_ = g.AddEdge(compose.START, InputToTemplate)
	_ = g.AddEdge(InputToTemplate, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ChatModel)
	_ = g.AddEdge(ChatModel, OutputToSummary)
	_ = g.AddEdge(OutputToSummary, compose.END)

	r, err := g.Compile(ctx, compose.WithGraphName("SummaryAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, nil
}
