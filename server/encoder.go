package main

import (
	"bytes"
	"encoding/binary"
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
	out       io.WriteCloser
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

func reSample8(buf []byte, out []int16) {
	for i := 0; i != len(out); i++ {

	}
}

func reSample16(buf []byte, out []int16) {
	reader := bytes.NewReader(buf)
	err := binary.Read(reader, binary.LittleEndian, &out)
	if err != nil {
		panic(err)
	}
}

func reSample24(buf []byte, out []int16) {
	for i := 0; i != len(out); i++ {
	}
}

func reSample32(buf []byte, out []int16) {
	for i := 0; i != len(out); i++ {
	}
}

// 重采样 无需理会声道 逐一处理即可
func reSample(buf []byte, out []int16, bits int) {
	if len(out) != (len(buf) / (bits / 8)) {
		panic("buffer length not match")
	}
	if bits == 8 {
		// reSample8(buf, out)
		panic("invalid now 8")
	} else if bits == 16 {
		reSample16(buf, out)
	} else if bits == 24 {
		// reSample24(buf, out)
		panic("invalid now 24")
	} else if bits == 32 {
		// reSample32(buf, out)
		panic("invalid now 32")
	}
}

func (m *mp3Writer) Write(p []byte) (n int, err error) {
	if len(p)%(m.chanCount*m.bits) > 0 {
		return 0, fmt.Errorf("not match, length: %d, chan: %d, bits: %d", len(p), m.chanCount, m.bits)
	}
	fmt.Printf("length: %d, chan: %d, bits: %d\n", len(p), m.chanCount, m.bits)

	var bufOut = make([]int16, len(p)/(m.bits/8))
	reSample(p, bufOut, m.bits)

	logger.Info("buffer", "len", len(bufOut), "buf", len(p))
	// 重整为 int16
	err = m.encoder.Write(m.out, bufOut)
	return len(p), err
}

func (m *mp3Writer) Close() error {
	return m.out.Close()
}

type PcmEncoder struct {
	out io.WriteCloser
}

func NewPcmWriter(fileName string, sampleRate, bitDepth, numChans int) (io.WriteCloser, error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open file failed %w", err)
	}

	return &PcmEncoder{
		out: out,
	}, nil
}

func (m *PcmEncoder) Write(p []byte) (n int, err error) {
	return m.out.Write(p)
}

func (m *PcmEncoder) Close() error {
	return m.out.Close()
}
