package proto

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
)

type Connection struct {
	conn   io.ReadWriteCloser
	ctx    context.Context
	cancel context.CancelFunc
	chIn   chan *Frame
	chOut  chan *Frame
	writer io.WriteCloser
}

func NewConnection(ctx context.Context, r io.ReadWriteCloser) *Connection {
	c, cancel := context.WithCancel(ctx)
	return &Connection{
		conn:   r,
		ctx:    c,
		cancel: cancel,
		chIn:   make(chan *Frame),
		chOut:  make(chan *Frame),
	}
}

func (m *Connection) ReadLoop() {
	defer m.Close()

	for {
		frame, err := fromReader(m.conn)
		if err != nil {
			logger.Warn("read from reader failed", "error", err)
			return
		}
		m.chIn <- frame
	}
}

func (m *Connection) Close() {
	err := m.conn.Close()
	if err != nil {
		logger.Warn("close failed", "error", err)
	}
	if m.writer != nil {
		m.writer.Close()
		m.writer = nil
	}
	m.cancel()
}

func (m *Connection) Handle() {
	var err error
	for {
		select {
		case frameIn, ok := <-m.chIn:
			if !ok {
				logger.Warn("channel in error")
				return
			}
			go func() {
				err = m.handleFrame(frameIn)
				if err != nil {
					logger.Warn("frame handle failed", "error", err)
				}
			}()
		case frameOut, ok := <-m.chOut:
			if !ok {
				logger.Warn("channel out error")
				return
			}
			logger.Info("channel out received frame", "cmd", frameOut.cmd)
			err = m.outFrame(frameOut)
			if err != nil {
				logger.Warn("send frame failed", "error", err)
			}
		case <-m.ctx.Done():
			logger.Warn("Done!")
			return
		}
	}
}

// 暂时直接返回成功
func (m *Connection) handleLoginDevice(r *loginRequest) error {
	logger.Info("process login request", "session", r.Session, "device", r.Device)
	m.writeFrame(NewLoginResponse(0x00))
	return nil
}

// 先返回一个 uint64 随机数
func (m *Connection) handleStartStream(req *startStreamRequest) error {
	if m.writer != nil {
		return fmt.Errorf("writer initialized, cannot init twice")
	}
	// w, err := NewMp3Writer("test1.mp3", req.SampleRate, req.Bits, req.Channels)
	// w, err := NewPcmWriter("test1.pcm", req.SampleRate, req.Bits, req.Channels)
	// w, err := NewWavWriter("test1.wav", req.SampleRate, req.Bits, req.Channels)
	w, err := NewGzipWavWriter("test1.wav", req.SampleRate, req.Bits, req.Channels)
	if err != nil {
		return fmt.Errorf("new mp3 writer failed %w", err)
	}
	m.writer = w
	out, err := NewStartStreamResponseFrame(0, 0x1001)
	if err != nil {
		return fmt.Errorf("new start stream response failed %w", err)
	}
	m.writeFrame(out)
	return nil
}

func (m *Connection) handleAudioStream(f *Frame) error {
	if m.writer == nil {
		return fmt.Errorf("writer not initialized")
	}
	_, err := m.writer.Write(f.body)
	return err
}

func (m *Connection) writeFrame(f *Frame) {
	m.chOut <- f
}

func (m *Connection) outFrame(f *Frame) error {
	_, err := f.WriteTo(m.conn)
	return err
}

func (m *Connection) handleFrame(f *Frame) error {
	switch f.cmd {
	case devLogin:
		loginReq, err := parseLoginFrame(f)
		if err != nil {
			return fmt.Errorf("parse login frame failed %w", err)
		}
		return m.handleLoginDevice(loginReq)
	case startStream:
		req, err := parseStartStreamRequest(f)
		if err != nil {
			return fmt.Errorf("parse start stream frame failed %w", err)
		}
		return m.handleStartStream(req)
	case stopStream:
		err := m.writer.Close()
		m.writer = nil
		return err
	case audioStream:
		return m.handleAudioStream(f)
	default:
		return fmt.Errorf("unsupported command %d", f.cmd)
	}
}
