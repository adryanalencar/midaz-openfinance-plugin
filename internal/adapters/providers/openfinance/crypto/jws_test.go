package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateClientAssertionSuccess tests generating and verifying a signed private_key_jwt with PS256
func TestGenerateClientAssertionSuccess(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})

	signer, err := NewJWSSigner("key-test-01", "client-id-123", keyPEM)
	require.NoError(t, err)

	tokenEndpoint := "https://auth.banco.com.br/token"
	jwtAssertion, err := signer.GenerateClientAssertion(tokenEndpoint, 5*time.Minute)

	assert.NoError(t, err)
	assert.NotEmpty(t, jwtAssertion)

	// Verify signature using the public key
	err = VerifySignature(jwtAssertion, &priv.PublicKey)
	assert.NoError(t, err, "PS256 signature must be cryptographically valid")
}

// TestNewJWSSignerInvalidPEMError tests error handling on malformed private key
func TestNewJWSSignerInvalidPEMError(t *testing.T) {
	_, err := NewJWSSigner("key-01", "client-01", []byte("INVALID PEM CONTENT"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse PEM block")
}
