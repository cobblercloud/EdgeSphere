package tests

import (
	"net"
	"testing"
	"time"
	
	"edgesphere/internal/gateway"
)

func TestFailoverRecovery(t *testing.T) {
	sm := gateway.NewSessionManager()
	
	// 创建主连接
	mainConn := net.Pipe()
	adapter := mqtt.NewMQTTAdapter(mainConn)
	sm.HandleConnection(context.Background(), "device-001", adapter)
	
	// 发送命令
	sm.SendCommand("device-001", []byte("command1"))
	
	// 模拟主节点故障
	mainConn.Close()
	time.Sleep(500 * time.Millisecond) // 等待心跳检测
	
	// 触发故障转移
	sm.FailoverToBackup("device-001")
	
	// 验证命令恢复
	backupConn := getBackupConnection("device-001")
	if backupConn == nil {
		t.Fatal("Backup connection not established")
	}
	
	// 发送新命令
	err := sm.SendCommand("device-001", []byte("command2"))
	if err != nil {
		t.Errorf("Command failed after failover: %v", err)
	}
	
	// 验证离线命令恢复
	commands := sm.cache.GetCommands("device-001")
	if len(commands) != 0 {
		t.Errorf("Expected 0 cached commands, got %d", len(commands))
	}
}