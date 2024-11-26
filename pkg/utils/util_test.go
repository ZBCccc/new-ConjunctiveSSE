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

func TestComputeExp(t *testing.T) {
	tests := []struct {
		name    string
		x       *big.Int
		wantErr bool
	}{
		{
			name:    "small number",
			x:       big.NewInt(123),
			wantErr: false,
		},
		{
			name:    "zero",
			x:       big.NewInt(0),
			wantErr: false,
		},
		{
			name: "large number",
			x: func() *big.Int {
				n, _ := new(big.Int).SetString("123456789123456789123456789", 10)
				return n
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			got, err := ComputeExp(tt.x)
			duration := time.Since(start)

			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeExp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Error("ComputeExp() returned nil result")
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

func BenchmarkComputeExp(b *testing.B) {
	x, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeExp(x)
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

// 测试不同指数大小的性能
func BenchmarkComputeExp_DifferentSizes(b *testing.B) {
	sizes := []int{32, 64, 128, 256}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("exponent_bits_%d", size), func(b *testing.B) {
			x, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), uint(size)))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = ComputeExp(x)
			}
		})
	}
} 