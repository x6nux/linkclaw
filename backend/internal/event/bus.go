package event

import "sync"

// Handler 事件处理函数
type Handler func(e Event)

// Bus 轻量级进程内发布订阅
type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]Handler
}

var Global = &Bus{handlers: make(map[Type][]Handler)}

// Subscribe 订阅指定类型的事件（返回取消订阅函数）
func (b *Bus) Subscribe(t Type, h Handler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[t] = append(b.handlers[t], h)
	idx := len(b.handlers[t]) - 1

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		handlers := b.handlers[t]
		if idx < len(handlers) {
			handlers[idx] = handlers[len(handlers)-1]
			b.handlers[t] = handlers[:len(handlers)-1]
		}
	}
}

// Publish 发布事件（同步广播给所有订阅者）
func (b *Bus) Publish(e Event) {
	b.mu.RLock()
	handlers := append([]Handler(nil), b.handlers[e.Type]...)
	b.mu.RUnlock()

	for _, h := range handlers {
		h(e)
	}
}
