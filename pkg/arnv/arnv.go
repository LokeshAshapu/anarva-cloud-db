package arnv

import (
	"fmt"
	"strings"
)

// GenerateARNV constructs a unique, stable, non-sensitive Anarva Resource Identifier
// Example: arnv:db:ap-hyderabad-1:proj-default:database/production-db
func GenerateARNV(resourceType, regionID, projectID, resourceName string) string {
	typeSlug := strings.ToLower(resourceType)
	switch typeSlug {
	case "database":
		typeSlug = "db"
	case "storage_bucket", "storage":
		typeSlug = "s3"
	case "compute":
		typeSlug = "vm"
	case "network":
		typeSlug = "vpc"
	case "backup":
		typeSlug = "bak"
	case "replica":
		typeSlug = "rep"
	}

	cleanName := strings.ToLower(strings.ReplaceAll(resourceName, " ", "-"))
	return fmt.Sprintf("arnv:%s:%s:%s:%s/%s", typeSlug, regionID, projectID, resourceType, cleanName)
}
