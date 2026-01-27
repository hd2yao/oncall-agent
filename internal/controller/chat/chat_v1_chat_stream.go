package chat

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/hd2yao/oncall-agent/api/chat/v1"
	"github.com/hd2yao/oncall-agent/internal/ai/agent/chat_pipeline"
	"github.com/hd2yao/oncall-agent/utility/log_call_back"
	"github.com/hd2yao/oncall-agent/utility/mem"
)

func (c *ControllerV1) ChatStream(ctx context.Context, req *v1.ChatStreamReq) (res *v1.ChatStreamRes, err error) {
	id := req.Id
	msg := req.Question

	ctx = context.WithValue(ctx, "client_id", req.Id)
	// 创建流式对话客户端
	client, err := c.service.Create(ctx, g.RequestFromCtx(ctx))
	if err != nil {
		return nil, err
	}

	userMessage := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   msg,
		History: mem.GetSimpleMemory(id).GetMessagesWithSummary(),
	}

	runner, err := chat_pipeline.BuildChatAgent(ctx)
	// 使用 stream 流式输出模式
	sr, err := runner.Stream(ctx, userMessage, compose.WithCallbacks(log_call_back.LogCallback(nil)))
	if err != nil {
		client.SendToClient("error", err.Error())
		return nil, err
	}
	defer sr.Close()

	var fullResponse strings.Builder

	defer func() {
		completeResponse := fullResponse.String()
		if completeResponse != "" {
			mem.GetSimpleMemory(id).SetMessages(schema.UserMessage(msg))
			mem.GetSimpleMemory(id).SetMessages(schema.SystemMessage(completeResponse))
			compressMemory(ctx, id)
		}
	}()

	for {
		// 从流中读取消息
		chunk, err := sr.Recv()
		if errors.Is(err, io.EOF) {
			client.SendToClient("done", "Stream completed")
			return &v1.ChatStreamRes{}, nil
		}
		if err != nil {
			client.SendToClient("error", err.Error())
			return &v1.ChatStreamRes{}, nil
		}
		fullResponse.WriteString(chunk.Content)
		client.SendToClient("message", chunk.Content)
	}
}
