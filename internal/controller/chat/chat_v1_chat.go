package chat

import (
	"context"

	"github.com/hd2yao/oncall-agent/api/chat/v1"
)

func (c *ControllerV1) Chat(ctx context.Context, req *v1.ChatReq) (res *v1.ChatRes, err error) {
	return &v1.ChatRes{Answer: "chat demo"}, nil
}
