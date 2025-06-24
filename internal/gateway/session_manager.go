package gateway

import (
	"context"
	"errors"
	"sync"
	"time"
	
	"edgesphere/internal/pkg/types"
)

type SessionManager struct {
	sessions  *ConnectionPool
	cache     *SQLiteCache
	heartbeat map[string]*time.Ticker
	mu        sync.RWMutex
}

func NewSessionManager() *SessionManager {
	cache, _ := NewSQLiteCache("/data/offline.db")
	return &SessionManager{
		sessions:  NewConnectionPool(10000),
		cache:     cache,
		heartbeat: make(map[string]*time.Ticker),
	}
}

// 设备连接处理
func (sm *SessionManager) HandleConnection(ctx context.Context, deviceID string, adapter types.ProtocolAdapter) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// 添加到连接池
	conn := &types.DeviceConnection{
		ID:        deviceID,
		Adapter:   adapter,
		LastSeen:  time.Now(),
		Status:    types.Online,
	}
	sm.sessions.Put(deviceID, conn)
	
	// 启动心跳检测
	ticker := time.NewTicker(calculateHeartbeatInterval())
	sm.heartbeat[deviceID] = ticker
	
	go func() {
		for {
			select {
			case <-ticker.C:
				if time.Since(conn.LastSeen) > 30*time.Second {
					sm.handleDisconnection(deviceID)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

// 自适应心跳算法
func calculateHeartbeatInterval() time.Duration {
	base := 15 * time.Second
	// 模拟网络延迟检测 (实际从监控系统获取)
	latency := 150 * time.Millisecond 
	return base + time.Duration(float64(base)*0.1*math.Log(float64(latency)))
}

// 断网处理
func (sm *SessionManager) handleDisconnection(deviceID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if conn, ok := sm.sessions.Get(deviceID); ok {
		conn.Status = types.Offline
		sm.cache.SaveSession(deviceID, conn)
		sm.sessions.Remove(deviceID)
	}
	
	if ticker, ok := sm.heartbeat[deviceID]; ok {
		ticker.Stop()
		delete(sm.heartbeat, deviceID)
	}
}

// 指令下发
func (sm *SessionManager) SendCommand(deviceID string, cmd []byte) error {
	conn, ok := sm.sessions.Get(deviceID)
	if !ok {
		// 设备离线，存入缓存
		return sm.cache.SaveCommand(deviceID, cmd)
	}
	
	return conn.Adapter.Send(cmd)
}

// 故障转移
func (sm *SessionManager) FailoverToBackup(deviceID string) {
	if backupConn := getBackupNode(deviceID); backupConn != nil {
		sm.sessions.Put(deviceID, backupConn)
		// 重发缓存命令
		commands, _ := sm.cache.GetCommands(deviceID)
		for _, cmd := range commands {
			backupConn.Adapter.Send(cmd)
		}
	}
}