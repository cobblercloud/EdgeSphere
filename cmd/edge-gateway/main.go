package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	
	"edgesphere/internal/gateway"
	"edgesphere/internal/protocol/mqtt"
	"edgesphere/internal/pkg/utils"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初始化会话管理器
	sessionMgr := gateway.NewSessionManager()
	
	// 初始化一致性哈希
	hashRing := utils.NewConsistentHash(50)
	hashRing.AddNode("edge-gateway-1")
	hashRing.AddNode("edge-gateway-2")
	
	// 启动MQTT监听
	go startMQTTListener(ctx, sessionMgr, 1883)
	
	// 启动HTTP管理接口
	go startAdminAPI(sessionMgr, 8080)
	
	log.Println("Edge Gateway started successfully")
	
	// 等待终止信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down edge gateway...")
}

func startMQTTListener(ctx context.Context, mgr *gateway.SessionManager, port int) {
	addr := net.JoinHostPort("", strconv.Itoa(port))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start MQTT listener: %v", err)
	}
	defer listener.Close()
	log.Printf("MQTT listening on :%d", port)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		
		go handleMQTTConnection(ctx, conn, mgr)
	}
}

func handleMQTTConnection(ctx context.Context, conn net.Conn, mgr *gateway.SessionManager) {
	defer conn.Close()
	
	// 解析MQTT连接包
	connectInfo, err := mqtt.DecodeConnectPacket(conn)
	if err != nil {
		log.Printf("MQTT decode error: %v", err)
		return
	}
	
	deviceID, ok := connectInfo["client_id"].(string)
	if !ok || deviceID == "" {
		log.Println("Invalid device ID")
		return
	}
	
	// 创建协议适配器
	adapter := mqtt.NewMQTTAdapter(conn)
	go adapter.Listen()
	
	// 管理会话
	mgr.HandleConnection(ctx, deviceID, adapter)
	log.Printf("Device %s connected", deviceID)
	
	// 等待连接关闭
	<-adapter.Context().Done()
	log.Printf("Device %s disconnected", deviceID)
}