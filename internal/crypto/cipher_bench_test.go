package crypto

import (
	"testing"
)

func BenchmarkEncrypt(b *testing.B) {
	key, _ := GenerateKey()
	plaintext := []byte("secret-value-for-benchmark-testing")

	b.ResetTimer()
	for b.Loop() {
		_, _, _ = Encrypt(key, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key, _ := GenerateKey()
	plaintext := []byte("secret-value-for-benchmark-testing")
	ciphertext, nonce, _ := Encrypt(key, plaintext)

	b.ResetTimer()
	for b.Loop() {
		_, _ = Decrypt(key, ciphertext, nonce)
	}
}

func BenchmarkEncryptLarge(b *testing.B) {
	key, _ := GenerateKey()
	plaintext := make([]byte, 64*1024) // 64KB

	b.ResetTimer()
	for b.Loop() {
		_, _, _ = Encrypt(key, plaintext)
	}
}

func BenchmarkDeriveKey(b *testing.B) {
	salt, _ := GenerateSalt()

	b.ResetTimer()
	for b.Loop() {
		_, _ = DeriveKey("machine-id-value", salt)
	}
}

func BenchmarkSealUnsealMasterKey(b *testing.B) {
	masterKey, _ := GenerateKey()
	salt, _ := GenerateSalt()
	machineID := "bench-machine-id"

	sealed, nonce, _ := SealMasterKey(masterKey, machineID, salt)

	b.Run("Seal", func(b *testing.B) {
		for b.Loop() {
			_, _, _ = SealMasterKey(masterKey, machineID, salt)
		}
	})

	b.Run("Unseal", func(b *testing.B) {
		for b.Loop() {
			_, _ = UnsealMasterKey(sealed, nonce, machineID, salt)
		}
	})
}

func BenchmarkGenerateKey(b *testing.B) {
	for b.Loop() {
		_, _ = GenerateKey()
	}
}

func BenchmarkGenerateSalt(b *testing.B) {
	for b.Loop() {
		_, _ = GenerateSalt()
	}
}
