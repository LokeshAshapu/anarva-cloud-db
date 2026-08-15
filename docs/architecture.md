# ANARVA CLOUD — TERRAFORM PROVIDER ARCHITECTURE

This document describes the architectural principles and control-plane flow for the official **Anarva Terraform Provider (`anarva/anarva`)**.

---

## 🏛️ Control-Plane Architecture

```
+-------------------------------------------------------------+
|                      Terraform CLI                          |
|         (declarative HCL: main.tf, variables.tf)            |
+-------------------------------------------------------------+
                              |
                              v
+-------------------------------------------------------------+
|               Anarva Terraform Provider                     |
|           (terraform-provider-anarva v0.1.0)                |
+-------------------------------------------------------------+
                              |
                     REST API v1 (HTTPS)
                 Authorization: Bearer <API_KEY>
                              v
+-------------------------------------------------------------+
|               Anarva Control Plane Gateway                  |
|              (Authorization & IAM Service)                  |
+-------------------------------------------------------------+
                              |
                              v
+-------------------------------------------------------------+
|                 Anarva Provisioning Engine                  |
+-------------------------------------------------------------+
                              |
                              v
+-------------------------------------------------------------+
|                 AWS Infrastructure (EC2/RDS/S3)             |
+-------------------------------------------------------------+
```

---

## 🔒 Security & Isolation Directives

1. **NO Direct AWS Access**: The Terraform provider NEVER communicates directly with AWS API endpoints, NEVER accepts AWS credentials (`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`), and NEVER embeds the AWS SDK.
2. **Server-Side Credentials**: Anarva Cloud owns and manages cloud infrastructure credentials server-side.
3. **Secret Redaction**: API keys (`anarva_live_...`) and tokens are NEVER stored in plaintext in diagnostic logs or error outputs. If an error occurs, secrets are replaced with `[REDACTED_API_KEY]`.
4. **Drift Detection**: When `Read` queries the Anarva API and receives HTTP status `404 Not Found`, the resource is cleanly removed from Terraform state so Terraform can recreate it on subsequent `terraform apply`.
