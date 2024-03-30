package merkledag

import (
	"crypto/sha256"
	"hash"
	"sync"
)

type HashPool interface {
	Get() hash.Hash
}
type SimpleHashPool struct {
	pool sync.Pool
}

// NewSimpleHashPool 创建一个新的SimpleHashPool实例
func NewSimpleHashPool() *SimpleHashPool {
	return &SimpleHashPool{
		pool: sync.Pool{
			New: func() interface{} {
				return sha256.New() // 这里我们假设总是返回sha256的实例
			},
		},
	}
}

// Get 从池中获取一个hash.Hash实例
func (h *SimpleHashPool) Get() hash.Hash {
	return h.pool.Get().(hash.Hash)
}

// Put 将hash.Hash实例放回池中，以便后续重用
func (h *SimpleHashPool) Put(hasher hash.Hash) {
	h.pool.Put(hasher)
}
