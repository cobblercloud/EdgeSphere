package tests

import (
	"testing"
	"time"
	
	"edgesphere/internal/gateway"
)

func TestConnectionScaling(t *testing.T) {
	sm := gateway.NewSessionManager()
	
	// 模拟10K设备连接
	for i := 0; i < 10000; i++ {
		deviceID := fmt.Sprintf("device-%d", i)
		adapter := mqtt.NewMockAdapter()
		go sm.HandleConnection(context.Background(), deviceID, adapter)
	}
	
	time.Sleep(2 * time.Second)
	
	// 验证连接数
	if count := sm.sessions.Count(); count != 10000 {
		t.Errorf("Expected 10000 connections, got %d", count)
	}
	
	// 模拟指令下发
	start := time.Now()
	for i := 0; i < 50000; i++ {
		deviceID := fmt.Sprintf("device-%d", rand.Intn(10000))
		sm.SendCommand(deviceID, []byte("test-command"))
	}
	duration := time.Since(start)
	
	t.Logf("50K commands processed in %v (%.0f TPS)", 
		duration, 50000/duration.Seconds())
	
	if duration > 5*time.Second {
		t.Error("Performance below requirement")
	}
}