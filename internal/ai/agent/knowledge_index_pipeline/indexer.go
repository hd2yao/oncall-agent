package knowledge_index_pipeline

import (
	"context"

	"github.com/cloudwego/eino/components/indexer"

	indexer2 "github.com/hd2yao/oncall-agent/internal/ai/indexer"
)

// newIndexer component initialization function of node 'MilvusIndexer' in graph 'KnowledgeIndexing'
func newIndexer(ctx context.Context) (idr indexer.Indexer, err error) {
	return indexer2.NewMilvusIndexer(ctx)
}
