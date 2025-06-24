package gateway

import (
	"sync"
	
	"edgesphere/internal/pkg/types"
)

type ConnectionPool struct {
	pool map[string]*types.DeviceConnection
	mu   sync.RWMutex
}

func NewConnectionPool(size int) *ConnectionPool {
	return &ConnectionPool{
		pool: make(map[string]*types.DeviceConnection, size),
	}
}

func (p *ConnectionPool) Get(id string) (*types.DeviceConnection, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	conn, exists := p.pool[id]
	return conn, exists
}

func (p *ConnectionPool) Put(id string, conn *types.DeviceConnection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.pool[id] = conn
}

func (p *ConnectionPool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	delete(p.pool, id)
}

// 零拷贝优化
func (p *ConnectionPool) SendWithZeroCopy(id string, data []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	if conn, ok := p.pool[id]; ok {
		// 使用io_uring系统调用 (Linux)
		_, err := unix.IoUringSubmit(conn.Fd(), data)
		return err
	}
	return errors.New("connection not found")
}