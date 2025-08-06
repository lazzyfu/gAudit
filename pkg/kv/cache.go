package kv

import (
	"sync"
)

// 泛型Cache
type Cache[T any] struct {
	value T
}

// 泛型KVCache
type KVCache[T any] struct {
	sync.RWMutex
	Id    string
	Items map[string]*Cache[T]
}

// Put 写入
func (c *KVCache[T]) Put(key string, value T) {
	c.Lock()
	defer c.Unlock()
	c.Items[key] = &Cache[T]{value: value}
}

// Get 查询
func (c *KVCache[T]) Get(key string) (T, bool) {
	c.RLock()
	defer c.RUnlock()
	item, ok := c.Items[key]
	if !ok {
		var zero T
		return zero, false
	}
	return item.value, true
}

// Delete 删除
func (c *KVCache[T]) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.Items, key)
}

// NewKVCache 新建缓存
func NewKVCache[T any](Id string) *KVCache[T] {
	return &KVCache[T]{Id: Id, Items: make(map[string]*Cache[T])}
}
