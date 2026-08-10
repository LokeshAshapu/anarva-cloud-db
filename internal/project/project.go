package project

import (
	"fmt"
	"time"
)

type Project struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Environment    string    `json:"environment"`
	DefaultRegion  string    `json:"defaultRegion"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Store struct {
	projects map[string]*Project
}

func NewStore() *Store {
	now := time.Now()
	s := &Store{projects: make(map[string]*Project)}
	s.projects["proj-default"] = &Project{
		ID:             "proj-default",
		OrganizationID: "org-default",
		Name:           "Anarva Cloud Platform",
		Slug:           "anarva-cloud-platform",
		Description:    "Primary Cloud Infrastructure Project",
		Environment:    "Production",
		DefaultRegion:  "ap-hyderabad-1",
		CreatedAt:      now.Add(-720 * time.Hour),
		UpdatedAt:      now,
	}
	return s
}

func (s *Store) GetByID(id, orgID string) (*Project, error) {
	p, ok := s.projects[id]
	if !ok {
		return nil, fmt.Errorf("project not found")
	}
	if orgID != "" && p.OrganizationID != orgID {
		return nil, fmt.Errorf("access denied")
	}
	return p, nil
}

func (s *Store) List(orgID string) []*Project {
	var result []*Project
	for _, p := range s.projects {
		if orgID != "" && p.OrganizationID != orgID {
			continue
		}
		result = append(result, p)
	}
	return result
}
