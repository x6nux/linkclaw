package mcp

import (
	"sync"
	"time"

	"github.com/linkclaw/backend/internal/domain"
)

// Session 表示一个活跃的 MCP SSE 连接
type Session struct {
	ID              string
	Agent           *domain.Agent
	ProtocolVersion string
	ClientInfo      ClientInfo
	ConnectedAt     time.Time
	Initialized     bool

	// SSE 写通道，Handler 通过此发送事件
	send chan string
	done chan struct{}
	once sync.Once
}

func newSession(id string, agent *domain.Agent) *Session {
	return &Session{
		ID:          id,
		Agent:       agent,
		ConnectedAt: time.Now(),
		send:        make(chan string, 64),
		done:        make(chan struct{}),
	}
}

// Send 发送 SSE 事件（非阻塞，发送成功返回 true，通道满返回 false）
func (s *Session) Send(event string) bool {
	select {
	case s.send <- event:
		return true
	default:
		return false
	}
}

// Close 关闭 session
func (s *Session) Close() {
	s.once.Do(func() { close(s.done) })
}

// Done 返回关闭信号
func (s *Session) Done() <-chan struct{} { return s.done }

// SendCh 返回发送通道（只读）
func (s *Session) SendCh() <-chan string { return s.send }

// SessionStore 线程安全的 session 存储
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func newSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]*Session)}
}

func (s *SessionStore) Set(id string, sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = sess
}

func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *SessionStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}
