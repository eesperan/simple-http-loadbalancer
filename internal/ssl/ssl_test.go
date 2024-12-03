package ssl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"
)

// Helper function to create test certificates
func createTestCertificates(t *testing.T) (certFile, keyFile, caFile string, cleanup func()) {
	// Create temporary files
	certFile = "test-cert.pem"
	keyFile = "test-key.pem"
	caFile = "test-ca.pem"

	// Generate CA key pair
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate CA key: %v", err)
	}

	// Create CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:             x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create CA certificate: %v", err)
	}

	// Save CA certificate
	caOut, err := os.Create(caFile)
	if err != nil {
		t.Fatalf("Failed to create CA file: %v", err)
	}
	pem.Encode(caOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})
	caOut.Close()

	// Generate server key pair
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate server key: %v", err)
	}

	// Create server certificate
	server := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost"},
		IPAddresses:  nil,
	}

	serverBytes, err := x509.CreateCertificate(rand.Reader, server, ca, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("Failed to create server certificate: %v", err)
	}

	// Save server certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		t.Fatalf("Failed to create cert file: %v", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: serverBytes})
	certOut.Close()

	// Save server key
	keyOut, err := os.Create(keyFile)
	if err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}
	pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})
	keyOut.Close()

	cleanup = func() {
		os.Remove(certFile)
		os.Remove(keyFile)
		os.Remove(caFile)
	}

	return
}

func TestSSLManager(t *testing.T) {
	certFile, keyFile, caFile, cleanup := createTestCertificates(t)
	defer cleanup()

	// Test basic SSL manager creation
	manager, err := New(&Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create SSL manager: %v", err)
	}

	if manager.GetTLSConfig() == nil {
		t.Error("Expected non-nil TLS config")
	}

	// Test mutual TLS
	err = manager.EnableMutualTLS(caFile)
	if err != nil {
		t.Errorf("Failed to enable mutual TLS: %v", err)
	}

	tlsConfig := manager.GetTLSConfig()
	if tlsConfig.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Error("Expected client certificate verification to be required")
	}

	// Test certificate reloading
	err = manager.ReloadCertificates()
	if err != nil {
		t.Errorf("Failed to reload certificates: %v", err)
	}

	// Test certificate update
	err = manager.UpdateCertificates(certFile, keyFile)
	if err != nil {
		t.Errorf("Failed to update certificates: %v", err)
	}

	// Test disabling mutual TLS
	err = manager.DisableMutualTLS()
	if err != nil {
		t.Errorf("Failed to disable mutual TLS: %v", err)
	}

	tlsConfig = manager.GetTLSConfig()
	if tlsConfig.ClientAuth != tls.NoClientCert {
		t.Error("Expected client certificate verification to be disabled")
	}
}

func TestSSLManagerEdgeCases(t *testing.T) {
	// Test nil config
	_, err := New(nil)
	if err == nil {
		t.Error("Expected error with nil config")
	}

	// Test invalid certificate path
	manager, err := New(&Config{
		CertFile: "nonexistent.pem",
		KeyFile:  "nonexistent.key",
	})
	if err == nil {
		t.Error("Expected error with invalid certificate paths")
	}

	// Test invalid CA file
	certFile, keyFile, _, cleanup := createTestCertificates(t)
	defer cleanup()

	manager, err = New(&Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create SSL manager: %v", err)
	}

	err = manager.EnableMutualTLS("nonexistent.pem")
	if err == nil {
		t.Error("Expected error with invalid CA file")
	}
}

func TestSSLManagerCertReloadHook(t *testing.T) {
	certFile, keyFile, _, cleanup := createTestCertificates(t)
	defer cleanup()

	manager, err := New(&Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create SSL manager: %v", err)
	}

	hookCalled := false
	manager.SetCertReloadHook(func() {
		hookCalled = true
	})

	err = manager.ReloadCertificates()
	if err != nil {
		t.Errorf("Failed to reload certificates: %v", err)
	}

	if !hookCalled {
		t.Error("Expected cert reload hook to be called")
	}
}

func TestSSLManagerConcurrency(t *testing.T) {
	certFile, keyFile, _, cleanup := createTestCertificates(t)
	defer cleanup()

	manager, err := New(&Config{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("Failed to create SSL manager: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = manager.GetTLSConfig()
			_ = manager.ReloadCertificates()
		}()
	}

	wg.Wait()
}
