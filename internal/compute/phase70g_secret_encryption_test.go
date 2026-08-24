package compute_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	computeDelivery "github.com/anarva-cloud/anarva-cloud-db/internal/compute/delivery/http"
	computeDomain "github.com/anarva-cloud/anarva-cloud-db/internal/compute/domain"
	computeProvider "github.com/anarva-cloud/anarva-cloud-db/internal/compute/provider"
	computeUseCase "github.com/anarva-cloud/anarva-cloud-db/internal/compute/usecase"
	mapping "github.com/anarva-cloud/anarva-cloud-db/internal/providers/mapping"
	"github.com/anarva-cloud/anarva-cloud-db/internal/security"
	"github.com/anarva-cloud/anarva-cloud-db/pkg/crypto"
)

func TestPhase70G_ComputeSecretEncryption(t *testing.T) {
	testKeyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 32 bytes hex
	cipher, err := crypto.NewAESGCMCipher(testKeyHex, "v1")
	require.NoError(t, err)
	crypto.SetGlobalCipher(cipher)

	t.Run("1. AES-256-GCM Encryption & Decryption Round Trip", func(t *testing.T) {
		plaintext := []byte(`{"NODE_ENV":"production","DB_PASSWORD":"supersecretpassword123"}`)
		ciphertext, encErr := cipher.Encrypt(plaintext)
		require.NoError(t, encErr)

		assert.True(t, strings.HasPrefix(ciphertext, "anarva:v1:"), "Ciphertext must have 'anarva:v1:' header prefix")

		decrypted, decErr := cipher.Decrypt(ciphertext)
		require.NoError(t, decErr)
		assert.Equal(t, string(plaintext), string(decrypted))
	})

	t.Run("2. Fresh Nonce Per Encryption (No Nonce Reuse)", func(t *testing.T) {
		plaintext := []byte(`{"SECRET_KEY":"sk_live_constant_val"}`)
		ct1, err1 := cipher.Encrypt(plaintext)
		require.NoError(t, err1)

		ct2, err2 := cipher.Encrypt(plaintext)
		require.NoError(t, err2)

		assert.NotEqual(t, ct1, ct2, "Encrypted outputs must differ due to fresh random nonce generation")
	})

	t.Run("3. Ciphertext Tampering Rejection", func(t *testing.T) {
		plaintext := []byte(`{"PORT":"8080"}`)
		ciphertext, _ := cipher.Encrypt(plaintext)

		// Tamper with the last hex character of the ciphertext payload
		tampered := ciphertext[:len(ciphertext)-1] + "f"
		if tampered == ciphertext {
			tampered = ciphertext[:len(ciphertext)-1] + "0"
		}

		_, decErr := cipher.Decrypt(tampered)
		assert.ErrorIs(t, decErr, crypto.ErrDecryptionFailed)
	})

	t.Run("4. Wrong-Key Decryption Rejection", func(t *testing.T) {
		wrongKeyHex := "9999999999abcdef9999999999abcdef9999999999abcdef9999999999abcdef"
		wrongCipher, _ := crypto.NewAESGCMCipher(wrongKeyHex, "v1")

		plaintext := []byte(`{"API_KEY":"secret"}`)
		ciphertext, _ := cipher.Encrypt(plaintext)

		_, decErr := wrongCipher.Decrypt(ciphertext)
		assert.ErrorIs(t, decErr, crypto.ErrDecryptionFailed)
	})

	t.Run("5. Malformed Ciphertext Rejection", func(t *testing.T) {
		malformedList := []string{
			"invalid-prefix:v1:nonce:ct",
			"anarva:v1:shortnonce:ct",
			"anarva:v1:badhexnonce:badhexct",
		}
		for _, m := range malformedList {
			_, decErr := cipher.Decrypt(m)
			assert.Error(t, decErr, "Malformed ciphertext '%s' must fail decryption", m)
		}
	})

	t.Run("6. Legacy Plaintext JSON Read & Dual-Read Fallback", func(t *testing.T) {
		legacyJSON := `{"PORT":"8080","ENV":"staging"}`
		decryptedBytes, decErr := cipher.Decrypt(legacyJSON)
		require.NoError(t, decErr)
		assert.Equal(t, legacyJSON, string(decryptedBytes), "Legacy plaintext JSON must be passed through directly on read")
	})

	t.Run("7. Legacy Plaintext Auto Re-Encryption on Save", func(t *testing.T) {
		inst := &computeDomain.ComputeInstance{
			ID:          "acu-inst-legacy-01",
			Name:        "legacy-node",
			EnvVarsJSON: `{"DB_USER":"root","DB_PASS":"legacy_pass"}`,
		}

		// Trigger AfterFind (Legacy JSON unmarshaled to EnvVars)
		errFind := inst.AfterFind(nil)
		require.NoError(t, errFind)
		assert.Equal(t, "legacy_pass", inst.EnvVars["DB_PASS"])

		// Trigger BeforeSave (EnvVars encrypted into anarva:v1: format)
		errSave := inst.BeforeSave(nil)
		require.NoError(t, errSave)
		assert.True(t, strings.HasPrefix(inst.EnvVarsJSON, "anarva:v1:"), "Legacy row must be re-encrypted to anarva:v1: on save")
	})

	t.Run("8. API EnvVars Secret Redaction in Public Responses", func(t *testing.T) {
		cRepo := newMemComputeRepo()
		mapRepo := mapping.NewInMemoryMappingRepository()
		prov := computeProvider.NewLocalDockerComputeProvider()
		uc := computeUseCase.NewComputeUseCase(cRepo, nil, prov)
		uc.SetMappingRepository(mapRepo)

		handler := computeDelivery.NewComputeHandler(uc, nil)
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux)

		// Create Instance with EnvVars via Usecase
		ctx := context.Background()
		inst, createErr := uc.CreateInstance(ctx, &computeDomain.ComputeInstance{
			ID:             "acu-inst-secret-70g",
			OrganizationID: "org-tenant-a",
			ProjectID:      "proj-tenant-a",
			Name:           "secret-worker-70g",
			ACU:            1.0,
			RegionID:       "us-east-1",
			EnvVars:        map[string]string{"SECRET_API_TOKEN": "shhh_super_secret_token_12345"},
		})
		require.NoError(t, createErr)

		// GET /api/v1/compute/instances/{id}
		req := httptest.NewRequest("GET", "/api/v1/compute/instances/"+inst.ID, nil)
		reqCtx := context.WithValue(req.Context(), security.OrgIDKey, "org-tenant-a")
		reqCtx = context.WithValue(reqCtx, security.ProjectIDKey, "proj-tenant-a")
		req = req.WithContext(reqCtx)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "shhh_super_secret_token_12345", "API response must NEVER contain unmasked secret tokens")
		assert.NotContains(t, rec.Body.String(), "SECRET_API_TOKEN", "API response must NOT include envVars map")
	})

	t.Run("9. Missing / Invalid Production Key Validation", func(t *testing.T) {
		_, errEmpty := crypto.NewAESGCMCipher("", "v1")
		assert.ErrorIs(t, errEmpty, crypto.ErrMissingEncryptionKey)

		_, errShort := crypto.NewAESGCMCipher("12345", "v1")
		assert.ErrorIs(t, errShort, crypto.ErrInvalidKeySize)
	})

	t.Run("10. Gateway Restart & Provider Re-Hydration with Encryption", func(t *testing.T) {
		ctx := context.Background()
		cRepo := newMemComputeRepo()
		mapRepo := mapping.NewInMemoryMappingRepository()
		prov1 := computeProvider.NewLocalDockerComputeProvider()
		uc1 := computeUseCase.NewComputeUseCase(cRepo, nil, prov1)
		uc1.SetMappingRepository(mapRepo)

		// Create instance
		inst, err := uc1.CreateInstance(ctx, &computeDomain.ComputeInstance{
			ID:             "acu-inst-restart-70g",
			OrganizationID: "org-tenant-a",
			ProjectID:      "proj-tenant-a",
			Name:           "restart-worker-70g",
			ACU:            1.0,
			RegionID:       "us-east-1",
			EnvVars:        map[string]string{"FOO": "BAR"},
		})
		require.NoError(t, err)

		// Simulate gateway process restart (erasing RAM map)
		prov2 := computeProvider.NewLocalDockerComputeProvider()
		uc2 := computeUseCase.NewComputeUseCase(cRepo, nil, prov2)
		uc2.SetMappingRepository(mapRepo)

		// Re-hydrate and fetch instance
		rehydrated, getErr := uc2.GetInstanceForTenant(ctx, "org-tenant-a", "proj-tenant-a", inst.ID)
		require.NoError(t, getErr)
		assert.Equal(t, inst.ID, rehydrated.ID)
		assert.Equal(t, "BAR", rehydrated.EnvVars["FOO"], "Decrypted envVars must survive process restart")
	})

	t.Run("11. Tenant Isolation Regression Verification", func(t *testing.T) {
		ctx := context.Background()
		cRepo := newMemComputeRepo()
		mapRepo := mapping.NewInMemoryMappingRepository()
		prov := computeProvider.NewLocalDockerComputeProvider()
		uc := computeUseCase.NewComputeUseCase(cRepo, nil, prov)
		uc.SetMappingRepository(mapRepo)

		inst, _ := uc.CreateInstance(ctx, &computeDomain.ComputeInstance{
			ID:             "acu-inst-tenant-70g",
			OrganizationID: "org-tenant-a",
			ProjectID:      "proj-tenant-a",
			Name:           "tenant-worker-70g",
			ACU:            1.0,
			RegionID:       "us-east-1",
			EnvVars:        map[string]string{"SECRET": "VAL"},
		})

		// Tenant B query attempt
		_, getErr := uc.GetInstanceForTenant(ctx, "org-tenant-b", "proj-tenant-b", inst.ID)
		require.Error(t, getErr)
		assert.Contains(t, getErr.Error(), "TENANT_ISOLATION_VIOLATION")
	})
}
