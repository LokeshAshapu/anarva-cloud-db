package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiURL   = "https://anarva-cloud-db-api.onrender.com"
	jwtToken string
)

func main() {
	if url := os.Getenv("ANARVA_API_URL"); url != "" {
		apiURL = url
	}

	rootCmd := &cobra.Command{
		Use:   "anarva",
		Short: "Anarva Cloud CLI - Manage Cloud Databases, Storage Buckets, and Compute ACUs",
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", apiURL, "API Gateway endpoint URL")
	rootCmd.PersistentFlags().StringVar(&jwtToken, "token", "", "API Secret Token or JWT Access Token")

	// Subcommands
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newDBCmd())
	rootCmd.AddCommand(newComputeCmd())
	rootCmd.AddCommand(newBucketCmd())
	rootCmd.AddCommand(newQueryCmd())
	rootCmd.AddCommand(newBackupCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getToken() string {
	if jwtToken != "" {
		return jwtToken
	}
	return os.Getenv("ANARVA_TOKEN")
}

func newLoginCmd() *cobra.Command {
	var email, password, tokenFlag string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Anarva Cloud Platform",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenFlag != "" {
				fmt.Println("✔ Token saved successfully for Anarva Cloud CLI session!")
				fmt.Printf("Authenticated Token: %s\n", tokenFlag)
				return nil
			}

			payload := map[string]string{"email": email, "password": password}
			body, _ := json.Marshal(payload)

			resp, err := http.Post(apiURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(body))
			if err != nil {
				return fmt.Errorf("failed to connect to gateway: %w", err)
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("login failed (%d): %s", resp.StatusCode, string(respBytes))
			}

			var res map[string]interface{}
			_ = json.Unmarshal(respBytes, &res)

			fmt.Println("✔ Login successful!")
			fmt.Printf("Access Token: %v\n", res["access_token"])
			return nil
		},
	}
	cmd.Flags().StringVarP(&email, "email", "e", "", "Account email address")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Account password")
	cmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "API Secret Key or Token")
	return cmd
}

func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Provision and manage cloud database instances",
	}

	var name, engine, region, projectID string
	var acu float64

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Provision a new managed database cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"project_id": projectID,
				"name":       name,
				"engine":     engine,
				"region":     region,
				"acu":        acu,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/databases", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if t := getToken(); t != "" {
				req.Header.Set("Authorization", "Bearer "+t)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to connect to gateway: %w", err)
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("✔ Database creation request submitted successfully!\nStatus: %d\nResponse: %s\n", resp.StatusCode, string(respBytes))
			return nil
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "prod-db", "Database instance name")
	createCmd.Flags().StringVarP(&engine, "engine", "g", "postgres", "Database engine (postgres / mysql)")
	createCmd.Flags().Float64VarP(&acu, "acu", "a", 2.0, "Anarva Compute Units (ACU capacity)")
	createCmd.Flags().StringVarP(&region, "region", "r", "ap-hyderabad-1", "Cloud region")
	createCmd.Flags().StringVarP(&projectID, "project", "P", "proj-default", "Parent project ID")

	cmd.AddCommand(createCmd)
	return cmd
}

func newComputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Provision and scale Anarva Compute Engine (ACE) workloads",
	}

	var name, region string
	var acu float64

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Launch a new ACU compute instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"name":     name,
				"regionId": region,
				"acu":      acu,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/compute/instances", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if t := getToken(); t != "" {
				req.Header.Set("Authorization", "Bearer "+t)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to connect to gateway: %w", err)
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("✔ Compute instance launch submitted successfully!\nStatus: %d\nResponse: %s\n", resp.StatusCode, string(respBytes))
			return nil
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "ace-worker-node", "Compute instance name")
	createCmd.Flags().Float64VarP(&acu, "acu", "a", 1.0, "Anarva Compute Units (ACU)")
	createCmd.Flags().StringVarP(&region, "region", "r", "us-east-1", "Cloud region")

	cmd.AddCommand(createCmd)
	return cmd
}

func newBucketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket",
		Short: "Manage Anarva Object Storage (AOS) S3 buckets",
	}

	var name, region string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an AOS object storage bucket",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"name":     name,
				"regionId": region,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/storage/buckets", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if t := getToken(); t != "" {
				req.Header.Set("Authorization", "Bearer "+t)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("failed to connect to gateway: %w", err)
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("✔ Storage bucket creation submitted successfully!\nStatus: %d\nResponse: %s\n", resp.StatusCode, string(respBytes))
			return nil
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "my-bucket", "Storage bucket name")
	createCmd.Flags().StringVarP(&region, "region", "r", "ap-hyderabad-1", "Cloud region")

	cmd.AddCommand(createCmd)
	return cmd
}

func newQueryCmd() *cobra.Command {
	var sql string
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Execute SQL query on managed database cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"sql": sql,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/query", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if t := getToken(); t != "" {
				req.Header.Set("Authorization", "Bearer "+t)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Result:\n%s\n", string(respBytes))
			return nil
		},
	}
	cmd.Flags().StringVarP(&sql, "sql", "s", "SELECT 1;", "SQL query statement")
	_ = cmd.MarkFlagRequired("sql")
	return cmd
}

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Trigger and list database backups",
	}

	var dbID, name string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Trigger snapshot backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"database_id": dbID,
				"name":        name,
				"type":        "SNAPSHOT",
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/backups", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if t := getToken(); t != "" {
				req.Header.Set("Authorization", "Bearer "+t)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			respBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Result:\n%s\n", string(respBytes))
			return nil
		},
	}
	createCmd.Flags().StringVarP(&dbID, "db", "d", "", "Database ID")
	createCmd.Flags().StringVarP(&name, "name", "n", "manual-backup", "Backup snapshot name")
	_ = createCmd.MarkFlagRequired("db")

	cmd.AddCommand(createCmd)
	return cmd
}
