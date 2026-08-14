package aws

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAWSCloudWatch_EC2Metrics_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	cwClient := NewMockCloudWatchClient(true)
	metrics, err := cwClient.GetEC2Metrics(context.Background(), "i-0a8f9c1b2d3e4f5a6", 300)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	cpu, exists := metrics["CPUUtilization"]
	require.True(t, exists)
	assert.Equal(t, "AWS/EC2", cpu.Namespace)
	assert.Equal(t, "Percent", cpu.Unit)
	assert.Equal(t, "AWS CloudWatch", cpu.Source)
	assert.NotEmpty(t, cpu.Datapoints)
	assert.Greater(t, cpu.Datapoints[0].Value, 0.0)
}

func TestAWSCloudWatch_RDSMetrics_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	cwClient := NewMockCloudWatchClient(true)
	metrics, err := cwClient.GetRDSMetrics(context.Background(), "anarva-rds-prod-01", 300)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	conns, exists := metrics["DatabaseConnections"]
	require.True(t, exists)
	assert.Equal(t, "AWS/RDS", conns.Namespace)
	assert.Equal(t, "Count", conns.Unit)
	assert.Equal(t, "AWS CloudWatch", conns.Source)
	assert.NotEmpty(t, conns.Datapoints)
}

func TestAWSCloudWatch_S3Metrics_Success(t *testing.T) {
	os.Setenv("AWS_ENABLED", "true")
	defer os.Unsetenv("AWS_ENABLED")

	cwClient := NewMockCloudWatchClient(true)
	metrics, err := cwClient.GetS3Metrics(context.Background(), "anarva-production-media-assets", 86400)
	require.NoError(t, err)
	assert.NotNil(t, metrics)

	size, exists := metrics["BucketSizeBytes"]
	require.True(t, exists)
	assert.Equal(t, "AWS/S3", size.Namespace)
	assert.Equal(t, "Bytes", size.Unit)
	assert.Equal(t, "AWS CloudWatch", size.Source)

	// Verify unconfigured metric returns NOT_CONFIGURED, NOT fake 0
	requests, existsReq := metrics["AllRequests"]
	require.True(t, existsReq)
	assert.Equal(t, "NOT_CONFIGURED", requests.Status)
	assert.Empty(t, requests.Datapoints)
}

func TestAWSCloudWatch_DisabledMode(t *testing.T) {
	cwClient := NewMockCloudWatchClient(false)
	_, err := cwClient.GetEC2Metrics(context.Background(), "i-disabled", 300)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "AUTH_FAILED")
}
