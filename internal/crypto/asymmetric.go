package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// KeyAlgorithm represents a supported asymmetric key algorithm.
type KeyAlgorithm string

const (
	AlgorithmEd25519   KeyAlgorithm = "ed25519"
	AlgorithmECDSAP256 KeyAlgorithm = "ecdsa-p256"
)

// KeyPair holds a generated asymmetric key pair as raw bytes.
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
	Algorithm  KeyAlgorithm
}

// GenerateKeyPair generates a new asymmetric key pair for the given algorithm.
func GenerateKeyPair(algorithm KeyAlgorithm) (*KeyPair, error) {
	switch algorithm {
	case AlgorithmEd25519:
		return generateEd25519()
	case AlgorithmECDSAP256:
		return generateECDSAP256()
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func generateEd25519() (*KeyPair, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519 key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privBytes,
		PublicKey:  pubBytes,
		Algorithm:  AlgorithmEd25519,
	}, nil
}

func generateECDSAP256() (*KeyPair, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ecdsa key: %w", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, fmt.Errorf("marshal private key: %w", err)
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privBytes,
		PublicKey:  pubBytes,
		Algorithm:  AlgorithmECDSAP256,
	}, nil
}

// Sign signs data with the given PKCS8-encoded private key and algorithm.
func Sign(privateKeyBytes []byte, algorithm KeyAlgorithm, data []byte) ([]byte, error) {
	privKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	switch algorithm {
	case AlgorithmEd25519:
		edKey, ok := privKey.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not ed25519")
		}

		return ed25519.Sign(edKey, data), nil

	case AlgorithmECDSAP256:
		ecKey, ok := privKey.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not ecdsa")
		}

		hash := sha256.Sum256(data)

		return ecdsa.SignASN1(rand.Reader, ecKey, hash[:])

	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// Verify verifies a signature with the given PKIX-encoded public key and algorithm.
func Verify(publicKeyBytes []byte, algorithm KeyAlgorithm, data, signature []byte) (bool, error) {
	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("parse public key: %w", err)
	}

	switch algorithm {
	case AlgorithmEd25519:
		edKey, ok := pubKey.(ed25519.PublicKey)
		if !ok {
			return false, fmt.Errorf("key is not ed25519")
		}

		return ed25519.Verify(edKey, data, signature), nil

	case AlgorithmECDSAP256:
		ecKey, ok := pubKey.(*ecdsa.PublicKey)
		if !ok {
			return false, fmt.Errorf("key is not ecdsa")
		}

		hash := sha256.Sum256(data)

		return ecdsa.VerifyASN1(ecKey, hash[:], signature), nil

	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

// Fingerprint returns the first 8 bytes of SHA-256(publicKey) as hex.
func Fingerprint(publicKey []byte) string {
	h := sha256.Sum256(publicKey)
	return hex.EncodeToString(h[:8])
}

// MarshalPrivateKeyPEM encodes a PKCS8 private key as PEM.
func MarshalPrivateKeyPEM(privateKey []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKey,
	})
}

// MarshalPublicKeyPEM encodes a PKIX public key as PEM.
func MarshalPublicKeyPEM(publicKey []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKey,
	})
}

// ParsePEMKey parses a PEM-encoded key and returns the raw bytes, whether it's private, and the algorithm.
func ParsePEMKey(pemData []byte) (keyBytes []byte, isPrivate bool, algorithm KeyAlgorithm, err error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, false, "", fmt.Errorf("failed to decode PEM block")
	}

	switch block.Type {
	case "PRIVATE KEY":
		privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, false, "", fmt.Errorf("parse private key: %w", err)
		}

		alg, err := detectAlgorithmFromPrivateKey(privKey)
		if err != nil {
			return nil, false, "", err
		}

		return block.Bytes, true, alg, nil

	case "PUBLIC KEY":
		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, false, "", fmt.Errorf("parse public key: %w", err)
		}

		alg, err := detectAlgorithmFromPublicKey(pubKey)
		if err != nil {
			return nil, false, "", err
		}

		return block.Bytes, false, alg, nil

	default:
		return nil, false, "", fmt.Errorf("unsupported PEM type: %s", block.Type)
	}
}

// ExtractPublicKeyFromPrivate derives the public key bytes from a PKCS8 private key.
func ExtractPublicKeyFromPrivate(privateKeyBytes []byte, algorithm KeyAlgorithm) ([]byte, error) {
	privKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	signer, ok := privKey.(crypto.Signer)
	if !ok {
		return nil, fmt.Errorf("private key does not implement crypto.Signer")
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(signer.Public())
	if err != nil {
		return nil, fmt.Errorf("marshal public key: %w", err)
	}

	return pubBytes, nil
}

func detectAlgorithmFromPrivateKey(key any) (KeyAlgorithm, error) {
	switch key.(type) {
	case ed25519.PrivateKey:
		return AlgorithmEd25519, nil
	case *ecdsa.PrivateKey:
		return AlgorithmECDSAP256, nil
	default:
		return "", fmt.Errorf("unsupported private key type: %T", key)
	}
}

func detectAlgorithmFromPublicKey(key any) (KeyAlgorithm, error) {
	switch key.(type) {
	case ed25519.PublicKey:
		return AlgorithmEd25519, nil
	case *ecdsa.PublicKey:
		return AlgorithmECDSAP256, nil
	default:
		return "", fmt.Errorf("unsupported public key type: %T", key)
	}
}
