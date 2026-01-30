package chat_pipeline

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"

	retriever2 "github.com/hd2yao/oncall-agent/internal/ai/retriever"
)

// newRetriever component initialization function of node 'MilvusRetriever' in graph 'EinoAgent'
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	milvusRetriever, err := retriever2.NewMilvusRetriever(ctx)
	if err != nil {
		log.Printf("[warn] milvus retriever disabled: %v", err)
		return &safeRetriever{disabled: true}, nil
	}
	return &safeRetriever{inner: milvusRetriever}, nil
}

// safeRetriever degrades to empty results when milvus is unavailable or errors.
type safeRetriever struct {
	inner    retriever.Retriever
	disabled bool
}

func (s *safeRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	if s.disabled || s.inner == nil {
		return []*schema.Document{}, nil
	}
	docs, err := s.inner.Retrieve(ctx, query, opts...)
	if err != nil {
		log.Printf("[warn] milvus retriever error, fallback to empty docs: %v", err)
		return []*schema.Document{}, nil
	}
	return docs, nil
}
