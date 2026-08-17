export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://anarva-cloud-db-api.onrender.com'

export interface User {
  id: string
  email: string
  full_name: string
  role: string
  status: string
}

export interface Organization {
  id: string
  owner_id: string
  name: string
  slug: string
  created_at: string
}

export interface Project {
  id: string
  org_id: string
  name: string
  slug: string
  region: string
  max_databases: number
  max_storage_bytes: number
  created_at: string
}

export interface DatabaseInstance {
  id: string
  project_id: string
  name: string
  engine: string
  status: string
  host: string
  port: number
  db_name: string
  username: string
  storage_size_gb: number
  cpu_cores: number
  memory_mb: number
  created_at: string
}

export interface QueryResult {
  columns: Array<{ name: string; type: string }>
  rows: Array<Record<string, any>>
  rows_affected: number
  execution_time_ms: number
}

export interface BackupSnapshot {
  id: string
  database_id: string
  project_id: string
  name: string
  storage_path: string
  size_bytes: number
  backup_type: string
  status: string
  created_at: string
}

export function getAuthHeaders(): Record<string, string> {
  let token: string | null = null
  if (typeof window !== 'undefined') {
    token = localStorage.getItem('access_token') || localStorage.getItem('anarva_token') || localStorage.getItem('token')
  }
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  return headers
}

export async function fetchAPI<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const authHeaders = getAuthHeaders()

  const headers: Record<string, string> = {
    ...authHeaders,
    ...(options.headers as Record<string, string>),
  }

  const response = await fetch(`${API_BASE_URL}${endpoint}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `API request failed with status ${response.status}`)
  }

  return response.json()
}
