package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// JWSSigner handles generation of FAPI-compliant signed JWTs (PS256) for client assertions and consents.
type JWSSigner struct {
	keyID      string
	clientID   string
	privateKey *rsa.PrivateKey
}

// NewJWSSigner creates a new JWS signer from an RSA private key PEM.
func NewJWSSigner(keyID, clientID string, keyPEM []byte) (*JWSSigner, error) {
	block, _ := pem.Decode(keyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing private key")
	}

	var privKey *rsa.PrivateKey
	var err error

	if block.Type == "RSA PRIVATE KEY" {
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else if block.Type == "PRIVATE KEY" {
		key, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", parseErr)
		}
		var ok bool
		privKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not RSA")
		}
	} else {
		return nil, fmt.Errorf("unsupported key type: %s", block.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key: %w", err)
	}

	return &JWSSigner{
		keyID:      keyID,
		clientID:   clientID,
		privateKey: privKey,
	}, nil
}

// GenerateClientAssertion creates a signed private_key_jwt for OAuth2 token endpoints.
func (s *JWSSigner) GenerateClientAssertion(tokenEndpoint string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()

	header := map[string]string{
		"alg": "PS256",
		"typ": "JWT",
		"kid": s.keyID,
	}

	claims := map[string]any{
		"iss": s.clientID,
		"sub": s.clientID,
		"aud": tokenEndpoint,
		"jti": uuid.New().String(),
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedClaims := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := encodedHeader + "." + encodedClaims
	hashed := sha256.Sum256([]byte(signingInput))

	// PS256 signing (RSASSA-PSS with SHA-256)
	signature, err := rsa.SignPSS(rand.Reader, s.privateKey, crypto.SHA256, hashed[:], &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	})
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + encodedSignature, nil
}

// VerifySignature validates a PS256 signature against an RSA public key.
func VerifySignature(jwtToken string, pubKey *rsa.PublicKey) error {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT token parts count: %d", len(parts))
	}

	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	hashed := sha256.Sum256([]byte(signingInput))

	return rsa.VerifyPSS(pubKey, crypto.SHA256, hashed[:], signature, &rsa.PSSOptions{
		SaltLength: rsa.PSSSaltLengthEqualsHash,
	})
}
