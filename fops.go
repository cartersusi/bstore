package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"os"

	"golang.org/x/build/pargzip"
)

func Compress(buf *bytes.Buffer, outfname string) error {
	outfile, err := os.Create(outfname)
	if err != nil {
		return err
	}
	defer outfile.Close()

	gzipWriter := gzip.NewWriter(outfile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, buf)
	if err != nil {
		return err
	}

	return nil
}

func pCompress(buf *bytes.Buffer, outfname string) error {
	outfile, err := os.Create(outfname)
	if err != nil {
		return err
	}
	defer outfile.Close()

	gzipWriter := pargzip.NewWriter(outfile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, buf)
	if err != nil {
		return err
	}

	return nil
}

func Decompress(fpath string) (*bytes.Buffer, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gzipReader)
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
