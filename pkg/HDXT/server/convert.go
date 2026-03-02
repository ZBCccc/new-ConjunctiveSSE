package server

func xor(s1, s2 string) string {
	// Convert string to byte slice
	b1 := []byte(s1)
	b2 := []byte(s2)

	// Get the shorter length
	minLen := len(b1)
	if len(b2) < minLen {
		minLen = len(b2)
	}

	// Use the longer slice as result
	var result []byte
	if len(b1) > len(b2) {
		result = make([]byte, len(b1))
		copy(result, b1)
	} else {
		result = make([]byte, len(b2))
		copy(result, b2)
	}

	// XOR the part with minimum length
	for i := 0; i < minLen; i++ {
		result[i] = b1[i] ^ b2[i]
	}

	return string(result)
}
