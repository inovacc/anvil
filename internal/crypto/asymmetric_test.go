package crypto

import (
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	tests := []struct {
		name      string
		algorithm KeyAlgorithm
		wantErr   bool
	}{
		{"ed25519", AlgorithmEd25519, false},
		{"ecdsa-p256", AlgorithmECDSAP256, false},
		{"unsupported", KeyAlgorithm("rsa"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kp, err := GenerateKeyPair(tt.algorithm)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(kp.PrivateKey) == 0 {
				t.Fatal("empty private key")
			}

			if len(kp.PublicKey) == 0 {
				t.Fatal("empty public key")
			}

			if kp.Algorithm != tt.algorithm {
				t.Fatalf("algorithm mismatch: got %s, want %s", kp.Algorithm, tt.algorithm)
			}
		})
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		algorithm KeyAlgorithm
		data      []byte
	}{
		{"ed25519/hello", AlgorithmEd25519, []byte("hello world")},
		{"ed25519/empty", AlgorithmEd25519, []byte{}},
		{"ed25519/large", AlgorithmEd25519, make([]byte, 10000)},
		{"ecdsa-p256/hello", AlgorithmECDSAP256, []byte("hello world")},
		{"ecdsa-p256/empty", AlgorithmECDSAP256, []byte{}},
		{"ecdsa-p256/large", AlgorithmECDSAP256, make([]byte, 10000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kp, err := GenerateKeyPair(tt.algorithm)
			if err != nil {
				t.Fatalf("generate key: %v", err)
			}

			sig, err := Sign(kp.PrivateKey, tt.algorithm, tt.data)
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			valid, err := Verify(kp.PublicKey, tt.algorithm, tt.data, sig)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}

			if !valid {
				t.Fatal("signature should be valid")
			}

			// Tamper with data
			tampered := append([]byte{}, tt.data...)
			tampered = append(tampered, 0xFF)

			valid, err = Verify(kp.PublicKey, tt.algorithm, tampered, sig)
			if err != nil {
				t.Fatalf("verify tampered: %v", err)
			}

			if valid {
				t.Fatal("tampered data should not verify")
			}
		})
	}
}

func TestFingerprint(t *testing.T) {
	kp, err := GenerateKeyPair(AlgorithmEd25519)
	if err != nil {
		t.Fatal(err)
	}

	fp := Fingerprint(kp.PublicKey)
	if len(fp) != 16 { // 8 bytes = 16 hex chars
		t.Fatalf("fingerprint length: got %d, want 16", len(fp))
	}

	// Deterministic
	fp2 := Fingerprint(kp.PublicKey)
	if fp != fp2 {
		t.Fatal("fingerprint should be deterministic")
	}
}

func TestPEMRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		algorithm KeyAlgorithm
	}{
		{"ed25519", AlgorithmEd25519},
		{"ecdsa-p256", AlgorithmECDSAP256},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/private", func(t *testing.T) {
			kp, err := GenerateKeyPair(tt.algorithm)
			if err != nil {
				t.Fatal(err)
			}

			pemData := MarshalPrivateKeyPEM(kp.PrivateKey)

			keyBytes, isPrivate, alg, err := ParsePEMKey(pemData)
			if err != nil {
				t.Fatalf("parse PEM: %v", err)
			}

			if !isPrivate {
				t.Fatal("should be private")
			}

			if alg != tt.algorithm {
				t.Fatalf("algorithm: got %s, want %s", alg, tt.algorithm)
			}

			// Sign with parsed key, verify with original
			sig, err := Sign(keyBytes, alg, []byte("test"))
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			valid, err := Verify(kp.PublicKey, alg, []byte("test"), sig)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}

			if !valid {
				t.Fatal("should verify")
			}
		})

		t.Run(tt.name+"/public", func(t *testing.T) {
			kp, err := GenerateKeyPair(tt.algorithm)
			if err != nil {
				t.Fatal(err)
			}

			pemData := MarshalPublicKeyPEM(kp.PublicKey)

			keyBytes, isPrivate, alg, err := ParsePEMKey(pemData)
			if err != nil {
				t.Fatalf("parse PEM: %v", err)
			}

			if isPrivate {
				t.Fatal("should be public")
			}

			if alg != tt.algorithm {
				t.Fatalf("algorithm: got %s, want %s", alg, tt.algorithm)
			}

			// Verify with parsed key
			sig, err := Sign(kp.PrivateKey, alg, []byte("test"))
			if err != nil {
				t.Fatalf("sign: %v", err)
			}

			valid, err := Verify(keyBytes, alg, []byte("test"), sig)
			if err != nil {
				t.Fatalf("verify: %v", err)
			}

			if !valid {
				t.Fatal("should verify")
			}
		})
	}
}

func TestExtractPublicKeyFromPrivate(t *testing.T) {
	tests := []struct {
		name      string
		algorithm KeyAlgorithm
	}{
		{"ed25519", AlgorithmEd25519},
		{"ecdsa-p256", AlgorithmECDSAP256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kp, err := GenerateKeyPair(tt.algorithm)
			if err != nil {
				t.Fatal(err)
			}

			pubKey, err := ExtractPublicKeyFromPrivate(kp.PrivateKey, tt.algorithm)
			if err != nil {
				t.Fatalf("extract: %v", err)
			}

			// Verify using extracted public key
			sig, err := Sign(kp.PrivateKey, tt.algorithm, []byte("test"))
			if err != nil {
				t.Fatal(err)
			}

			valid, err := Verify(pubKey, tt.algorithm, []byte("test"), sig)
			if err != nil {
				t.Fatal(err)
			}

			if !valid {
				t.Fatal("should verify with extracted public key")
			}
		})
	}
}

func TestParsePEMKeyInvalid(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"not pem", []byte("not a pem")},
		{"bad type", []byte("-----BEGIN CERTIFICATE-----\nMQ==\n-----END CERTIFICATE-----\n")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := ParsePEMKey(tt.data)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
