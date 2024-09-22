package main

import (
	"fmt"
	"io"
	"os"

	"github.com/braheezy/shine-mp3/pkg/mp3"
	"github.com/go-audio/wav"
)

type wavWriter struct {
	out *wav.Encoder
}

func NewWavWriter(fileName string, sampleRate, bitDepth, numChans int) (io.WriteCloser, error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open dst file failed %w", err)
	}
	encoder := wav.NewEncoder(out, sampleRate, bitDepth, numChans, 1)
	return &wavWriter{
		out: encoder,
	}, nil
}

func (m *wavWriter) Write(p []byte) (n int, err error) {
	// m.out.Write()
	return 0, nil
}

func (m *wavWriter) Close() error {
	return nil
}

type mp3Writer struct {
	// 1, 2, 3, 4, 8 => in8, int16, int24, int32/float32, int64
	chanCount int
	bits      int
	out       io.Writer
	encoder   *mp3.Encoder
}

func NewMp3Writer(fileName string, sampleRate, bitDepth, numChans int) (io.WriteCloser, error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open dst file failed %w", err)
	}

	encoder := mp3.NewEncoder(sampleRate, numChans)
	return &mp3Writer{
		chanCount: numChans,
		bits:      bitDepth,
		out:       out,
		encoder:   encoder,
	}, nil
}

func (m *mp3Writer) Write(p []byte) (n int, err error) {
	if len(p)%(m.chanCount*m.bits) > 0 {
		return 0, fmt.Errorf("not match, length: %d, chan: %d, bits: %d", len(p), m.chanCount, m.bits)
	}
	// 重整为 int16
	// m.encoder.Write(m.out, )
	return 0, nil
}

func (m *mp3Writer) Close() error {
	return nil
}
