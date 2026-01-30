package chat

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"

	v1 "github.com/hd2yao/oncall-agent/api/chat/v1"
	"github.com/hd2yao/oncall-agent/internal/ai/agent/knowledge_index_pipeline"
	loader2 "github.com/hd2yao/oncall-agent/internal/ai/loader"
	"github.com/hd2yao/oncall-agent/utility/client"
	"github.com/hd2yao/oncall-agent/utility/common"
	"github.com/hd2yao/oncall-agent/utility/log_call_back"
)

func (c *ControllerV1) FileUpload(ctx context.Context, req *v1.FileUploadReq) (res *v1.FileUploadRes, err error) {
	// 1. 从请求中获取上传文件
	r := g.RequestFromCtx(ctx)
	uploadFile := r.GetUploadFile("file")
	if uploadFile == nil {
		return nil, gerror.New("请上传文件")
	}

	// 2. 确保保存目录存在
	if !gfile.Exists(common.FileDir) {
		// 不存在则创建
		if err := gfile.Mkdir(common.FileDir); err != nil {
			return nil, gerror.Wrapf(err, "创建目录失败：%s", common.FileDir)
		}
	}

	// 3. 保存文件
	// 获取原始文件名
	newFileName := uploadFile.Filename
	// 完整的保存路径
	savePath := filepath.Join(common.FileDir)
	// 保存文件
	_, err = uploadFile.Save(savePath, false)
	if err != nil {
		return nil, gerror.Wrapf(err, "保存文件失败")
	}
	// 获取保存后的文件信息
	fileInfo, err := os.Stat(savePath)
	if err != nil {
		return nil, gerror.Wrapf(err, "获取文件信息失败")
	}
	res = &v1.FileUploadRes{
		FileName: newFileName,
		FilePath: savePath,
		FileSize: fileInfo.Size(),
	}

	// 4. 构建知识库索引
	err = buildIntoIndex(ctx, common.FileDir+"/"+newFileName)
	if err != nil {
		return nil, gerror.Wrapf(err, "构建知识库失败")
	}

	return res, nil
}

func buildIntoIndex(ctx context.Context, path string) error {
	// 1. 构建知识索引管道（Start → FileLoader → Splitter → Indexer → End）
	r, err := knowledge_index_pipeline.BuildKnowledgeIndexing(ctx)
	if err != nil {
		return err
	}

	// 2. 删除 biz 数据 metadata 中 _source 一样的数据
	// 创建文件加载器
	loader, err := loader2.NewFileLoader(ctx)
	if err != nil {
		return err
	}

	// 加载文件到内存，得到 document[]
	docs, err := loader.Load(ctx, document.Source{URI: path})
	if err != nil {
		return err
	}
	if len(docs) == 0 {
		return fmt.Errorf("no documents loaded from path: %s", path)
	}
	cli, err := client.NewMilvusClient(ctx)
	if err != nil {
		return err
	}

	// 查询所有 metadata 中 _source 一样的数据并删除
	// 例如：{"_file_name":"告警处理手册.md","_extension":".md","title":"与下游对账发现差异","_source":"docs/告警处理手册.md"}
	expr := fmt.Sprintf(`metadata["_source"] == "%s"`, docs[0].MetaData["_source"])
	queryResult, err := cli.Query(ctx, common.MilvusCollectionName, []string{}, expr, []string{"id"})
	if err != nil {
		return err
	} else if len(queryResult) > 0 {
		// 提取所有需要删除的 id
		var idsToDelete []string
		for _, column := range queryResult {
			if column.Name() == "id" {
				for i := 0; i < column.Len(); i++ {
					id, err := column.GetAsString(i)
					if err == nil {
						idsToDelete = append(idsToDelete, id)
					}
				}
			}
		}

		// 删除旧数据(避免重复索引)
		if len(idsToDelete) > 0 {
			deleteExpr := fmt.Sprintf(`id in ["%s"]`, strings.Join(idsToDelete, `","`))
			err = cli.Delete(ctx, common.MilvusCollectionName, "", deleteExpr)
			if err != nil {
				fmt.Printf("[warn] delete existing data failed: %v\n", err)
			} else {
				fmt.Printf("[info] deleted %d existing records with _source: %s\n", len(idsToDelete), docs[0].MetaData["_source"])
			}

		}
	}

	// 调用索引通道，重新构建并存储到 Milvus
	ids, err := r.Invoke(ctx, document.Source{URI: path}, compose.WithCallbacks(log_call_back.LogCallback(nil)))
	if err != nil {
		return fmt.Errorf("invoke index graph failed: %w", err)
	}
	fmt.Printf("[done] indexing file: %s, len of parts: %d\n", path, len(ids))
	return nil
}
