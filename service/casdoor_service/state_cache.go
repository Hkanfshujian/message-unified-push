package casdoor_service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// StateInfo 登录状态信息
type StateInfo struct {
	Nonce     string
	ClientIP  string
	CreatedAt time.Time
}

// StateCache 登录状态缓存（内存实现）
type StateCache struct {
	mu     sync.RWMutex
	states map[string]*StateInfo
	ttl    time.Duration
}

var (
	globalStateCache *StateCache
	once             sync.Once
)

// GetStateCache 获取全局状态缓存实例
func GetStateCache() *StateCache {
	once.Do(func() {
		globalStateCache = &StateCache{
			states: make(map[string]*StateInfo),
			ttl:    5 * time.Minute,
		}
		// 启动定期清理过期状态的协程
		go globalStateCache.cleanup()
	})
	return globalStateCache
}

// GenerateState 生成新的登录状态
func (c *StateCache) GenerateState(clientIP string) (state string, nonce string, err error) {
	// 生成随机 state
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", err
	}
	state = hex.EncodeToString(stateBytes)

	// 生成随机 nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", err
	}
	nonce = hex.EncodeToString(nonceBytes)

	// 存储状态
	c.mu.Lock()
	c.states[state] = &StateInfo{
		Nonce:     nonce,
		ClientIP:  clientIP,
		CreatedAt: time.Now(),
	}
	c.mu.Unlock()

	return state, nonce, nil
}

// ConsumeState 验证并消费状态（一次性使用）
// 返回状态信息和是否有效
func (c *StateCache) ConsumeState(state string) (*StateInfo, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	info, exists := c.states[state]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Since(info.CreatedAt) > c.ttl {
		delete(c.states, state)
		return nil, false
	}

	// 立即删除，确保一次性使用
	delete(c.states, state)
	return info, true
}

// cleanup 定期清理过期状态
func (c *StateCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for state, info := range c.states {
			if now.Sub(info.CreatedAt) > c.ttl {
				delete(c.states, state)
			}
		}
		c.mu.Unlock()
	}
}

// Count 获取当前缓存数量（用于调试）
func (c *StateCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.states)
}
