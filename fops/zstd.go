package fops

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/klauspost/compress/zstd"
)

func Compress(buf *bytes.Buffer, file *os.File, level int, encrypt bool) error {
	var compression_lvl zstd.EncoderLevel
	switch level {
	case 1:
		compression_lvl = zstd.SpeedFastest
	case 2:
		compression_lvl = zstd.SpeedDefault
	case 3:
		compression_lvl = zstd.SpeedBetterCompression
	case 4:
		compression_lvl = zstd.SpeedBestCompression
	default:
		compression_lvl = zstd.SpeedDefault
	}
	opts := []zstd.EOption{zstd.WithEncoderLevel(compression_lvl)}

	enc, err := zstd.NewWriter(file, opts...)
	if err != nil {
		return err
	}
	defer enc.Close()

	var data []byte
	if encrypt {
		log.Println("Encrypting data")
		data, err = Encrypt(buf.Bytes())
		if err != nil {
			return err
		}
	} else {
		data = buf.Bytes()
	}

	log.Println("Done Encrypting data")

	_, err = enc.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func Decompress(fpath string, encrypt bool) ([]byte, error) {
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

	if encrypt {
		return Decrypt(buf.Bytes())
	}
	return buf.Bytes(), nil
}
