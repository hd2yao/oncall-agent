package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"github.com/hd2yao/oncall-agent/internal/controller/chat"
	_ "github.com/hd2yao/oncall-agent/internal/packed"
)

func main() {
	//cmd.Main.Run(gctx.GetInitCtx())

	s := g.Server()
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(ResponseMiddleware)
		group.Bind(chat.NewV1())
	})
	s.SetPort(6871)
	s.Run()
}

func ResponseMiddleware(r *ghttp.Request) {
	r.Middleware.Next()
	var (
		msg string
		res = r.GetHandlerResponse()
		err = r.GetError()
	)
	if err != nil {
		msg = err.Error()
	} else {
		msg = "ok"
	}
	r.Response.WriteJson(Response{
		Message: msg,
		Data:    res,
	})
}

type Response struct {
	Message string      `json:"message" dc:"消息提示"`
	Data    interface{} `json:"data" dc:"执行结果"`
}
