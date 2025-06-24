package utils

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type ConsistentHash struct {
	circle      map[uint32]string
	sortedKeys  []uint32
	replicas    int
	virtualNodes map[string]int
	mu          sync.RWMutex
}

func NewConsistentHash(replicas int) *ConsistentHash {
	return &ConsistentHash{
		circle:      make(map[uint32]string),
		replicas:    replicas,
		virtualNodes: make(map[string]int),
	}
}

func (c *ConsistentHash) AddNode(node string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.virtualNodes[node]; exists {
		return
	}
	
	c.virtualNodes[node] = c.replicas
	for i := 0; i < c.replicas; i++ {
		virtualKey := node + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(virtualKey))
		c.circle[hash] = node
		c.sortedKeys = append(c.sortedKeys, hash)
	}
	sort.Slice(c.sortedKeys, func(i, j int) bool {
		return c.sortedKeys[i] < c.sortedKeys[j]
	})
}

func (c *ConsistentHash) RemoveNode(node string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	if _, exists := c.virtualNodes[node]; !exists {
		return
	}
	
	delete(c.virtualNodes, node)
	for i := 0; i < c.replicas; i++ {
		virtualKey := node + "#" + strconv.Itoa(i)
		hash := crc32.ChecksumIEEE([]byte(virtualKey))
		delete(c.circle, hash)
		
		// 从排序列表中移除
		idx := sort.Search(len(c.sortedKeys), func(i int) bool {
			return c.sortedKeys[i] >= hash
		})
		if idx < len(c.sortedKeys) && c.sortedKeys[idx] == hash {
			c.sortedKeys = append(c.sortedKeys[:idx], c.sortedKeys[idx+1:]...)
		}
	}
}

func (c *ConsistentHash) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if len(c.circle) == 0 {
		return ""
	}
	
	hash := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(c.sortedKeys), func(i int) bool {
		return c.sortedKeys[i] >= hash
	})
	
	if idx >= len(c.sortedKeys) {
		idx = 0
	}
	
	return c.circle[c.sortedKeys[idx]]
}

func (c *ConsistentHash) GetNodes(key string, count int) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	if len(c.circle) == 0 {
		return nil
	}
	
	hash := crc32.ChecksumIEEE([]byte(key))
	idx := sort.Search(len(c.sortedKeys), func(i int) bool {
		return c.sortedKeys[i] >= hash
	})
	
	nodes := make(map[string]struct{})
	result := make([]string, 0, count)
	
	// 顺时针查找节点
	for i := 0; i < len(c.sortedKeys) && len(nodes) < count; i++ {
		nodeIdx := (idx + i) % len(c.sortedKeys)
		node := c.circle[c.sortedKeys[nodeIdx]]
		if _, exists := nodes[node]; !exists {
			nodes[node] = struct{}{}
			result = append(result, node)
		}
	}
	
	return result
}