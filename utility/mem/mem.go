package mem

import (
	"sync"

	"github.com/cloudwego/eino/schema"
)

var SimpleMemoryMap = make(map[string]*SimpleMemory)
var mu sync.Mutex

func GetSimpleMemory(id string) *SimpleMemory {
	mu.Lock()
	defer mu.Unlock()
	// 如果存在就返回，不存在就创建
	if mem, ok := SimpleMemoryMap[id]; ok {
		return mem
	} else {
		newMem := &SimpleMemory{
			ID:            id,
			Message:       []*schema.Message{},
			Summary:       "",
			MaxWindowSize: 6,
		}
		SimpleMemoryMap[id] = newMem
		return newMem
	}
}

type SimpleMemory struct {
	ID            string            `json:"id"`
	Message       []*schema.Message `json:"message"`
	Summary       string            `json:"summary"`
	MaxWindowSize int
	mu            sync.Mutex
}

func (s *SimpleMemory) SetMessages(msg *schema.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Message = append(s.Message, msg)
}

func (s *SimpleMemory) GetMessages() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Message
}

// GetMessagesWithSummary returns the summary (if any) prepended to the current message window.
func (s *SimpleMemory) GetMessagesWithSummary() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Summary == "" {
		return append([]*schema.Message(nil), s.Message...)
	}

	summaryMsg := schema.SystemMessage("以下是历史对话的总结，请在回答时考虑这些背景信息：" + s.Summary)
	history := make([]*schema.Message, 0, len(s.Message)+1)
	history = append(history, summaryMsg)
	history = append(history, s.Message...)
	return history
}

// ExtractExcessPairs removes the oldest messages beyond the window size, keeping message pairs aligned.
func (s *SimpleMemory) ExtractExcessPairs() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Message) <= s.MaxWindowSize {
		return nil
	}

	excess := len(s.Message) - s.MaxWindowSize
	if excess%2 != 0 {
		excess++
	}
	if excess > len(s.Message) {
		excess = len(s.Message)
	}

	dropped := append([]*schema.Message(nil), s.Message[:excess]...)
	s.Message = s.Message[excess:]
	return dropped
}

// PrependMessages restores dropped messages to the front of the memory window.
func (s *SimpleMemory) PrependMessages(dropped []*schema.Message) {
	if len(dropped) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	restored := make([]*schema.Message, 0, len(dropped)+len(s.Message))
	restored = append(restored, dropped...)
	restored = append(restored, s.Message...)
	s.Message = restored
}

func (s *SimpleMemory) GetSummary() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Summary
}

func (s *SimpleMemory) SetSummary(summary string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Summary = summary
}
