package proto

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
)

type Frame struct {
	cmd  byte
	opt  byte
	body []byte
}

const (
	devLogin    byte = 0x01
	startStream byte = 0x02
	stopStream  byte = 0x03
	audioStream byte = 0x04
)

func NewFrame(cmd, opt byte, body []byte) *Frame {
	return &Frame{
		cmd:  cmd,
		opt:  opt,
		body: body,
	}
}

/*
opt:
前两位 压缩，即支持3种压缩算法？

*/

func fromReader(r io.Reader) (*Frame, error) {
	var head = make([]byte, 4)
	_, err := io.ReadFull(r, head)
	if err != nil {
		return nil, fmt.Errorf("read head failed %w", err)
	}

	left := binary.BigEndian.Uint16(head[:2])
	var data = make([]byte, left-4)
	_, err = io.ReadFull(r, data)
	if err != nil {
		return nil, fmt.Errorf("read body failed %w", err)
	}

	// 此处考虑加密与压缩
	opt := head[3]
	var body []byte
	var compressMethod byte = (opt & (3 << 6)) >> 6
	if compressMethod == 0x01 {
		body, err = NewZstd().Decompress(data)
		if err != nil {
			return nil, fmt.Errorf("decompress data failed")
		}
	} else if compressMethod == 0x00 {
		body = data
	}

	return &Frame{
		cmd:  head[2],
		opt:  opt,
		body: body,
	}, nil
}

func NewLoginResponse(code byte) *Frame {
	return &Frame{
		cmd:  devLogin,
		opt:  0,
		body: []byte{code},
	}
}

func (m *Frame) WriteTo(w io.Writer) (int64, error) {
	var err error
	var data []byte
	var cm byte = (m.opt & (3 << 6)) >> 6
	if cm == 0x00 {
		data = m.body
	} else if cm == 0x01 {
		data, err = NewZstd().Compress(m.body)
		if err != nil {
			return 0, fmt.Errorf("compress data failed %w", err)
		}
	}

	if (len(data) + 4) >= 0xFFFF {
		return 0, fmt.Errorf("too big body %d", len(data))
	}
	var head = make([]byte, 4)
	var length = uint16(4 + len(data))
	binary.BigEndian.PutUint16(head[:2], length)
	head[2] = m.cmd
	head[3] = m.opt
	headerSent, err := w.Write(head)
	if err != nil {
		return 0, fmt.Errorf("send head failed %w", err)
	}
	if headerSent != len(head) {
		return 0, fmt.Errorf("send head failed, length not match %d %d", headerSent, len(head))
	}

	bodySent, err := w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("send body failed %w", err)
	}
	if bodySent != len(data) {
		return 0, fmt.Errorf("send body failed, length not match %d %d", bodySent, len(data))
	}
	return int64(headerSent + bodySent), nil
}

// 设备登录时，提供用户会话值与设备UUID
type loginRequest struct {
	Session string `json:"session"`
	Device  string `json:"device"`
}

func parseLoginFrame(f *Frame) (*loginRequest, error) {
	var rs loginRequest
	err := json.Unmarshal(f.body, &rs)
	return &rs, err
}

type startStreamRequest struct {
	Bits       int `json:"bits"`
	Channels   int `json:"channels"`
	SampleRate int `json:"sampleRate"`
}

func parseStartStreamRequest(f *Frame) (*startStreamRequest, error) {
	var rs startStreamRequest
	err := json.Unmarshal(f.body, &rs)
	return &rs, err
}

// 回复客户端 code 为0代表成功
type startStreamResponse struct {
	Code int `json:"code"`
	ID   int `json:"id"`
}

func NewStartStreamResponseFrame(code, id int) (*Frame, error) {
	var rs = &startStreamResponse{
		Code: code,
		ID:   id,
	}
	body, err := json.Marshal(rs)
	if err != nil {
		return nil, fmt.Errorf("json marshal body failed %w", err)
	}
	return NewFrame(startStream, 0, body), nil
}
