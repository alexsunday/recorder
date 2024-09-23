package proto

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/henryleu/go-wav"
)

type wavWriter struct {
	out *wav.Writer
}

func NewWavWriter(fileName string, sampleRate, bitDepth, numChans int) (io.WriteCloser, error) {
	out, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open dst file failed %w", err)
	}
	encoder, err := wav.NewWriter(wav.WriterParam{
		Out:           out,
		Channel:       numChans,
		SampleRate:    sampleRate,
		BitsPerSample: bitDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("new wav writer failed %w", err)
	}
	return &wavWriter{
		out: encoder,
	}, nil
}

func NewGzipWavWriter(fileName string, sampleRate, bitDepth, numChans int) (io.WriteCloser, error) {
	gzipFileName := fmt.Sprintf("%s.gz", fileName)
	out, err := os.OpenFile(gzipFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("open dst file failed %w", err)
	}

	zw := gzip.NewWriter(out)
	zw.Name = fileName
	zw.Comment = fileName
	zw.ModTime = time.Now()

	encoder, err := wav.NewWriter(wav.WriterParam{
		Out:           zw,
		Channel:       numChans,
		SampleRate:    sampleRate,
		BitsPerSample: bitDepth,
	})
	if err != nil {
		return nil, fmt.Errorf("new wav writer failed %w", err)
	}
	return &wavWriter{
		out: encoder,
	}, nil
}

func (m *wavWriter) Write(p []byte) (n int, err error) {
	return m.out.Write(p)
}

func (m *wavWriter) Close() error {
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
