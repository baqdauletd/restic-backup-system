package logic

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
)

func encryptAndCompress(data []byte) (*bytes.Buffer, error) {
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
