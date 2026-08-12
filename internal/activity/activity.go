package activity

import (
	"sync"
	"time"
)

type EventAction string

const (
	ActionResourceCreated              EventAction = "RESOURCE_CREATED"
	ActionResourceUpdated              EventAction = "RESOURCE_UPDATED"
	ActionResourceDeleted              EventAction = "RESOURCE_DELETED"
	ActionResourceStarted              EventAction = "RESOURCE_STARTED"
	ActionResourceStopped              EventAction = "RESOURCE_STOPPED"
	ActionResourceConfigurationChanged EventAction = "RESOURCE_CONFIGURATION_CHANGED"
	ActionUserLogin                    EventAction = "USER_LOGIN"
	ActionAPIKeyCreated                EventAction = "API_KEY_CREATED"
	ActionBackupCreated                EventAction = "BACKUP_CREATED"
	ActionBackupCompleted              EventAction = "BACKUP_COMPLETED"
	ActionComputeCreated               EventAction = "COMPUTE_CREATED"
	ActionComputeStarted               EventAction = "COMPUTE_STARTED"
	ActionComputeStopped               EventAction = "COMPUTE_STOPPED"
	ActionComputeDeleted               EventAction = "COMPUTE_DELETED"
	ActionVolumeCreated                EventAction = "VOLUME_CREATED"
	ActionCommandExecuted               EventAction = "COMMAND_EXECUTED"
	ActionNetworkCreated               EventAction = "NETWORK_CREATED"
	ActionNetworkDeleted               EventAction = "NETWORK_DELETED"
	ActionSubnetCreated                EventAction = "SUBNET_CREATED"
	ActionSecurityGroupCreated         EventAction = "SECURITY_GROUP_CREATED"
	ActionIPAllocated                  EventAction = "IP_ALLOCATED"
	ActionDNSZoneCreated               EventAction = "DNS_ZONE_CREATED"
	ActionLoadBalancerCreated          EventAction = "LOAD_BALANCER_CREATED"
	ActionProvisioningPlanCreated      EventAction = "PROVISIONING_PLAN_CREATED"
	ActionProvisioningStarted          EventAction = "PROVISIONING_STARTED"
	ActionProvisioningCompleted        EventAction = "PROVISIONING_COMPLETED"
	ActionProvisioningFailed           EventAction = "PROVISIONING_FAILED"
	ActionRollbackStarted              EventAction = "ROLLBACK_STARTED"
	ActionRollbackCompleted            EventAction = "ROLLBACK_COMPLETED"
	ActionResourceDriftDetected        EventAction = "RESOURCE_DRIFT_DETECTED"
	ActionResourceReconciled           EventAction = "RESOURCE_RECONCILED"
)

type ActivityEvent struct {
	ID             string            `json:"id"`
	OrganizationID string            `json:"organizationId"`
	ProjectID      string            `json:"projectId"`
	ResourceID     string            `json:"resourceId,omitempty"`
	ActorID        string            `json:"actorId"`
	Action         EventAction       `json:"action"`
	Timestamp      time.Time         `json:"timestamp"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type Stream struct {
	mu     sync.RWMutex
	events []*ActivityEvent
}

func NewStream() *Stream {
	now := time.Now()
	s := &Stream{}
	s.events = []*ActivityEvent{
		{
			ID:             "act-101",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			ResourceID:     "arnv:db:ap-hyderabad-1:proj-default:database/production-db",
			ActorID:        "lokeshashapu@gmail.com",
			Action:         ActionResourceCreated,
			Timestamp:      now.Add(-2 * time.Hour),
			Metadata:       map[string]string{"type": "DATABASE", "name": "production-db"},
		},
		{
			ID:             "act-102",
			OrganizationID: "org-default",
			ProjectID:      "proj-default",
			ResourceID:     "arnv:s3:ap-hyderabad-1:proj-default:storage/anarva-media-assets",
			ActorID:        "lokeshashapu@gmail.com",
			Action:         ActionResourceCreated,
			Timestamp:      now.Add(-1 * time.Hour),
			Metadata:       map[string]string{"type": "STORAGE_BUCKET", "name": "anarva-media-assets"},
		},
	}
	return s
}

func (s *Stream) Record(event *ActivityEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	event.Timestamp = time.Now()
	s.events = append([]*ActivityEvent{event}, s.events...)
}

func (s *Stream) List(orgID string) []*ActivityEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*ActivityEvent
	for _, e := range s.events {
		if orgID != "" && e.OrganizationID != orgID {
			continue
		}
		result = append(result, e)
	}
	return result
}
