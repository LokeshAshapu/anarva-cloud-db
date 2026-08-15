terraform {
  required_providers {
    anarva = {
      source  = "anarva/anarva"
      version = "0.1.0"
    }
  }
}

provider "anarva" {
  # API key is automatically loaded from ANARVA_API_KEY environment variable
  organization_id = "org-default"
  project_id      = "proj-default"
}

# 1. Managed PostgreSQL Database Resource
resource "anarva_database" "production_db" {
  name       = "production-postgres-db"
  project_id = "proj-default"
  engine     = "POSTGRESQL"
  storage_gb = 50
  acu_units  = 2.0
  multi_az   = true
}

# 2. Compute ACU Instance Resource
resource "anarva_compute" "worker_node" {
  name       = "ace-worker-node-01"
  project_id = "proj-default"
  acu_units  = 1.0
  region_id  = "ap-south-1"
}

# 3. Object Storage S3 Bucket Resource
resource "anarva_storage_bucket" "media_assets" {
  name       = "anarva-production-media-assets"
  project_id = "proj-default"
  region_id  = "ap-south-1"
}

output "database_id" {
  value = anarva_database.production_db.id
}

output "database_status" {
  value = anarva_database.production_db.status
}
