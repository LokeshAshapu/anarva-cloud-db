package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/anarva-cloud/anarva-cloud-db/internal/terraform/client"
	"github.com/anarva-cloud/anarva-cloud-db/internal/terraform/provider"
)

var (
	version = "0.1.0"
)

func main() {
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, "Print provider version")
	flag.Parse()

	if versionFlag {
		fmt.Printf("terraform-provider-anarva v%s\n", version)
		os.Exit(0)
	}

	cfg := client.Config{
		APIKey:         os.Getenv("ANARVA_API_KEY"),
		APIURL:         os.Getenv("ANARVA_API_URL"),
		OrganizationID: os.Getenv("ANARVA_ORGANIZATION_ID"),
		ProjectID:      os.Getenv("ANARVA_PROJECT_ID"),
	}

	p := provider.NewProvider(cfg)
	_ = p

	fmt.Printf("Anarva Terraform Provider v%s initialized successfully.\n", version)
}
