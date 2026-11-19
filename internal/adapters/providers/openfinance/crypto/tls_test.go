package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestCertificate generates a temporary in-memory RSA key and self-signed X.509 certificate for testing
func generateTestCertificate(t *testing.T, notBefore, notAfter time.Time) ([]byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Midaz Open Finance Test"},
			CommonName:   "client.openfinance.test",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	return certPEM, keyPEM
}

// TestBuildMTLSClientSuccess tests creating an mTLS client with valid certificates
func TestBuildMTLSClientSuccess(t *testing.T) {
	notBefore := time.Now().Add(-1 * time.Hour)
	notAfter := time.Now().Add(24 * time.Hour)
	certPEM, keyPEM := generateTestCertificate(t, notBefore, notAfter)

	client, err := BuildMTLSClient(TLSConfigOptions{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
		Timeout: 5 * time.Second,
	})

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 5*time.Second, client.Timeout)
}

// TestBuildMTLSClientExpiredCertError tests error when certificate is already expired
func TestBuildMTLSClientExpiredCertError(t *testing.T) {
	notBefore := time.Now().Add(-48 * time.Hour)
	notAfter := time.Now().Add(-24 * time.Hour) // expired
	certPEM, keyPEM := generateTestCertificate(t, notBefore, notAfter)

	client, err := BuildMTLSClient(TLSConfigOptions{
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	})

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "mTLS certificate expired")
}

// TestBuildMTLSClientMissingCredentials tests error when neither file nor PEM is provided
func TestBuildMTLSClientMissingCredentials(t *testing.T) {
	client, err := BuildMTLSClient(TLSConfigOptions{})

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "no certificate or private key provided")
}
