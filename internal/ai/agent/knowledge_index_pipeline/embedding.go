package knowledge_index_pipeline

import (
    "context"

    "github.com/cloudwego/eino-ext/components/embedding/openai"
    "github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
    // TODO Modify component configuration here.
    config := &openai.EmbeddingConfig{}
    eb, err = openai.NewEmbedder(ctx, config)
    if err != nil {
        return nil, err
    }
    return eb, nil
}
