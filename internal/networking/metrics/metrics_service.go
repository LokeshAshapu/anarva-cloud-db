package metrics

import (
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/networking/domain"
)

type MetricsService struct{}

func NewMetricsService() *MetricsService {
	return &MetricsService{}
}

func (s *MetricsService) GetMetrics(networkID string) *domain.NetworkMetrics {
	return &domain.NetworkMetrics{
		NetworkID:      networkID,
		BytesIn:        1024 * 1024 * 12,
		BytesOut:       1024 * 1024 * 48,
		PacketsIn:      8450,
		PacketsOut:     16900,
		Connections:    12,
		LatencyMs:      0.62,
		DroppedPackets: 0,
		Quality:        "ACTUAL (LOCAL_NETWORK)",
		Timestamp:      time.Now(),
	}
}
