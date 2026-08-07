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
	apiURL   = "http://localhost:8080"
	jwtToken string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "anarva",
		Short: "Anarva Cloud DB CLI - Manage Cloud Databases, Execute Queries, and Trigger Snapshots",
	}

	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080", "API Gateway endpoint URL")
	rootCmd.PersistentFlags().StringVar(&jwtToken, "token", "", "JWT Access Token (or set ANARVA_TOKEN env var)")

	// Subcommands
	rootCmd.AddCommand(newLoginCmd())
	rootCmd.AddCommand(newDBCmd())
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
	var email, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Anarva Cloud DB platform",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Println("\nSet your token with: export ANARVA_TOKEN=<access_token>")
			return nil
		},
	}
	cmd.Flags().StringVarP(&email, "email", "e", "", "Account email address")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Account password")
	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func newDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Provision and manage cloud database instances",
	}

	var name, engine, projectID string
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Provision a new managed database instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload := map[string]interface{}{
				"project_id": projectID,
				"name":       name,
				"engine":     engine,
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest("POST", apiURL+"/api/v1/databases", bytes.NewBuffer(body))
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
			fmt.Printf("Status: %d\nResponse: %s\n", resp.StatusCode, string(respBytes))
			return nil
		},
	}
	createCmd.Flags().StringVarP(&name, "name", "n", "prod-db", "Database instance name")
	createCmd.Flags().StringVarP(&engine, "engine", "g", "postgres", "Database engine (postgres / mysql)")
	createCmd.Flags().StringVarP(&projectID, "project", "P", "proj-default", "Parent project ID")

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
