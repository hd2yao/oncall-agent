package chat

import (
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"github.com/hd2yao/oncall-agent/api/chat/v1"
	"github.com/hd2yao/oncall-agent/internal/ai/agent/chat_pipeline"
	"github.com/hd2yao/oncall-agent/utility/log_call_back"
	"github.com/hd2yao/oncall-agent/utility/mem"
)

func (c *ControllerV1) Chat(ctx context.Context, req *v1.ChatReq) (res *v1.ChatRes, err error) {
	id := req.Id
	msg := req.Question
	// 1. 构造结构体
	userMessage := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   msg,
		History: mem.GetSimpleMemory(id).GetMessagesWithSummary(),
	}

	// 2. 创建对话 Agent 的执行器
	runner, err := chat_pipeline.BuildChatAgent(ctx)
	if err != nil {
		return nil, err
	}

	// 3. 执行
	out, err := runner.Invoke(ctx, userMessage, compose.WithCallbacks(log_call_back.LogCallback(nil)))
	if err != nil {
		return nil, err
	}

	// 4. 将本轮对话存入系统
	mem.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
	mem.GetSimpleMemory(id).SetMessages(schema.SystemMessage(out.Content))
	compressMemory(ctx, id)

	// 5. 返回消息
	res = &v1.ChatRes{
		Answer: out.Content,
	}
	return res, nil
}
