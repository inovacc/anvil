package crypto

import (
	"bytes"
	"testing"
)

func TestMasterKeyToMnemonicRoundTrip(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	mnemonic, err := MasterKeyToMnemonic(key)
	if err != nil {
		t.Fatal(err)
	}

	recovered, err := MnemonicToMasterKey(mnemonic)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(key, recovered) {
		t.Fatal("round-trip failed: recovered key does not match original")
	}
}

func TestMasterKeyToMnemonic24Words(t *testing.T) {
	key := make([]byte, 32)

	mnemonic, err := MasterKeyToMnemonic(key)
	if err != nil {
		t.Fatal(err)
	}

	words := 0

	for _, c := range mnemonic {
		if c == ' ' {
			words++
		}
	}

	words++ // last word has no trailing space

	if words != 24 {
		t.Fatalf("expected 24 words, got %d", words)
	}
}

func TestMasterKeyToMnemonicBadSize(t *testing.T) {
	_, err := MasterKeyToMnemonic([]byte{1, 2, 3})
	if err == nil {
		t.Fatal("expected error for wrong key size")
	}
}

func TestMnemonicToMasterKeyInvalid(t *testing.T) {
	_, err := MnemonicToMasterKey("not a valid mnemonic phrase")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

func TestMnemonicHash(t *testing.T) {
	h1 := MnemonicHash("abandon about")
	h2 := MnemonicHash("abandon about")
	h3 := MnemonicHash("zoo wrong")

	if !bytes.Equal(h1, h2) {
		t.Fatal("same mnemonic should produce same hash")
	}

	if bytes.Equal(h1, h3) {
		t.Fatal("different mnemonics should produce different hashes")
	}

	if len(h1) != 32 {
		t.Fatalf("hash should be 32 bytes, got %d", len(h1))
	}
}
