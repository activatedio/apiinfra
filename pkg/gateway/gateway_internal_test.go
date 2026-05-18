package gateway

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// issueSelfSignedCert writes a self-signed cert + key pair into dir
// and returns (certPath, keyPath).
func issueSelfSignedCert(t *testing.T, dir, name string) (string, string) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)

	certPath := filepath.Join(dir, name+".crt")
	keyPath := filepath.Join(dir, name+".key")
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}), 0o600))
	return certPath, keyPath
}

func TestServerConfig_Addr(t *testing.T) {
	cases := []struct {
		name string
		cfg  ServerConfig
		want string
	}{
		{"host-and-port", ServerConfig{Host: "127.0.0.1", Port: 8080}, "127.0.0.1:8080"},
		{"all-interfaces", ServerConfig{Port: 80}, ":80"},
		{"negative-port-becomes-zero", ServerConfig{Host: "h", Port: -1}, "h:0"},
		{"zero-port-stays-zero", ServerConfig{Host: "h", Port: 0}, "h:0"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.cfg.Addr())
		})
	}
}

func TestResolveMTLS_Disabled(t *testing.T) {
	assert.False(t, resolveMTLS(MTLSDisabled, &ServerConfig{MTLS: true, TLS: true, TLSCAPath: "x"}))
}

func TestResolveMTLS_FromConfig(t *testing.T) {
	assert.True(t, resolveMTLS(MTLSFromConfig, &ServerConfig{MTLS: true}))
	assert.False(t, resolveMTLS(MTLSFromConfig, &ServerConfig{MTLS: false}))
}

// The zero value of MTLSMode is MTLSFromConfig so that
// ProvideServer() with no WithMTLS option is runtime-driven.
func TestMTLSMode_ZeroValueIsFromConfig(t *testing.T) {
	var zero MTLSMode
	assert.Equal(t, MTLSFromConfig, zero)
}

func TestResolveMTLS_Always_OK(t *testing.T) {
	assert.True(t, resolveMTLS(MTLSAlways, &ServerConfig{TLS: true, TLSCAPath: "ca.pem"}))
}

func TestResolveMTLS_Always_PanicsWithoutTLS(t *testing.T) {
	assert.PanicsWithValue(t, "gateway: MTLSAlways requires server.tls=true",
		func() { resolveMTLS(MTLSAlways, &ServerConfig{TLS: false, TLSCAPath: "ca.pem"}) })
}

func TestResolveMTLS_Always_PanicsWithoutCA(t *testing.T) {
	assert.PanicsWithValue(t, "gateway: MTLSAlways requires server.tlsCaPath",
		func() { resolveMTLS(MTLSAlways, &ServerConfig{TLS: true, TLSCAPath: ""}) })
}

func TestBuildServerTLS_ServerOnly(t *testing.T) {
	dir := t.TempDir()
	cert, key := issueSelfSignedCert(t, dir, "server")
	cfg := &ServerConfig{TLS: true, TLSCertPath: cert, TLSKeyPath: key}

	tlsCfg, err := buildServerTLS(cfg, false)
	require.NoError(t, err)
	require.NotNil(t, tlsCfg)
	assert.Len(t, tlsCfg.Certificates, 1)
	assert.Nil(t, tlsCfg.ClientCAs)
	assert.Equal(t, uint16(0), uint16(tlsCfg.ClientAuth))
	assert.Equal(t, []string{"h2", "http/1.1"}, tlsCfg.NextProtos)
}

func TestBuildServerTLS_MTLS(t *testing.T) {
	dir := t.TempDir()
	cert, key := issueSelfSignedCert(t, dir, "server")
	ca, _ := issueSelfSignedCert(t, dir, "ca")
	cfg := &ServerConfig{TLS: true, TLSCertPath: cert, TLSKeyPath: key, TLSCAPath: ca}

	tlsCfg, err := buildServerTLS(cfg, true)
	require.NoError(t, err)
	require.NotNil(t, tlsCfg.ClientCAs)
	assert.NotEmpty(t, tlsCfg.ClientCAs.Subjects()) //nolint:staticcheck // SA1019 acceptable in test
	assert.Equal(t, uint16(4), uint16(tlsCfg.ClientAuth))
}

func TestBuildServerTLS_MTLS_MissingCA(t *testing.T) {
	dir := t.TempDir()
	cert, key := issueSelfSignedCert(t, dir, "server")
	cfg := &ServerConfig{TLS: true, TLSCertPath: cert, TLSKeyPath: key}

	_, err := buildServerTLS(cfg, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tlsCaPath is empty")
}

func TestBuildServerTLS_BadCertPath(t *testing.T) {
	cfg := &ServerConfig{TLS: true, TLSCertPath: "/nope/cert.pem", TLSKeyPath: "/nope/key.pem"}
	_, err := buildServerTLS(cfg, false)
	require.Error(t, err)
}

func TestBuildServerTLS_BadCAPath(t *testing.T) {
	dir := t.TempDir()
	cert, key := issueSelfSignedCert(t, dir, "server")
	cfg := &ServerConfig{TLS: true, TLSCertPath: cert, TLSKeyPath: key, TLSCAPath: "/nope/ca.pem"}
	_, err := buildServerTLS(cfg, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read tlsCaPath")
}
