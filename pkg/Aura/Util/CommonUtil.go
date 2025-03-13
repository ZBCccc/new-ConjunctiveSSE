package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"errors"
)

const (
	AESBlockSize = 16
	DIGESTSize   = 32
	GGMSize      = 65536
	HashSize     = 13
)

type Operation int

const (
	Insert Operation = iota
	Delete
)

func AesEncrypt(plaintext, key, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		return nil, errors.New("iv length must equal block size")
	}
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("invalid key size")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	ciphertext := make([]byte, len(plaintext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext, plaintext)
	return ciphertext, nil
}

func AesDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}

func Sha256Digest(plaintext []byte) []byte {
	digest := sha256.Sum256(plaintext)
	return digest[:]
}

func HmacDigest(plaintext, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(plaintext)
	return h.Sum(nil)
}

func KeyDerivation(plaintext, key []byte) []byte {
	h := hmac.New(md5.New, key)
	h.Write(plaintext)
	return h.Sum(nil)
}
