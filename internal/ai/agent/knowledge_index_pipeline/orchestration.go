package knowledge_index_pipeline

import (
    "context"

    "github.com/cloudwego/eino/compose"
)

func BuildKnowledgeIndexing(ctx context.Context) (r compose.Runnable[any, any], err error) {
    const (
        FileLoader       = "FileLoader"
        MarkdownSplitter = "MarkdownSplitter"
        MilvusIndexer    = "MilvusIndexer"
    )
    g := compose.NewGraph[any, any]()
    fileLoaderKeyOfLoader, err := newLoader(ctx)
    if err != nil {
        return nil, err
    }
    _ = g.AddLoaderNode(FileLoader, fileLoaderKeyOfLoader)
    markdownSplitterKeyOfDocumentTransformer, err := newDocumentTransformer(ctx)
    if err != nil {
        return nil, err
    }
    _ = g.AddDocumentTransformerNode(MarkdownSplitter, markdownSplitterKeyOfDocumentTransformer)
    milvusIndexerKeyOfIndexer, err := newIndexer(ctx)
    if err != nil {
        return nil, err
    }
    _ = g.AddIndexerNode(MilvusIndexer, milvusIndexerKeyOfIndexer)
    _ = g.AddEdge(compose.START, FileLoader)
    _ = g.AddEdge(MilvusIndexer, compose.END)
    _ = g.AddEdge(FileLoader, MarkdownSplitter)
    _ = g.AddEdge(MarkdownSplitter, MilvusIndexer)
    r, err = g.Compile(ctx, compose.WithGraphName("KnowledgeIndexing"))
    if err != nil {
        return nil, err
    }
    return r, err
}
