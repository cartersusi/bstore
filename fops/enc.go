package fops

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
	"os"
)

func DecryptFile(fpath string) ([]byte, error) {
	file, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return nil, err
	}

	return Decrypt(buf.Bytes())
}

func Encrypt(data []byte) ([]byte, error) {
	key_string := os.Getenv("BSTORE_ENC_KEY")
	if key_string == "" {
		return nil, errors.New("BSTORE_ENC_KEY not set")
	}
	key := []byte(key_string)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func Decrypt(data []byte) ([]byte, error) {
	key_string := os.Getenv("BSTORE_ENC_KEY")
	if key_string == "" {
		return nil, errors.New("BSTORE_ENC_KEY not set")
	}
	key := []byte(key_string)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
