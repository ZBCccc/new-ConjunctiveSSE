package utils

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"
)

// FAesni encrypts input using AES-ECB mode and processes the result based on the option
func FAesni(key []byte, input []byte, option int) ([]byte, error) {
	// Parameter validation
	if key == nil || input == nil {
		return nil, errors.New("key or input cannot be nil")
	}
	if option != 1 && option != 2 {
		return nil, errors.New("invalid option: must be 1 or 2")
	}
	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Calculate required number of blocks
	blockSize := aes.BlockSize
	blockCount := (len(input) + blockSize - 1) / blockSize

	// Create input block array
	inputBlocks := make([]byte, blockCount*blockSize)
	copy(inputBlocks, input)

	// Pad the last block
	if len(input)%blockSize != 0 {
		inputBlocks[len(input)] = 0x80
	}

	// Encrypt each block using AES-ECB
	for i := 0; i < len(inputBlocks); i += blockSize {
		block.Encrypt(inputBlocks[i:i+blockSize], inputBlocks[i:i+blockSize])
	}

	// Process result based on option
	// If option is 1 and input length <= 16, return encrypted data.
	// If option is 1 and input length > 16, return first 16 bytes of SHA256 hash of encrypted data.
	// If option is 2, return full SHA256 hash of encrypted data.
	switch {
	case option == 1 && len(input) <= 16:
		return inputBlocks[:16], nil
	case option == 1 && len(input) > 16:
		hash := sha256.Sum256(inputBlocks)
		return hash[:16], nil
	case option == 2:
		hash := sha256.Sum256(inputBlocks)
		return hash[:], nil
	default:
		return nil, nil
	}
}

// Xor performs XOR operation on two byte slices
func Xor(s1, s2 []byte) []byte {
	if len(s1) > len(s2) {
		fmt.Printf("not sufficient size: %d, %d\n", len(s1), len(s2))
		return nil
	}

	result := make([]byte, len(s1))
    copy(result, s1)
	for i := 0; i < len(result); i++ {
		result[i] ^= s2[i]
	}
	return result
}
