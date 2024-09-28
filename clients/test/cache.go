package main

import "io"

type audioCache struct {
	sampleRate int
	bitsDepth  int
	chanCount  int
	buf        []byte
	out        io.ReadWriteCloser
}

func NewAudioCache(o io.ReadWriteCloser, sampleRate, bitsDepth, chanCount int) *audioCache {
	return &audioCache{
		sampleRate: sampleRate,
		bitsDepth:  bitsDepth,
		chanCount:  chanCount,
		buf:        make([]byte, 960*100),
		out:        o,
	}
}

/*
将音频 PCM 数据加入缓存
控制，每次发出 50 ms 数据
*/
func (m *audioCache) Add(d []byte) {
	// nil 为结束，清空缓冲区
	if d == nil && len(m.buf) > 0 {
		m.send()
		return
	}

	m.buf = append(m.buf, d...)
	// 首先计算 50 ms 的数据量
	bits := m.sampleRate / 1000 * (m.bitsDepth / 8) * 50
	if len(m.buf) > bits {
		m.send()
	}
}

func (m *audioCache) send() {
	var opt byte = 1 << 6
	_, err := NewFrame(audioStream, opt, m.buf).WriteTo(m.out)
	if err != nil {
		logger.Warn("输出音频流失败", "error", err)
	}
	m.buf = make([]byte, 0)
}
