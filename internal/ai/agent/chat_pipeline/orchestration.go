package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildChatAgent(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	const (
		InputToRag      = "InputToRag"
		ChatTemplate    = "ChatTemplate"
		ReactAgent      = "ReactAgent"
		MilvusRetriever = "MilvusRetriever"
		InputToChat     = "InputToChat"
	)

	// 创建一个有向无环图（DAG）
	g := compose.NewGraph[*UserMessage, *schema.Message]()

	// 定义图中的节点
	// 添加 Lambda 节点（执行任意自定义函数）
	// InputToRag、InputToChat: 内部标识符，图内部用于引用节点；Eino 框架代码（AddEdge、编译）
	// compose.WithNodeName(): 显示名称，给开发者看的；调试、日志、UI 展示；由于是自定义函数节点，需要额外设置
	_ = g.AddLambdaNode(InputToRag, compose.InvokableLambdaWithOption(newInputToRagLambda), compose.WithNodeName("UserMessageToRag"))
	_ = g.AddLambdaNode(InputToChat, compose.InvokableLambdaWithOption(newInputToChatLambda), compose.WithNodeName("UserMessageToChat"))

	// 添加 RAG 检索节点（专门用于向量检索）
	milvusRetrieverKeyOfRetriever, err := newRetriever(ctx)
	// milvusRetrieverKeyOfRetriever 实现 retriever.Retriever 接口: 接口输入是 string，输出是 []*schema.Document
	if err != nil {
		return nil, err
	}
	// 注意下面的 output key 设置，把检索结果存入 "documents" 字段，匹配 ChatTemplate 里面 prompt 中用 {documents} 占位符读取输出
	_ = g.AddRetrieverNode(MilvusRetriever, milvusRetrieverKeyOfRetriever, compose.WithOutputKey("documents"))

	// 添加提示词模板节点（专门用于提示词模板）
	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	// 返回 prompt.ChatTemplate 接口: 输入是 map[string]any，输出是 []*schema.Message
	// 使用 AllPredecessor 模式，ChatTemplate 节点会接收多个输入，并自动合并前置节点的输出
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)

	// 创建 ReAct Agent，包装成 compose.Lambda
	// Agent 的输入是 []*schema.Message，输出是 *schema.Message
	reactAgentKeyOfLambda, err := newReactAgentLambda(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda, compose.WithNodeName("ReActAgent"))

	// 定义节点之间的边（执行顺序）
	_ = g.AddEdge(compose.START, InputToRag)     // 开始 → RAG预处理
	_ = g.AddEdge(compose.START, InputToChat)    // 开始 → 聊天预处理
	_ = g.AddEdge(ReactAgent, compose.END)       // ReAct Agent → 结束
	_ = g.AddEdge(InputToRag, MilvusRetriever)   // RAG预处理 → 检索器
	_ = g.AddEdge(MilvusRetriever, ChatTemplate) // 检索结果 → 提示词模板
	_ = g.AddEdge(InputToChat, ChatTemplate)     // 聊天预处理 → 提示词模板
	_ = g.AddEdge(ChatTemplate, ReactAgent)      // 提示词 → ReAct Agent

	// 编译图
	// 使用 AllPredecessor 模式，确保所有前置节点都执行完毕后再执行当前节点
	r, err = g.Compile(ctx, compose.WithGraphName("ChatAgent"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
