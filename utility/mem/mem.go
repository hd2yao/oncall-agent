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
			MaxWindowSize: 6,
		}
		SimpleMemoryMap[id] = newMem
		return newMem
	}
}

type SimpleMemory struct {
	ID            string            `json:"id"`
	Message       []*schema.Message `json:"message"`
	MaxWindowSize int
	mu            sync.Mutex
}

func (s *SimpleMemory) SetMessages(msg *schema.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Message = append(s.Message, msg)
	if len(s.Message) > s.MaxWindowSize {
		// 确保成对丢弃消息，保持对话配对关系
		// 计算需要丢弃的消息数量（必须是偶数）
		excess := len(s.Message) - s.MaxWindowSize
		if excess%2 != 0 {
			excess++
		}
		// 丢弃前面的消息，保持对话配对
		s.Message = s.Message[excess:]
	}
}

func (s *SimpleMemory) GetMessages() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.Message
}
