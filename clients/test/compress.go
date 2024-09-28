package main

import (
	"fmt"

	"github.com/klauspost/compress/zstd"
)

type zstdReadWriter struct {
}

func NewZstd() *zstdReadWriter {
	return &zstdReadWriter{}
}

func (m *zstdReadWriter) Compress(raw []byte) ([]byte, error) {
	encoder, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, fmt.Errorf("new zstd writer failed")
	}
	defer encoder.Close()

	return encoder.EncodeAll(raw, nil), nil
}

func (m *zstdReadWriter) Decompress(raw []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil)
	if err != nil {
		return nil, fmt.Errorf("new zstd writer failed")
	}
	defer decoder.Close()
	return decoder.DecodeAll(raw, nil)
}
