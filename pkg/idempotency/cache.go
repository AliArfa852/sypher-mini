package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// Cache provides session deduplication for inbound messages.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

type cacheEntry struct {
	taskID   string
	result   string
	inserted time.Time
}

// New creates an idempotency cache with the given TTL (e.g. 60s).
func New(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func hash(sessionKey, content string) string {
	h := sha256.Sum256([]byte(sessionKey + "|" + content))
	return hex.EncodeToString(h[:16])
}

// Get returns cached result if (sessionKey, content) was seen within TTL.
func (c *Cache) Get(sessionKey, content string) (taskID, result string, ok bool) {
	key := hash(sessionKey, content)
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Since(e.inserted) > c.ttl {
		return "", "", false
	}
	return e.taskID, e.result, true
}

// Set stores a result for (sessionKey, content).
func (c *Cache) Set(sessionKey, content, taskID, result string) {
	key := hash(sessionKey, content)
	c.mu.Lock()
	c.entries[key] = cacheEntry{taskID: taskID, result: result, inserted: time.Now()}
	c.mu.Unlock()
}

// Cleanup removes expired entries. Call periodically.
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.entries {
		if now.Sub(e.inserted) > c.ttl {
			delete(c.entries, k)
		}
	}
}
