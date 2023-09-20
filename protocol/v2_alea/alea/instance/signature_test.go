package instance

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEddsaSigningMock(t *testing.T) {
	MockEDDSASignature()
}
func TestEddsaVerificationMock(t *testing.T) {
	require.True(t, MockEDDSAVerification())
}

func BenchmarkEddsaSigningMock(t *testing.B) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		MockEDDSASignature()
	}
}
func BenchmarkEddsaVerificationMock(t *testing.B) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		MockEDDSAVerification()
	}
}

func TestRsaSigningMock(t *testing.T) {
	MockRSASigning()
}
func TestRsaVerificationMock(t *testing.T) {
	require.NoError(t, MockRSAVerification())
}

func BenchmarkRsaSigningMock(t *testing.B) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		MockRSASigning()
	}
}
func BenchmarkRsaVerificationMock(t *testing.B) {
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		MockRSAVerification()
	}
}

func TestGenerateEDDSA(t *testing.T) {
	// Generate a new Ed25519 private key
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Errorf("Error generating private key: %v", err)
		return
	}
	fmt.Printf("Mock private key: %x\n", privateKey)
	fmt.Printf("Mock public key: %x\n", publicKey)

	// Calculate the signature
	signature := ed25519.Sign(privateKey, []byte{1, 2, 3, 4})
	fmt.Printf("Mock signature: %x\n", signature)
}

func TestGenerateRSA(t *testing.T) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	publicKey := &privateKey.PublicKey
	fmt.Printf("Mock private key: %x\n", privateKey)
	fmt.Printf("Mock public key: %x\n", publicKey)

	hashed := sha256.Sum256([]byte{1, 2, 3, 4})
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		panic(err)
	}
	fmt.Printf("Mock signature: %x\n", signature)

	h := sha256.New()
	h.Write([]byte{1, 2, 3, 4})
	dig := h.Sum(nil)
	fmt.Printf("Mock dig: %x\n", dig)

	si, err := MockRSASigning()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Mock signature: %x\n", si)
}
