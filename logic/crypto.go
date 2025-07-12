package logic

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"io"
)

func EncryptAndCompress(data []byte) (*bytes.Buffer, error) {
	// compress
	var compBuf bytes.Buffer
	gz := gzip.NewWriter(&compBuf)
	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}
	gz.Close()

	// encrypt
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, iv)

	in := compBuf.Bytes()
	out := make([]byte, len(in))
	stream.XORKeyStream(out, in)

	return bytes.NewBuffer(out), nil
}

func DecryptAndDecompress(encData []byte) ([]byte, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	stream := cipher.NewCTR(block, iv)

	decrypted := make([]byte, len(encData))
	stream.XORKeyStream(decrypted, encData)

	// Decompress
	buf := bytes.NewBuffer(decrypted)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	var out bytes.Buffer
	_, err = io.Copy(&out, gz)
	if err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}

