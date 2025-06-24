package mqtt

import (
	"encoding/binary"
	"errors"
	
	"edgesphere/internal/pkg/types"
)

type MQTTAdapter struct {
	conn      net.Conn
	deviceID  string
	messageCh chan []byte
}

func NewMQTTAdapter(conn net.Conn) *MQTTAdapter {
	return &MQTTAdapter{
		conn:      conn,
		messageCh: make(chan []byte, 100),
	}
}

func (a *MQTTAdapter) Send(data []byte) error {
	if a.conn == nil {
		return errors.New("connection closed")
	}
	
	// MQTT协议简化帧: [类型(1)|长度(2)|数据(N)]
	header := make([]byte, 3)
	header[0] = 0x30 // PUBLISH
	binary.BigEndian.PutUint16(header[1:], uint16(len(data)))
	
	frame := append(header, data...)
	_, err := a.conn.Write(frame)
	return err
}

func (a *MQTTAdapter) Listen() {
	defer a.conn.Close()
	buf := make([]byte, 1024)
	
	for {
		n, err := a.conn.Read(buf)
		if err != nil {
			break
		}
		a.messageCh <- buf[:n]
	}
}