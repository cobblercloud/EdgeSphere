package device

import (
	"context"
	"sync"
	
	"edgesphere/internal/pkg/types"
)

type DeviceManager struct {
	store  DeviceStore
	cache  DeviceCache
	mu     sync.RWMutex
}

func NewDeviceManager(store DeviceStore, cache DeviceCache) *DeviceManager {
	return &DeviceManager{
		store: store,
		cache: cache,
	}
}

// 设备注册 (使用Bloom过滤器防重)
func (dm *DeviceManager) RegisterDevice(ctx context.Context, device *types.Device) error {
	if dm.cache.Exists(device.ID) {
		return errors.New("device already exists")
	}
	
	if err := dm.store.Save(ctx, device); err != nil {
		return err
	}
	
	dm.cache.Add(device.ID)
	return nil
}

// 批量注册优化
func (dm *DeviceManager) BatchRegister(devices []*types.Device) error {
	// 预分配ID范围
	ids := make([]string, len(devices))
	for i, d := range devices {
		ids[i] = d.ID
	}
	
	// 批量检查存在性
	if exists := dm.cache.BatchExists(ids); exists {
		return errors.New("some devices already exist")
	}
	
	return dm.store.BatchSave(devices)
}

// 状态更新
func (dm *DeviceManager) UpdateStatus(deviceID string, status types.DeviceStatus) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	// 先更新缓存
	dm.cache.SetStatus(deviceID, status)
	
	// 异步更新数据库
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		dm.store.UpdateStatus(ctx, deviceID, status)
	}()
}