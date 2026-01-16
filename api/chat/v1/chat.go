package v1

import "github.com/gogf/gf/v2/frame/g"

type ChatReq struct {
	g.Meta   `path:"/chat" method:"get" summary:"对话"`
	Id       string
	Question string
}
type ChatRes struct {
	Answer string `json:"answer"`
}
