package utils

import (
	"crypto/aes"
	"crypto/sha256"
	"errors"
	"fmt"
)

// FAesni 使用AES-ECB模式加密输入，并根据选项处理结果
func FAesni(key []byte, input []byte, option int) ([]byte, error) {
	// 参数验证
	if key == nil || input == nil {
		return nil, errors.New("key or input cannot be nil")
	}
	if option != 1 && option != 2 {
		return nil, errors.New("invalid option: must be 1 or 2")
	}
	// 创建AES加密器
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 计算需要的块数
	blockSize := aes.BlockSize
	blockCount := (len(input) + blockSize - 1) / blockSize

	// 创建输入块数组
	inputBlocks := make([]byte, blockCount*blockSize)
	copy(inputBlocks, input)

	// 填充最后一个块
	if len(input)%blockSize != 0 {
		inputBlocks[len(input)] = 0x80
	}

	// 对每个块进行AES-ECB加密
	for i := 0; i < len(inputBlocks); i += blockSize {
		block.Encrypt(inputBlocks[i:i+blockSize], inputBlocks[i:i+blockSize])
	}

	// 根据选项处理结果
	// 如果选项为1且输入长度<=16，返回加密后的数据。
	// 如果选项为1且输入长度>16，返回加密数据的SHA256哈希的前16字节。
	// 如果选项为2，返回加密数据的完整SHA256哈希。
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

// Xor 对两个byte切片进行异或运算
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
