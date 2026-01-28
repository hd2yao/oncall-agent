package summary_agent

import "github.com/cloudwego/eino/schema"

type SummaryInput struct {
	ExistingSummary string            `json:"existing_summary"`
	Dropped         []*schema.Message `json:"dropped"`
}
