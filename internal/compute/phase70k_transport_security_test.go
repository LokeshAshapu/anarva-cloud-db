package compute_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
)

func generateTestCA(t *testing.T) ([]byte, *rsa.PrivateKey, *x509.Certificate) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "ANARVA Test CA"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)

	pemBlock := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	parsedCert, _ := x509.ParseCertificate(certBytes)
	return pemBlock, priv, parsedCert
}

func generateSignedCert(t *testing.T, caCert *x509.Certificate, caPriv *rsa.PrivateKey, commonName string, dnsNames []string, isClient bool) ([]byte, []byte) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	extUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	if isClient {
		extUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: commonName},
		DNSNames:     dnsNames,
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  extUsage,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &priv.PublicKey, caPriv)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	return certPEM, keyPEM
}

func TestPhase70K_TransportSecurityHardening(t *testing.T) {
	caPEM, caPriv, caCert := generateTestCA(t)
	serverCertPEM, serverKeyPEM := generateSignedCert(t, caCert, caPriv, "localhost", []string{"localhost", "127.0.0.1"}, false)
	clientCertPEM, clientKeyPEM := generateSignedCert(t, caCert, caPriv, "anarva-gateway-client", nil, true)

	serverTLSCert, err := tls.X509KeyPair(serverCertPEM, serverKeyPEM)
	require.NoError(t, err)

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caPEM)

	// Create mTLS mock worker HTTPS server
	var receivedClientCommonName string
	var lastAuthToken string

	mtlsWorker := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastAuthToken = r.Header.Get("X-Anarva-Worker-Token")
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			receivedClientCommonName = r.TLS.PeerCertificates[0].Subject.CommonName
		}

		if r.URL.Path == "/v1/redirect-target" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"REDIRECTED"}`))
			return
		}
		if strings.Contains(r.URL.Path, "redirect") {
			http.Redirect(w, r, "/v1/redirect-target", http.StatusFound)
			return
		}

		if r.Method == "POST" && r.URL.Path == "/v1/workloads" {
			var req computeProvider.CreateWorkloadRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(computeProvider.WorkloadResponse{
				WorkloadID: "workload-" + req.WorkloadID,
				Status:     "RUNNING",
				Health:     "HEALTHY",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	mtlsWorker.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}
	mtlsWorker.StartTLS()
	defer mtlsWorker.Close()

	ctx := context.Background()
	testToken := "secret-mtls-service-token-70k"

	t.Run("1. Valid mTLS connection succeeds and presents client cert", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
			ServerName:    "localhost",
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		inst := &computeDomain.ComputeInstance{ID: "inst-mtls-1", ResourceID: "res-1"}
		created, createErr := prov.CreateInstance(ctx, inst)
		require.NoError(t, createErr)

		assert.Equal(t, "workload-inst-mtls-1", created.ProviderInstanceID)
		assert.Equal(t, "anarva-gateway-client", receivedClientCommonName, "Worker must receive valid client certificate CommonName")
		assert.Equal(t, testToken, lastAuthToken)
	})

	t.Run("2. Missing CA certificate rejection", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		inst := &computeDomain.ComputeInstance{ID: "inst-mtls-2"}
		_, createErr := prov.CreateInstance(ctx, inst)
		assert.Error(t, createErr)
		assert.Contains(t, createErr.Error(), "tls error")
	})

	t.Run("3. Invalid CA certificate PEM rejection", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     "INVALID_CA_PEM_DATA",
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		assert.Error(t, pErr)
		assert.Contains(t, pErr.Error(), "invalid CA certificate")
	})

	t.Run("4. Untrusted worker server certificate rejection", func(t *testing.T) {
		otherCAPEM, _, _ := generateTestCA(t)
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(otherCAPEM), // CA that did not sign server cert
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		_, createErr := prov.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "inst-untrusted"})
		assert.Error(t, createErr)
		assert.Contains(t, createErr.Error(), "tls error")
	})

	t.Run("5. Worker hostname mismatch rejection", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
			ServerName:    "wrong-hostname.internal", // Mismatch
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		_, createErr := prov.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "inst-mismatch"})
		assert.Error(t, createErr)
		assert.Contains(t, createErr.Error(), "tls error")
	})

	t.Run("6. Missing client certificate in mTLS", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:    string(caPEM),
			ClientKeyPEM: string(clientKeyPEM),
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		assert.Error(t, pErr)
		assert.Contains(t, pErr.Error(), "both client certificate and client private key are required")
	})

	t.Run("7. Missing client private key in mTLS", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		assert.Error(t, pErr)
		assert.Contains(t, pErr.Error(), "both client certificate and client private key are required")
	})

	t.Run("8. Invalid client certificate PEM", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: "INVALID_CLIENT_CERT",
			ClientKeyPEM:  string(clientKeyPEM),
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		assert.Error(t, pErr)
	})

	t.Run("9. Invalid client private key PEM", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  "INVALID_CLIENT_KEY",
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		assert.Error(t, pErr)
	})

	t.Run("10. Worker rejecting unauthenticated client certificate", func(t *testing.T) {
		_, otherCAPriv, otherCACert := generateTestCA(t)
		badClientCert, badClientKey := generateSignedCert(t, otherCACert, otherCAPriv, "untrusted-client", nil, true)

		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM), // Valid server CA
			ClientCertPEM: string(badClientCert),
			ClientKeyPEM:  string(badClientKey),
			ServerName:    "localhost",
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		_, createErr := prov.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "inst-bad-client"})
		assert.Error(t, createErr)
		assert.Contains(t, createErr.Error(), "tls error")
	})

	t.Run("11. HTTPS Endpoint Success", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
			ServerName:    "localhost",
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)
		assert.True(t, strings.HasPrefix(mtlsWorker.URL, "https://"))

		inst := &computeDomain.ComputeInstance{ID: "https-success-node"}
		_, cErr := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr)

		err := prov.StartInstance(ctx, "https-success-node")
		assert.NoError(t, err)
	})

	t.Run("12. HTTP Endpoint Rejected when InsecureHTTP is false", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			InsecureHTTP: false,
		}
		_, pErr := computeProvider.NewRemoteComputeProviderWithTLS("http://worker.internal:8080", testToken, opts, nil)
		assert.ErrorIs(t, pErr, computeProvider.ErrInsecureEndpoint)
	})

	t.Run("13. HTTP Redirect Rejection & Credential Leakage Prevention", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
			ServerName:    "localhost",
		}
		prov, pErr := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		require.NoError(t, pErr)

		inst := &computeDomain.ComputeInstance{ID: "redirect-src", ProviderInstanceID: "redirect-src"}
		_, cErr := prov.CreateInstance(ctx, inst)
		require.NoError(t, cErr)

		// Trigger redirect endpoint
		_, err := prov.GetInstanceMetrics(ctx, "redirect-src")
		assert.ErrorIs(t, err, computeProvider.ErrRedirectProhibited)
	})

	t.Run("14. Token Remains Absent From Sanitized Error Outputs", func(t *testing.T) {
		opts := &computeProvider.TLSConfigOptions{
			CACertPEM:     string(caPEM),
			ClientCertPEM: string(clientCertPEM),
			ClientKeyPEM:  string(clientKeyPEM),
			ServerName:    "wrong-name",
		}
		prov, _ := computeProvider.NewRemoteComputeProviderWithTLS(mtlsWorker.URL, testToken, opts, nil)
		_, err := prov.CreateInstance(ctx, &computeDomain.ComputeInstance{ID: "err-test"})
		require.Error(t, err)
		assert.NotContains(t, err.Error(), testToken)
		assert.NotContains(t, err.Error(), string(clientKeyPEM))
	})
}
