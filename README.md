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
