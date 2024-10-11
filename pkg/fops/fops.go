package fops

import (
	"errors"
	"os"
	"path/filepath"
)

func ReadFile(fpath string, decrypt bool) ([]byte, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	if decrypt {
		data, err = Decrypt(data)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func WriteFile(file *os.File, data []byte, encrypt bool) error {
	var err error
	if encrypt {
		data, err = Encrypt(data)
		if err != nil {
			return err
		}
	}

	_, err = file.Write(data)
	return err
}

func MkDir(fpath, base_path string, compress bool) (string, error) {
	if fpath == "" {
		return "", errors.New("file path is required")
	}

	fpath = filepath.Join(base_path, fpath)
	if compress {
		fpath += ".zst"
	}

	dir := filepath.Dir(fpath)
	if dir == "" {
		return "", errors.New("error getting directory")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", errors.New("error creating directory")
	}

	return fpath, nil
}
