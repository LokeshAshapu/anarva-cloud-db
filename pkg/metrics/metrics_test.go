package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricsRecording(t *testing.T) {
	RecordHTTPRequest(200, "GET", "/api/v1/health", 0.045)
	RecordDatabaseQuery("SELECT", "success", 0.012)
	ActiveConnections.Inc()
	ActiveConnections.Dec()
}

func TestMetricsHandler(t *testing.T) {
	handler := Handler()
	assert.NotNil(t, handler)

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "anarva_http_requests_total")
}
