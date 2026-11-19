package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"
)

// TLSConfigOptions holds paths and settings for building a compliant FAPI mTLS client.
type TLSConfigOptions struct {
	CertPath string
	KeyPath  string
	CertPEM  []byte
	KeyPEM   []byte
	Timeout  time.Duration
}

// BuildMTLSClient constructs an http.Client configured with mutual TLS and BACEN-approved cipher suites.
func BuildMTLSClient(opts TLSConfigOptions) (*http.Client, error) {
	var cert tls.Certificate
	var err error

	if len(opts.CertPEM) > 0 && len(opts.KeyPEM) > 0 {
		cert, err = tls.X509KeyPair(opts.CertPEM, opts.KeyPEM)
	} else if opts.CertPath != "" && opts.KeyPath != "" {
		certPEM, readErr := os.ReadFile(opts.CertPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read cert file: %w", readErr)
		}
		keyPEM, readErr := os.ReadFile(opts.KeyPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read key file: %w", readErr)
		}
		cert, err = tls.X509KeyPair(certPEM, keyPEM)
	} else {
		return nil, fmt.Errorf("no certificate or private key provided for mTLS")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse x509 key pair: %w", err)
	}

	// Validate certificate validity period
	if len(cert.Certificate) > 0 {
		parsedCert, parseErr := x509.ParseCertificate(cert.Certificate[0])
		if parseErr == nil {
			now := time.Now().UTC()
			if now.After(parsedCert.NotAfter) {
				return nil, fmt.Errorf("mTLS certificate expired on %s", parsedCert.NotAfter.Format(time.RFC3339))
			}
			if now.Before(parsedCert.NotBefore) {
				return nil, fmt.Errorf("mTLS certificate not valid until %s", parsedCert.NotBefore.Format(time.RFC3339))
			}
		}
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}
