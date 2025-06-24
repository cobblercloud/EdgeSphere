package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"edgesphere/internal/device"
	"edgesphere/internal/pkg/types"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 初始化存储
	pgStore, err := device.NewPostgresStore("postgres://user:pass@localhost/edgesphere")
	if err != nil {
		log.Fatalf("Failed to init Postgres: %v", err)
	}
	
	// 初始化缓存
	redisCache := device.NewRedisCache("localhost:6379", "", 0)
	
	// 创建设备管理器
	devMgr := device.NewDeviceManager(pgStore, redisCache)
	
	// 启动状态监听
	go watchDeviceStatus(ctx, devMgr, redisCache)
	
	// 启动gRPC服务
	go startGRPCServer(devMgr, 50051)
	
	// 启动HTTP API
	go startHTTPServer(devMgr, 8080)
	
	log.Println("Device Manager started")
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down device manager...")
}

func watchDeviceStatus(ctx context.Context, mgr *device.DeviceManager, cache *device.RedisCache) {
	updates := cache.SubscribeStatusUpdates()
	for {
		select {
		case msg := <-updates:
			var statusUpdate struct {
				DeviceID  string          `json:"device_id"`
				Status    types.DeviceStatus `json:"status"`
				Timestamp int64           `json:"timestamp"`
			}
			if err := json.Unmarshal([]byte(msg.Payload), &statusUpdate); err != nil {
				log.Printf("Failed to parse status update: %v", err)
				continue
			}
			
			// 更新设备状态
			mgr.UpdateStatus(statusUpdate.DeviceID, statusUpdate.Status)
			
		case <-ctx.Done():
			return
		}
	}
}

func startHTTPServer(mgr *device.DeviceManager, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		// 实现设备列表API
	})
	
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}
	
	log.Printf("HTTP server started on :%d", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}