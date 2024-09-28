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
	m.buf = append(m.buf, d...)
	// 首先计算 50 ms 的数据量
	bits := m.sampleRate / 1000 * (m.bitsDepth / 8) * 50
	if len(m.buf) > bits {
		err := m.send()
		if err != nil {
			logger.Warn("输出音频流失败", "error", err)
		}

		m.buf = make([]byte, 0)
	}
}

func (m *audioCache) send() error {
	var opt byte = 1 << 6
	_, err := NewFrame(audioStream, opt, m.buf).WriteTo(m.out)
	return err
}
