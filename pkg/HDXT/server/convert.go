package server

func xor(s1, s2 string) string {
	// 将字符串转换为字节切片
	b1 := []byte(s1)
	b2 := []byte(s2)

	// 获取较短的长度
	minLen := len(b1)
	if len(b2) < minLen {
		minLen = len(b2)
	}

	// 使用较长的切片作为结果
	var result []byte
	if len(b1) > len(b2) {
		result = make([]byte, len(b1))
		copy(result, b1)
	} else {
		result = make([]byte, len(b2))
		copy(result, b2)
	}

	// 对最小长度的部分进行异或操作
	for i := 0; i < minLen; i++ {
		result[i] = b1[i] ^ b2[i]
	}

	return string(result)
}