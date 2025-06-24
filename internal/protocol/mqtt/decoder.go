package mqtt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

type ControlPacket byte

const (
	Connect     ControlPacket = 1
	ConnAck     ControlPacket = 2
	Publish     ControlPacket = 3
	PubAck      ControlPacket = 4
	Subscribe   ControlPacket = 8
	SubAck      ControlPacket = 9
	Unsubscribe ControlPacket = 10
	UnsubAck    ControlPacket = 11
	PingReq     ControlPacket = 12
	PingResp    ControlPacket = 13
	Disconnect  ControlPacket = 14
)

type Header struct {
	Type      ControlPacket
	Flags     byte
	Remaining int
}

func DecodeHeader(r io.Reader) (*Header, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	header := &Header{
		Type:  ControlPacket(buf[0] >> 4),
		Flags: buf[0] & 0x0F,
	}

	// 解码剩余长度
	multiplier := 1
	value := 0
	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		digit := buf[0]
		value += int(digit&127) * multiplier
		multiplier *= 128
		if (digit & 128) == 0 {
			break
		}
	}

	header.Remaining = value
	return header, nil
}

func DecodeConnectPacket(r io.Reader) (map[string]interface{}, error) {
	header, err := DecodeHeader(r)
	if err != nil || header.Type != Connect {
		return nil, errors.New("invalid CONNECT packet")
	}

	// 读取协议名
	protoName, err := readString(r)
	if err != nil {
		return nil, err
	}

	// 协议版本
	version, err := readByte(r)
	if err != nil {
		return nil, err
	}

	// 连接标志
	flags, err := readByte(r)
	if err != nil {
		return nil, err
	}

	// 保活时间
	keepAlive, err := readUint16(r)
	if err != nil {
		return nil, err
	}

	// 客户端ID
	clientID, err := readString(r)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"protocol":   protoName,
		"version":    version,
		"flags":      flags,
		"keep_alive": keepAlive,
		"client_id":  clientID,
	}, nil
}

func readString(r io.Reader) (string, error) {
	lenBuf := make([]byte, 2)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint16(lenBuf)

	strBuf := make([]byte, length)
	if _, err := io.ReadFull(r, strBuf); err != nil {
		return "", err
	}
	return string(strBuf), nil
}

func readByte(r io.Reader) (byte, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	return buf[0], err
}

func readUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}