package instance

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"testing"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/stretchr/testify/require"
	"github.com/herumi/bls-eth-go-binary/bls"
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



func BenchmarkBLSVerification(b *testing.B) {
	// Initialize BLS library
	if err := bls.Init(bls.BLS12_381); err != nil {
		b.Fatal(err)
	}

	// Generate a BLS key pair
	secretKey := new(bls.SecretKey)
	secretKey.SetByCSPRNG()
	publicKey := secretKey.GetPublicKey()

	// Create a message to sign
	message := []byte("Hello, world!")

	// Sign the message
	signature := secretKey.Sign(string(message))

	// Reset the timer
	b.ResetTimer()

	// Perform signature verification in the benchmark
	for i := 0; i < b.N; i++ {
		if !signature.Verify(publicKey, string(message)) {
			b.Fatal("Signature verification failed")
		}
	}
}

func BenchmarkBLSSignature(b *testing.B) {
	// Initialize BLS library
	if err := bls.Init(bls.BLS12_381); err != nil {
		b.Fatal(err)
	}

	// Generate a BLS key pair
	secretKey := new(bls.SecretKey)
	secretKey.SetByCSPRNG()

	// Create a message to sign
	message := []byte("Hello, world!")

	// Reset the timer
	b.ResetTimer()

	// Perform signature verification in the benchmark
	for i := 0; i < b.N; i++ {
		// Sign the message
		secretKey.Sign(string(message))
	}
}


func BenchmarkBLSAggregate(b *testing.B) {

	inst := BaseInstance(1)
	inst2 := BaseInstance(2)
	inst3 := BaseInstance(3)
	inst4 := BaseInstance(4)

	inst.Start(TestingContent, TestingHeight)
	inst2.Start(TestingContent, TestingHeight)
	inst3.Start(TestingContent, TestingHeight)
	inst4.Start(TestingContent, TestingHeight)

	

	vcbc_ready_message_1_1 := VCBCReadyMessage(1, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_2 := VCBCReadyMessage(2, TestingHeight, 1, 1, TestingContentHash, inst2.State)
	vcbc_ready_message_1_3 := VCBCReadyMessage(3, TestingHeight, 1, 1, TestingContentHash, inst3.State)
	// vcbc_ready_message_1_4 := VCBCReadyMessage(4, TestingHeight, 1, 1, TestingContentHash, inst4.State)

	b.ResetTimer()

	// Perform signature verification in the benchmark
	for i := 0; i < b.N; i++ {
		_, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_1_1, vcbc_ready_message_1_2, vcbc_ready_message_1_3})
		if err != nil {
			panic(err)
		}
	}

}