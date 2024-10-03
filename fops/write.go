package fops

import (
	"bytes"
	"io"
	"os"
)

func WriteBuffer(buf *bytes.Buffer, fpath string) error {
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, buf)
	return err
}
