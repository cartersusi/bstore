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

func WriteNewFile(fpath string, data []byte, encrypt bool) error {
	file, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteFile(file, data, encrypt)
}

func MkDirExt(fpath, base_path string, compress bool) (string, error) {
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

func MkDir(fpath string) error {
	if fpath == "" {
		return errors.New("file path is required")
	}

	if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
		return errors.New("error creating directory")
	}

	return nil
}

func ListDir(fpath string) ([]string, error) {
	var files []string
	dir, err := os.ReadDir(fpath)
	if err != nil {
		return nil, err
	}

	for _, file := range dir {
		files = append(files, file.Name())
	}

	return files, nil
}

func Contains(source, target string) bool {
	length := len(target)
	if length > len(source) {
		return false
	}

	for i := 0; i <= len(source)-length; i++ {
		if source[i:i+length] == target {
			return true
		}
	}

	return false
}
