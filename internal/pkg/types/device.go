package types

import (
	"time"
	"net"
)

type DeviceStatus int

const (
	Online DeviceStatus = iota
	Offline
	Unregistered
)

type Device struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Type         string       `json:"type"`
	Status       DeviceStatus `json:"status"`
	LastSeen     time.Time    `json:"last_seen"`
	GatewayID    string       `json:"gateway_id"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	Metadata     map[string]string `json:"metadata"`
}

type DeviceConnection struct {
	ID        string
	Adapter   ProtocolAdapter
	Status    DeviceStatus
	LastSeen  time.Time
	Fd        int // 文件描述符用于零拷贝
}

type ProtocolAdapter interface {
	Send(data []byte) error
	Close() error
}

type Command struct {
	DeviceID  string    `json:"device_id"`
	Command   string    `json:"command"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}