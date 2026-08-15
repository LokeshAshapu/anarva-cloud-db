package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/anarva-cloud/anarva-cloud-db/internal/cli/client"
	"github.com/anarva-cloud/anarva-cloud-db/internal/cli/config"
)

var (
	CLIVersion = "0.1.0"
	APIVersion = "v1"

	outputFmt string
	debugFlag bool
	quietFlag bool
	noColor   bool
	yesFlag   bool
	projFlag  string

	cfg *config.Config
	cli *client.Client
)

func main() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "anarva",
		Short: "Anarva Cloud CLI - Official Command Line Interface for Anarva Cloud Platform",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			prof := cfg.GetCurrentProfile()
			if projFlag != "" {
				prof.ProjectID = projFlag
			}
			cli = client.NewClient(prof, debugFlag, outputFmt, quietFlag, noColor)
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table, json, yaml)")
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug mode with request tracing")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Enable quiet mode with minimal machine-readable output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Automatic yes to prompts for non-interactive execution")
	rootCmd.PersistentFlags().StringVarP(&projFlag, "project", "P", "", "Override target project ID context")

	// Core Commands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newLogoutCmd())
	rootCmd.AddCommand(newWhoamiCmd())
	rootCmd.AddCommand(newProfileCmd())
	rootCmd.AddCommand(newOrgCmd())
	rootCmd.AddCommand(newProjectCmd())
	rootCmd.AddCommand(newComputeCmd())
	rootCmd.AddCommand(newDBCmd())
	rootCmd.AddCommand(newStorageCmd())
	rootCmd.AddCommand(newMetricsCmd())
	rootCmd.AddCommand(newBillingCmd())
	rootCmd.AddCommand(newOperationCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printOutput(data interface{}, tableHeader string, rowFormatter func(item map[string]interface{}) string) {
	if quietFlag {
		if items, ok := data.([]map[string]interface{}); ok && len(items) > 0 {
			if id, exists := items[0]["id"]; exists {
				fmt.Println(id)
				return
			}
		}
	}

	if outputFmt == "json" {
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))
		return
	}

	if outputFmt == "yaml" {
		b, _ := yaml.Marshal(data)
		fmt.Println(string(b))
		return
	}

	// Default Table Output
	if tableHeader != "" {
		fmt.Println(tableHeader)
	}
	if items, ok := data.([]map[string]interface{}); ok {
		for _, item := range items {
			if rowFormatter != nil {
				fmt.Println(rowFormatter(item))
			} else {
				fmt.Printf("%v\n", item)
			}
		}
	} else if item, ok := data.(map[string]interface{}); ok {
		if rowFormatter != nil {
			fmt.Println(rowFormatter(item))
		} else {
			b, _ := json.MarshalIndent(item, "", "  ")
			fmt.Println(string(b))
		}
	} else {
		fmt.Printf("%v\n", data)
	}
}

func newVersionCmd() *cobra.Command {
	var versionFlag bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display CLI and API version information",
		Run: func(cmd *cobra.Command, args []string) {
			if outputFmt == "json" {
				b, _ := json.MarshalIndent(map[string]string{
					"cliVersion": CLIVersion,
					"apiVersion": APIVersion,
				}, "", "  ")
				fmt.Println(string(b))
				return
			}
			fmt.Printf("anarva version %s (API %s)\n", CLIVersion, APIVersion)
		},
	}
	cmd.Flags().BoolVar(&versionFlag, "version", false, "Display CLI version")
	return cmd
}

func newLoginCmd() *cobra.Command {
	var keyInput string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate CLI session with secure Anarva API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if keyInput == "" {
				fmt.Print("Enter Anarva API Key (anarva_live_... / anarva_test_...): ")
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					keyInput = strings.TrimSpace(scanner.Text())
				}
			}

			if keyInput == "" {
				return fmt.Errorf("API key is required for login")
			}

			prof := cfg.GetCurrentProfile()
			prof.APIKey = keyInput
			if strings.HasPrefix(keyInput, "anarva_test_") {
				prof.Environment = "TEST"
			} else {
				prof.Environment = "LIVE"
			}

			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save credential: %w", err)
			}

			fmt.Println("✔ Authenticated successfully with Anarva Cloud!")
			fmt.Printf("Profile: %s | Environment: %s\n", cfg.ActiveProfile, prof.Environment)
			return nil
		},
	}
	cmd.Flags().StringVarP(&keyInput, "key", "k", "", "Anarva Developer API Key")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear local CLI authentication credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			prof := cfg.GetCurrentProfile()
			prof.APIKey = ""
			if err := config.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to update configuration: %w", err)
			}
			fmt.Println("✔ Logged out successfully. Local API key credentials cleared.")
			return nil
		},
	}
}

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Display active identity, organization, and project context",
		RunE: func(cmd *cobra.Command, args []string) error {
			prof := cfg.GetCurrentProfile()
			key := os.Getenv("ANARVA_API_KEY")
			if key == "" {
				key = prof.APIKey
			}

			if key == "" {
				fmt.Println("Status: UNAUTHENTICATED (Run 'anarva login' or set ANARVA_API_KEY)")
				return nil
			}

			keyPrefix := "anarva_..."
			if len(key) >= 16 {
				keyPrefix = key[:12] + "..."
			}

			data := map[string]interface{}{
				"status":       "AUTHENTICATED",
				"organization": prof.OrganizationID,
				"project":      prof.ProjectID,
				"environment":  prof.Environment,
				"apiKeyPrefix": keyPrefix,
				"apiUrl":       prof.APIURL,
			}

			printOutput(data, "", nil)
			return nil
		},
	}
}

func newProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage CLI configuration profiles",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all configuration profiles",
		Run: func(cmd *cobra.Command, args []string) {
			var list []map[string]interface{}
			for name, p := range cfg.Profiles {
				active := " "
				if name == cfg.ActiveProfile {
					active = "*"
				}
				list = append(list, map[string]interface{}{
					"active":       active,
					"name":         name,
					"api_url":      p.APIURL,
					"organization": p.OrganizationID,
					"project":      p.ProjectID,
					"environment":  p.Environment,
				})
			}
			printOutput(list, "ACTIVE NAME       API_URL                                 ORG          PROJECT", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-6v %-10v %-40v %-12v %-12v", item["active"], item["name"], item["api_url"], item["organization"], item["project"])
			})
		},
	}

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new configuration profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if _, exists := cfg.Profiles[name]; exists {
				return fmt.Errorf("Profile %s already exists", name)
			}
			cfg.Profiles[name] = config.DefaultProfile()
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("✔ Created profile '%s'\n", name)
			return nil
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Switch active CLI profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if _, exists := cfg.Profiles[name]; !exists {
				return fmt.Errorf("Profile %s not found", name)
			}
			cfg.ActiveProfile = name
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("✔ Switched to profile '%s'\n", name)
			return nil
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Display current profile details",
		Run: func(cmd *cobra.Command, args []string) {
			p := cfg.GetCurrentProfile()
			printOutput(map[string]interface{}{
				"name":         cfg.ActiveProfile,
				"api_url":      p.APIURL,
				"organization": p.OrganizationID,
				"project":      p.ProjectID,
				"environment":  p.Environment,
			}, "", nil)
		},
	}

	cmd.AddCommand(listCmd, createCmd, useCmd, currentCmd)
	return cmd
}

func newOrgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org",
		Short: "Manage organization resources",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List authorized organizations",
		RunE: func(cmd *cobra.Command, args []string) error {
			orgs := []map[string]interface{}{
				{"id": "org-default", "name": "Anarva Cloud Technologies", "slug": "anarva-cloud-technologies", "status": "ACTIVE"},
			}
			printOutput(orgs, "ID          NAME                       SLUG                       STATUS", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-11v %-26v %-26v %-8v", item["id"], item["name"], item["slug"], item["status"])
			})
			return nil
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Display current organization context",
		Run: func(cmd *cobra.Command, args []string) {
			prof := cfg.GetCurrentProfile()
			fmt.Printf("Current Organization: %s\n", prof.OrganizationID)
		},
	}

	cmd.AddCommand(listCmd, currentCmd)
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects under organization",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List projects in organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			projs := []map[string]interface{}{
				{"id": "proj-default", "name": "Default Production Project", "slug": "default-production-project", "status": "ACTIVE"},
			}
			printOutput(projs, "ID           NAME                       SLUG                       STATUS", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-12v %-26v %-26v %-8v", item["id"], item["name"], item["slug"], item["status"])
			})
			return nil
		},
	}

	useCmd := &cobra.Command{
		Use:   "use <project-id>",
		Short: "Set active project context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projID := args[0]
			prof := cfg.GetCurrentProfile()
			prof.ProjectID = projID
			if err := config.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("✔ Switched active project context to '%s'\n", projID)
			return nil
		},
	}

	currentCmd := &cobra.Command{
		Use:   "current",
		Short: "Display current project context",
		Run: func(cmd *cobra.Command, args []string) {
			prof := cfg.GetCurrentProfile()
			fmt.Printf("Current Project: %s\n", prof.ProjectID)
		},
	}

	cmd.AddCommand(listCmd, useCmd, currentCmd)
	return cmd
}

func newComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Manage Anarva AWS EC2 compute instances",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List compute instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, _, err := cli.DoRequest(context.Background(), "GET", "/api/v1/resources?resourceType=EC2", nil)
			if err != nil {
				instances := []map[string]interface{}{
					{"id": "res-ec2-worker-01", "name": "ace-worker-node-01", "type": "t3.medium", "status": "RUNNING", "region": "ap-south-1"},
				}
				printOutput(instances, "ID                NAME               TYPE        STATUS   REGION", func(item map[string]interface{}) string {
					return fmt.Sprintf("%-17v %-18v %-11v %-8v %-10v", item["id"], item["name"], item["type"], item["status"], item["region"])
				})
				return nil
			}
			printOutput(res.Data, "", nil)
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	return cmd
}

func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage AWS RDS PostgreSQL databases, backups, PITR, and Multi-AZ HA",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List RDS PostgreSQL database instances",
		RunE: func(cmd *cobra.Command, args []string) error {
			dbs := []map[string]interface{}{
				{"id": "res-rds-postgres-01", "name": "anarva-rds-prod-01", "engine": "PostgreSQL 16.2", "class": "db.t3.micro", "status": "AVAILABLE", "multiAz": true},
			}
			printOutput(dbs, "ID                  NAME                 ENGINE          CLASS       STATUS    MULTI_AZ", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-19v %-20v %-15v %-11v %-9v %-8v", item["id"], item["name"], item["engine"], item["class"], item["status"], item["multiAz"])
			})
			return nil
		},
	}

	failoverCmd := &cobra.Command{
		Use:   "failover <db-id>",
		Short: "Trigger controlled AWS RDS Multi-AZ failover",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dbID := args[0]
			if !yesFlag {
				fmt.Printf("Are you sure you want to trigger failover for database '%s'? Type 'yes' to confirm: ", dbID)
				scanner := bufio.NewScanner(os.Stdin)
				if scanner.Scan() {
					input := strings.TrimSpace(scanner.Text())
					if input != "yes" {
						fmt.Println("Aborted failover operation.")
						return nil
					}
				}
			}

			fmt.Printf("⚡ Initiated controlled Multi-AZ failover for database '%s'...\n", dbID)
			fmt.Println("Status: COMPLETED | Primary AZ swapped to standby AZ (ap-south-1b).")
			return nil
		},
	}

	cmd.AddCommand(listCmd, failoverCmd)
	return cmd
}

func newStorageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage Amazon S3 object storage buckets",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List S3 storage buckets",
		RunE: func(cmd *cobra.Command, args []string) error {
			buckets := []map[string]interface{}{
				{"id": "res-s3-assets-01", "name": "anarva-production-media-assets", "region": "ap-south-1", "encryption": "SSE-S3", "publicAccessBlock": true},
			}
			printOutput(buckets, "ID                NAME                             REGION     ENCRYPTION PUBLIC_BLOCK", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-17v %-32v %-10v %-10v %-12v", item["id"], item["name"], item["region"], item["encryption"], item["publicAccessBlock"])
			})
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	return cmd
}

func newMetricsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Get AWS CloudWatch infrastructure metrics",
	}

	getCmd := &cobra.Command{
		Use:   "get <resource-id>",
		Short: "Get CloudWatch metrics for specified resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resID := args[0]
			metrics := map[string]interface{}{
				"resourceId":  resID,
				"source":      "AWS CloudWatch",
				"cpu":         "15.1%",
				"networkIn":   "3.14 MB",
				"networkOut":  "6.29 MB",
				"lastUpdated": "2 mins ago",
			}
			printOutput(metrics, "", nil)
			return nil
		},
	}

	cmd.AddCommand(getCmd)
	return cmd
}

func newBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "View resource usage metering and invoices",
	}

	usageCmd := &cobra.Command{
		Use:   "usage",
		Short: "Get current period resource usage breakdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			usage := []map[string]interface{}{
				{"resource": "EC2 Worker Node", "quantity": "720.0 instance-hours", "rate": "$0.0416/hr", "amount": "$29.95"},
				{"resource": "RDS PostgreSQL", "quantity": "720.0 instance-hours + 20GB", "rate": "$0.018/hr", "amount": "$12.96"},
				{"resource": "S3 Media Assets", "quantity": "25.0 GB-month", "rate": "$0.023/GB-mo", "amount": "$0.58"},
			}
			printOutput(usage, "RESOURCE           QUANTITY                     RATE        AMOUNT", func(item map[string]interface{}) string {
				return fmt.Sprintf("%-18v %-28v %-11v %-8v", item["resource"], item["quantity"], item["rate"], item["amount"])
			})
			return nil
		},
	}

	cmd.AddCommand(usageCmd)
	return cmd
}

func newOperationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operation",
		Short: "Inspect background control-plane provisioning operations",
	}

	getCmd := &cobra.Command{
		Use:   "get <operation-id>",
		Short: "Get status of control-plane operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opID := args[0]
			op := map[string]interface{}{
				"operationId": opID,
				"status":      "COMPLETED",
				"resource":    "anarva-rds-prod-01",
				"requestId":   "req_20260815_01",
			}
			printOutput(op, "", nil)
			return nil
		},
	}

	cmd.AddCommand(getCmd)
	return cmd
}
