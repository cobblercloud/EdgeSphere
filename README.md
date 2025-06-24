# EdgeSphere

edgesphere-v0.1/
├── cmd/
│ ├── edge-gateway/
│ │ └── main.go
│ ├── device-manager/
│ │ └── main.go
│ └── api-gateway/
│ └── main.go
├── internal/
│ ├── gateway/
│ │ ├── connection_pool.go
│ │ ├── session_manager.go
│ │ └── sqlite_cache.go
│ ├── protocol/
│ │ └── mqtt/
│ │ ├── adapter.go
│ │ └── decoder.go
│ ├── device/
│ │ ├── manager.go
│ │ ├── postgres_store.go
│ │ └── redis_cache.go
│ └── pkg/
│ ├── utils/
│ │ └── consistent_hash.go
│ └── types/
│ └── device.go
├── tests/
│ ├── load_test.go
│ └── failover_test.go
├── go.mod
└── Dockerfile

# 启动基础设施:

docker-compose up -d postgres redis mqtt

# 编译并运行服务:

## 边缘网关

go build -o bin/edge-gateway ./cmd/edge-gateway
./bin/edge-gateway

## 设备管理

go build -o bin/device-manager ./cmd/device-manager
./bin/device-manager

## API 网关

go build -o bin/api-gateway ./cmd/api-gateway
./bin/api-gateway

# 执行测试:

## 单元测试

go test -v ./...

## 负载测试 (10K 设备模拟)

go test -v ./tests -run TestConnectionScaling

## 故障转移测试

go test -v ./tests -run TestFailoverRecovery
