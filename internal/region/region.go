package region

type Region struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Location    string `json:"location"`
	Status      string `json:"status"` // AVAILABLE, COMING_SOON, MAINTENANCE
}

func ListRegions() []*Region {
	return []*Region{
		{ID: "ap-hyderabad-1", Name: "ap-hyderabad-1", DisplayName: "Asia Pacific — Hyderabad", Location: "Hyderabad", Status: "AVAILABLE"},
		{ID: "ap-mumbai-1", Name: "ap-mumbai-1", DisplayName: "Asia Pacific — Mumbai", Location: "Mumbai", Status: "AVAILABLE"},
		{ID: "ap-singapore-1", Name: "ap-singapore-1", DisplayName: "Asia Pacific — Singapore", Location: "Singapore", Status: "AVAILABLE"},
		{ID: "us-east-1", Name: "us-east-1", DisplayName: "US East — N. Virginia", Location: "Virginia", Status: "AVAILABLE"},
		{ID: "eu-west-1", Name: "eu-west-1", DisplayName: "Europe West — Frankfurt", Location: "Frankfurt", Status: "AVAILABLE"},
		{ID: "sa-east-1", Name: "sa-east-1", DisplayName: "South America — São Paulo", Location: "São Paulo", Status: "COMING_SOON"},
		{ID: "me-central-1", Name: "me-central-1", DisplayName: "Middle East — UAE", Location: "UAE", Status: "COMING_SOON"},
	}
}
