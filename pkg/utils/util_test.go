package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
	"time"
)

func TestFp(t *testing.T) {
	// 测试用例
	tests := []struct {
		name    string
		key     []byte
		message []byte
		p       *big.Int
		wantErr bool
	}{
		{
			name:    "basic test",
			key:     []byte("test-key"),
			message: []byte("test-message"),
			p:       new(big.Int).SetInt64(97), // 使用小素数便于验证
			wantErr: false,
		},
		{
			name:    "empty message",
			key:     []byte("test-key"),
			message: []byte{},
			p:       new(big.Int).SetInt64(97),
			wantErr: false,
		},
		{
			name:    "large prime",
			key:     []byte("test-key"),
			message: []byte("test-message"),
			p:       new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19)), // Curve25519 prime
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := Fp(tt.key, tt.message, tt.p)
			duration := time.Since(start)

			if (err != nil) != tt.wantErr {
				t.Errorf("Fp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Error("Fp() returned nil result")
				return
			}
			// 验证结果是否在正确范围内
			if got.Cmp(big.NewInt(1)) < 0 || got.Cmp(tt.p) >= 0 {
				t.Errorf("Fp() result not in range [1, p-1]")
			}
			// 验证结果是否有逆元
			gcd := new(big.Int).GCD(nil, nil, got, tt.p)
			if gcd.Cmp(big.NewInt(1)) != 0 {
				t.Errorf("Fp() result %v does not have multiplicative inverse modulo %v", got, tt.p)
			}
			t.Logf("Duration: %v", duration)
		})
	}
}

// 性能测试
func BenchmarkFp(b *testing.B) {
	key := make([]byte, 32)
	message := make([]byte, 64)
	rand.Read(key)
	rand.Read(message)
	p := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Fp(key, message, p)
	}
}

// 测试不同输入大小的性能
func BenchmarkFp_DifferentSizes(b *testing.B) {
	sizes := []int{32, 64, 128, 256, 512}
	p := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19))

	for _, size := range sizes {
		b.Run(fmt.Sprintf("input_size_%d", size), func(b *testing.B) {
			key := make([]byte, 32)
			message := make([]byte, size)
			rand.Read(key)
			rand.Read(message)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Fp(key, message, p)
			}
		})
	}
}