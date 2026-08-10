package organization

import (
	"fmt"
	"time"
)

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Store struct {
	orgs map[string]*Organization
}

func NewStore() *Store {
	now := time.Now()
	s := &Store{orgs: make(map[string]*Organization)}
	s.orgs["org-default"] = &Organization{
		ID:        "org-default",
		Name:      "Anarva Systems",
		Slug:      "anarva-systems",
		OwnerID:   "usr-default",
		CreatedAt: now.Add(-720 * time.Hour),
		UpdatedAt: now,
	}
	return s
}

func (s *Store) GetByID(id string) (*Organization, error) {
	org, ok := s.orgs[id]
	if !ok {
		return nil, fmt.Errorf("organization not found")
	}
	return org, nil
}

func (s *Store) List() []*Organization {
	var result []*Organization
	for _, o := range s.orgs {
		result = append(result, o)
	}
	return result
}
