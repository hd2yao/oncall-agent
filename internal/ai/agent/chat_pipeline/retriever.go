package chat_pipeline

import (
	"context"

	"github.com/cloudwego/eino/components/retriever"

	retriever2 "github.com/hd2yao/oncall-agent/internal/ai/retriever"
)

// newRetriever component initialization function of node 'MilvusRetriever' in graph 'EinoAgent'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	return retriever2.NewMilvusRetriever(ctx)
}
