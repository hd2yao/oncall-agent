package chat

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	summary_agent "github.com/hd2yao/oncall-agent/internal/ai/agent/summary_agent"
	"github.com/hd2yao/oncall-agent/utility/mem"
)

// compressMemory summarizes old dialogue turns when the memory window exceeds its limit.
func compressMemory(ctx context.Context, id string) {
	memory := mem.GetSimpleMemory(id)
	dropped := memory.ExtractExcessPairs()
	if len(dropped) == 0 {
		return
	}

	summary, err := summary_agent.SummarizeHistory(ctx, memory.GetSummary(), dropped)
	if err != nil {
		g.Log().Errorf(ctx, "summarize history failed for id=%s: %+v", id, err)
		// 避免总结失败导致上下文丢失
		memory.PrependMessages(dropped)
		return
	}
	memory.SetSummary(summary)
}
