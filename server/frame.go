package main

import (
	"encoding/binary"
	"encoding/hex"
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

func fromReader(r io.Reader) (*Frame, error) {
	var head = make([]byte, 4)
	_, err := io.ReadFull(r, head)
	if err != nil {
		return nil, fmt.Errorf("read head failed %w", err)
	}

	left := binary.BigEndian.Uint16(head[:2])
	var body = make([]byte, left-4)
	_, err = io.ReadFull(r, body)
	if err != nil {
		return nil, fmt.Errorf("read body failed %w", err)
	}

	return &Frame{
		cmd:  head[2],
		opt:  head[3],
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
	if (len(m.body) + 4) >= 0xFFFF {
		return 0, fmt.Errorf("too big body %d", len(m.body))
	}
	var head = make([]byte, 4)
	var length = uint16(4 + len(m.body))
	binary.BigEndian.PutUint16(head[:2], length)
	head[2] = m.cmd
	head[3] = m.opt
	logger.Info("HEAD", "hex", hex.EncodeToString(head))
	headerSent, err := w.Write(head)
	if err != nil {
		return 0, fmt.Errorf("send head failed %w", err)
	}
	if headerSent != len(head) {
		return 0, fmt.Errorf("send head failed, length not match %d %d", headerSent, len(head))
	}

	logger.Info("HEAD", "hex", hex.EncodeToString(m.body))
	bodySent, err := w.Write(m.body)
	if err != nil {
		return 0, fmt.Errorf("send body failed %w", err)
	}
	if bodySent != len(m.body) {
		return 0, fmt.Errorf("send body failed, length not match %d %d", bodySent, len(m.body))
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
