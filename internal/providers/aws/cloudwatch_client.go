package aws

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type CloudWatchDatapoint struct {
	Timestamp time.Time
	Value     float64
	Unit      string
}

type CloudWatchMetricSeries struct {
	MetricName  string
	Namespace   string
	Unit        string
	Datapoints  []CloudWatchDatapoint
	Status      string // OK, NO_DATA, NOT_CONFIGURED, ACCESS_DENIED
	Source      string // AWS CloudWatch
	LastUpdated time.Time
}

type CloudWatchClient interface {
	VerifyConnectivity(ctx context.Context) error
	GetEC2Metrics(ctx context.Context, instanceID string, periodSec int) (map[string]*CloudWatchMetricSeries, error)
	GetRDSMetrics(ctx context.Context, dbInstanceID string, periodSec int) (map[string]*CloudWatchMetricSeries, error)
	GetS3Metrics(ctx context.Context, bucketName string, periodSec int) (map[string]*CloudWatchMetricSeries, error)
}

type MockCloudWatchClient struct {
	mu          sync.RWMutex
	isConnected bool
}

func NewMockCloudWatchClient(isConnected bool) *MockCloudWatchClient {
	return &MockCloudWatchClient{isConnected: isConnected}
}

func (m *MockCloudWatchClient) VerifyConnectivity(ctx context.Context) error {
	if !m.isConnected {
		return fmt.Errorf("AUTH_FAILED: Invalid AWS credentials for CloudWatch operation")
	}
	return nil
}

func (m *MockCloudWatchClient) GetEC2Metrics(ctx context.Context, instanceID string, periodSec int) (map[string]*CloudWatchMetricSeries, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API GetEC2Metrics unauthorized")
	}

	now := time.Now()
	res := make(map[string]*CloudWatchMetricSeries)

	// Real EC2 CloudWatch Metrics Series
	res["CPUUtilization"] = &CloudWatchMetricSeries{
		MetricName:  "CPUUtilization",
		Namespace:   "AWS/EC2",
		Unit:        "Percent",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-15 * time.Minute), Value: 12.4, Unit: "Percent"},
			{Timestamp: now.Add(-10 * time.Minute), Value: 14.8, Unit: "Percent"},
			{Timestamp: now.Add(-5 * time.Minute), Value: 18.2, Unit: "Percent"},
			{Timestamp: now, Value: 15.1, Unit: "Percent"},
		},
	}

	res["NetworkIn"] = &CloudWatchMetricSeries{
		MetricName:  "NetworkIn",
		Namespace:   "AWS/EC2",
		Unit:        "Bytes",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-15 * time.Minute), Value: 1048576, Unit: "Bytes"},
			{Timestamp: now.Add(-10 * time.Minute), Value: 2097152, Unit: "Bytes"},
			{Timestamp: now.Add(-5 * time.Minute), Value: 1572864, Unit: "Bytes"},
			{Timestamp: now, Value: 3145728, Unit: "Bytes"},
		},
	}

	res["NetworkOut"] = &CloudWatchMetricSeries{
		MetricName:  "NetworkOut",
		Namespace:   "AWS/EC2",
		Unit:        "Bytes",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-15 * time.Minute), Value: 524288, Unit: "Bytes"},
			{Timestamp: now.Add(-10 * time.Minute), Value: 8388608, Unit: "Bytes"},
			{Timestamp: now.Add(-5 * time.Minute), Value: 4194304, Unit: "Bytes"},
			{Timestamp: now, Value: 6291456, Unit: "Bytes"},
		},
	}

	return res, nil
}

func (m *MockCloudWatchClient) GetRDSMetrics(ctx context.Context, dbInstanceID string, periodSec int) (map[string]*CloudWatchMetricSeries, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API GetRDSMetrics unauthorized")
	}

	now := time.Now()
	res := make(map[string]*CloudWatchMetricSeries)

	res["CPUUtilization"] = &CloudWatchMetricSeries{
		MetricName:  "CPUUtilization",
		Namespace:   "AWS/RDS",
		Unit:        "Percent",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-15 * time.Minute), Value: 8.5, Unit: "Percent"},
			{Timestamp: now.Add(-10 * time.Minute), Value: 11.2, Unit: "Percent"},
			{Timestamp: now.Add(-5 * time.Minute), Value: 9.8, Unit: "Percent"},
			{Timestamp: now, Value: 10.4, Unit: "Percent"},
		},
	}

	res["DatabaseConnections"] = &CloudWatchMetricSeries{
		MetricName:  "DatabaseConnections",
		Namespace:   "AWS/RDS",
		Unit:        "Count",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-15 * time.Minute), Value: 5.0, Unit: "Count"},
			{Timestamp: now.Add(-10 * time.Minute), Value: 8.0, Unit: "Count"},
			{Timestamp: now.Add(-5 * time.Minute), Value: 6.0, Unit: "Count"},
			{Timestamp: now, Value: 7.0, Unit: "Count"},
		},
	}

	res["FreeStorageSpace"] = &CloudWatchMetricSeries{
		MetricName:  "FreeStorageSpace",
		Namespace:   "AWS/RDS",
		Unit:        "Bytes",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now, Value: 18253611008, Unit: "Bytes"},
		},
	}

	return res, nil
}

func (m *MockCloudWatchClient) GetS3Metrics(ctx context.Context, bucketName string, periodSec int) (map[string]*CloudWatchMetricSeries, error) {
	if !m.isConnected {
		return nil, fmt.Errorf("AUTH_FAILED: AWS API GetS3Metrics unauthorized")
	}

	now := time.Now()
	res := make(map[string]*CloudWatchMetricSeries)

	res["BucketSizeBytes"] = &CloudWatchMetricSeries{
		MetricName:  "BucketSizeBytes",
		Namespace:   "AWS/S3",
		Unit:        "Bytes",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-24 * time.Hour), Value: 4294967296, Unit: "Bytes"},
		},
	}

	res["NumberOfObjects"] = &CloudWatchMetricSeries{
		MetricName:  "NumberOfObjects",
		Namespace:   "AWS/S3",
		Unit:        "Count",
		Status:      "OK",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints: []CloudWatchDatapoint{
			{Timestamp: now.Add(-24 * time.Hour), Value: 1420, Unit: "Count"},
		},
	}

	res["AllRequests"] = &CloudWatchMetricSeries{
		MetricName:  "AllRequests",
		Namespace:   "AWS/S3",
		Unit:        "Count",
		Status:      "NOT_CONFIGURED",
		Source:      "AWS CloudWatch",
		LastUpdated: now,
		Datapoints:  []CloudWatchDatapoint{},
	}

	return res, nil
}
