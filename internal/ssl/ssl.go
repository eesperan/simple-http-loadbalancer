package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"sync"

	"loadbalancer/internal/errors"
)

// Config holds SSL/TLS configuration
type Config struct {
	CertFile   string
	KeyFile    string
	CAFile     string // For client certificate validation
	ClientAuth tls.ClientAuthType
}

// Manager handles SSL/TLS configuration and certificate management
type Manager struct {
	mu              sync.RWMutex
	config          *Config
	tlsConfig       *tls.Config
	certReloadHook  func()
}

// New creates a new SSL manager
func New(config *Config) (*Manager, error) {
	if config == nil {
		return nil, errors.New(errors.ErrConfigInvalid, "SSL config is nil", nil)
	}

	manager := &Manager{
		config: config,
	}

	if err := manager.loadCertificates(); err != nil {
		return nil, err
	}

	return manager, nil
}

// loadCertificates loads and validates SSL certificates
func (m *Manager) loadCertificates() error {
	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return errors.New(errors.ErrSSLCertificate, "failed to load SSL certificate", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:  tls.VersionTLS12,
		ClientAuth:  m.config.ClientAuth,
	}

	// Load CA file if specified for client certificate validation
	if m.config.CAFile != "" {
		caData, err := ioutil.ReadFile(m.config.CAFile)
		if err != nil {
			return errors.New(errors.ErrSSLCertificate, "failed to read CA file", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caData) {
			return errors.New(errors.ErrSSLCertificate, "failed to parse CA certificate", nil)
		}

		tlsConfig.ClientCAs = certPool
	}

	m.mu.Lock()
	m.tlsConfig = tlsConfig
	m.mu.Unlock()

	return nil
}

// GetTLSConfig returns the current TLS configuration
func (m *Manager) GetTLSConfig() *tls.Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tlsConfig
}

// ReloadCertificates reloads certificates from disk
func (m *Manager) ReloadCertificates() error {
	if err := m.loadCertificates(); err != nil {
		return fmt.Errorf("failed to reload certificates: %v", err)
	}

	if m.certReloadHook != nil {
		m.certReloadHook()
	}

	return nil
}

// SetCertReloadHook sets a callback function to be called after certificate reload
func (m *Manager) SetCertReloadHook(hook func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.certReloadHook = hook
}

// EnableMutualTLS configures mutual TLS authentication
func (m *Manager) EnableMutualTLS(caFile string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.CAFile = caFile
	m.config.ClientAuth = tls.RequireAndVerifyClientCert

	return m.loadCertificates()
}

// DisableMutualTLS disables mutual TLS authentication
func (m *Manager) DisableMutualTLS() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config.CAFile = ""
	m.config.ClientAuth = tls.NoClientCert

	return m.loadCertificates()
}

// VerifyPeerCertificate provides custom certificate verification
func (m *Manager) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(rawCerts) == 0 {
		return errors.New(errors.ErrSSLCertificate, "no certificates provided", nil)
	}

	// Perform basic certificate parsing and validation
	_, err := x509.ParseCertificate(rawCerts[0])
	if err != nil {
		return errors.New(errors.ErrSSLCertificate, "failed to parse certificate", err)
	}

	// Additional custom verification can be added here
	// For example, checking certificate attributes, revocation status, etc.

	return nil
}

// UpdateCertificates updates the certificate and key files and reloads the configuration
func (m *Manager) UpdateCertificates(certFile, keyFile string) error {
	m.mu.Lock()
	m.config.CertFile = certFile
	m.config.KeyFile = keyFile
	m.mu.Unlock()

	return m.ReloadCertificates()
}
