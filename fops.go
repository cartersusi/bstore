package main

import (
	"bytes"

	"io"
	"os"

	//"golang.org/x/build/pargzip"
	"github.com/klauspost/compress/zstd"
)

func Decompress(fpath string) (*bytes.Buffer, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	zstdReader, err := zstd.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer zstdReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, zstdReader)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func WriteBuffer(buf *bytes.Buffer, fpath string) error {
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, buf)
	return err
}
