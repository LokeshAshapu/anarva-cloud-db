import { AnarvaClientConfig, ResolvedClientConfig, resolveConfig } from './config.js';
import { AnarvaError } from './errors.js';
import * as Types from './types/index.js';

export interface RequestOptions {
  signal?: AbortSignal;
  idempotencyKey?: string;
}

export class AnarvaClient {
  private readonly config: ResolvedClientConfig;

  public readonly organizations: OrganizationsAPI;
  public readonly projects: ProjectsAPI;
  public readonly compute: ComputeAPI;
  public readonly databases: DatabasesAPI;
  public readonly storage: StorageAPI;
  public readonly metrics: MetricsAPI;
  public readonly billing: BillingAPI;
  public readonly operations: OperationsAPI;
  public readonly feedback: FeedbackAPI;
  public readonly networking: NetworkingAPI;

  constructor(config?: AnarvaClientConfig) {
    this.config = resolveConfig(config);

    this.organizations = new OrganizationsAPI(this);
    this.projects = new ProjectsAPI(this);
    this.compute = new ComputeAPI(this);
    this.databases = new DatabasesAPI(this);
    this.storage = new StorageAPI(this);
    this.metrics = new MetricsAPI(this);
    this.billing = new BillingAPI(this);
    this.operations = new OperationsAPI(this);
    this.feedback = new FeedbackAPI(this);
    this.networking = new NetworkingAPI(this);
  }

  public async request<T>(method: string, path: string, body?: unknown, options?: RequestOptions): Promise<{ data: T; requestId?: string }> {
    const url = `${this.config.apiUrl}${path}`;
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'User-Agent': 'Anarva-TypeScript-SDK/0.1.0',
    };

    if (this.config.apiKey) {
      headers['Authorization'] = `Bearer ${this.config.apiKey}`;
    }

    if (options?.idempotencyKey) {
      headers['Idempotency-Key'] = options.idempotencyKey;
    }

    if (this.config.debug) {
      console.warn(`[SDK DEBUG] ${method} ${url}`);
      if (this.config.apiKey) {
        console.warn(`[SDK DEBUG] Authorization: Bearer REDACTED`);
      }
    }

    let attempt = 0;
    const retryStatusCodes = new Set([408, 429, 502, 503, 504]);

    while (attempt <= this.config.maxRetries) {
      attempt++;
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

      const combinedSignal = options?.signal
        ? AbortSignal.any([options.signal, controller.signal])
        : controller.signal;

      try {
        const response = await fetch(url, {
          method,
          headers,
          body: body ? JSON.stringify(body) : undefined,
          signal: combinedSignal,
        });

        clearTimeout(timeoutId);
        const reqId = response.headers.get('x-request-id') || undefined;

        if (!response.ok) {
          if (attempt <= this.config.maxRetries && retryStatusCodes.has(response.status) && method === 'GET') {
            let backoffMs = Math.pow(2, attempt) * 100;
            const retryAfterHeader = response.headers.get('retry-after');
            if (retryAfterHeader) {
              const parsedSeconds = parseInt(retryAfterHeader, 10);
              if (!isNaN(parsedSeconds) && parsedSeconds > 0) {
                backoffMs = parsedSeconds * 1000;
              }
            }
            // Add full randomized jitter (0.5 to 1.5x multiplier)
            const jitter = 0.5 + Math.random();
            backoffMs = Math.floor(backoffMs * jitter);
            await new Promise((resolve) => setTimeout(resolve, backoffMs));
            continue;
          }

          let errData: any = {};
          try {
            errData = await response.json();
          } catch {
            // Ignore parse error
          }

          const errorObj = errData?.error || {};
          throw new AnarvaError({
            code: errorObj.code || `HTTP_${response.status}`,
            message: errorObj.message || `Request failed with status ${response.status}`,
            requestId: errorObj.requestId || reqId,
            status: response.status,
            details: errorObj.details,
          });
        }

        let resData: any = {};
        try {
          resData = await response.json();
        } catch {
          throw new AnarvaError({
            code: 'INVALID_API_RESPONSE',
            message: 'Failed to parse API response JSON payload',
            status: response.status,
            requestId: reqId,
          });
        }

        return {
          data: (resData.data ?? resData) as T,
          requestId: resData.requestId || reqId,
        };
      } catch (err: any) {
        clearTimeout(timeoutId);
        if (err instanceof AnarvaError) {
          throw err;
        }

        if (attempt <= this.config.maxRetries && method === 'GET' && err.name !== 'AbortError') {
          const backoffMs = Math.pow(2, attempt) * 100;
          await new Promise((resolve) => setTimeout(resolve, backoffMs));
          continue;
        }

        throw new AnarvaError({
          code: err.name === 'AbortError' ? 'REQUEST_ABORTED' : 'NETWORK_ERROR',
          message: err.message || 'Network request failed',
        });
      }
    }

    throw new AnarvaError({
      code: 'MAX_RETRIES_EXCEEDED',
      message: `Failed after ${this.config.maxRetries} retry attempts`,
    });
  }
}

export class OrganizationsAPI {
  constructor(private client: AnarvaClient) {}

  public async list(options?: RequestOptions): Promise<Types.Organization[]> {
    const res = await this.client.request<Types.Organization[]>('GET', '/api/v1/organizations', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }

  public async get(id: string, options?: RequestOptions): Promise<Types.Organization> {
    const res = await this.client.request<Types.Organization>('GET', `/api/v1/organizations/${id}`, undefined, options);
    return res.data;
  }
}

export class ProjectsAPI {
  constructor(private client: AnarvaClient) {}

  public async list(options?: RequestOptions): Promise<Types.Project[]> {
    const res = await this.client.request<Types.Project[]>('GET', '/api/v1/projects', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }

  public async get(id: string, options?: RequestOptions): Promise<Types.Project> {
    const res = await this.client.request<Types.Project>('GET', `/api/v1/projects/${id}`, undefined, options);
    return res.data;
  }
}

export class ComputeAPI {
  constructor(private client: AnarvaClient) {}

  public async list(options?: RequestOptions): Promise<Types.ComputeInstance[]> {
    const res = await this.client.request<Types.ComputeInstance[]>('GET', '/api/v1/resources?resourceType=EC2', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }

  public async create(params: Types.CreateComputeParams, options?: RequestOptions): Promise<Types.ComputeInstance> {
    const res = await this.client.request<Types.ComputeInstance>('POST', '/api/v1/compute/instances', params, options);
    return res.data;
  }
}

export class DatabasesAPI {
  public readonly backups: DatabaseBackupsAPI;
  public readonly ha: DatabaseHAAPI;

  constructor(private client: AnarvaClient) {
    this.backups = new DatabaseBackupsAPI(client);
    this.ha = new DatabaseHAAPI(client);
  }

  public async list(options?: RequestOptions): Promise<Types.DatabaseInstance[]> {
    const res = await this.client.request<Types.DatabaseInstance[]>('GET', '/api/v1/resources?resourceType=RDS', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }

  public async get(id: string, options?: RequestOptions): Promise<Types.DatabaseInstance> {
    const res = await this.client.request<Types.DatabaseInstance>('GET', `/api/v1/databases/${id}`, undefined, options);
    return res.data;
  }

  public async create(params: Types.CreateDatabaseParams, options?: RequestOptions): Promise<Types.DatabaseInstance> {
    const res = await this.client.request<Types.DatabaseInstance>('POST', '/api/v1/databases', params, options);
    return res.data;
  }

  public async failover(id: string, options?: RequestOptions): Promise<Types.FailoverJob> {
    const res = await this.client.request<Types.FailoverJob>('POST', `/api/v1/databases/${id}/failover`, {}, options);
    return res.data;
  }
}

export class DatabaseBackupsAPI {
  constructor(private client: AnarvaClient) {}

  public async list(resourceId: string, options?: RequestOptions): Promise<Types.DatabaseBackup[]> {
    const res = await this.client.request<Types.DatabaseBackup[]>('GET', `/api/v1/backups?resourceId=${resourceId}`, undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }

  public async create(resourceId: string, snapshotName: string, options?: RequestOptions): Promise<Types.DatabaseBackup> {
    const res = await this.client.request<Types.DatabaseBackup>('POST', '/api/v1/backups', { resourceId, snapshotName }, options);
    return res.data;
  }
}

export class DatabaseHAAPI {
  constructor(private client: AnarvaClient) {}

  public async get(resourceId: string, options?: RequestOptions): Promise<Types.MultiAZConfig> {
    const res = await this.client.request<Types.MultiAZConfig>('GET', `/api/v1/ha/${resourceId}`, undefined, options);
    return res.data;
  }
}

export class StorageAPI {
  constructor(private client: AnarvaClient) {}

  public async list(options?: RequestOptions): Promise<Types.StorageBucket[]> {
    const res = await this.client.request<Types.StorageBucket[]>('GET', '/api/v1/resources?resourceType=S3', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }
}

export class MetricsAPI {
  constructor(private client: AnarvaClient) {}

  public async get(resourceId: string, options?: RequestOptions): Promise<Types.CloudWatchMetrics> {
    const res = await this.client.request<Types.CloudWatchMetrics>('GET', `/api/v1/metrics/${resourceId}`, undefined, options);
    return res.data;
  }
}

export class BillingAPI {
  constructor(private client: AnarvaClient) {}

  public async invoices(options?: RequestOptions): Promise<Types.BillingInvoice[]> {
    const res = await this.client.request<Types.BillingInvoice[]>('GET', '/api/v1/billing/invoices', undefined, options);
    return Array.isArray(res.data) ? res.data : [res.data];
  }
}

export class OperationsAPI {
  constructor(private client: AnarvaClient) {}

  public async get(id: string, options?: RequestOptions): Promise<Types.ControlPlaneOperation> {
    const res = await this.client.request<Types.ControlPlaneOperation>('GET', `/api/v1/operations/${id}`, undefined, options);
    return res.data;
  }

  public async wait(id: string, pollOptions?: { timeoutMs?: number; intervalMs?: number; signal?: AbortSignal }): Promise<Types.ControlPlaneOperation> {
    const timeoutMs = pollOptions?.timeoutMs ?? 60000;
    const intervalMs = pollOptions?.intervalMs ?? 2000;
    const startTime = Date.now();

    while (Date.now() - startTime < timeoutMs) {
      if (pollOptions?.signal?.aborted) {
        throw new AnarvaError({ code: 'WAIT_ABORTED', message: 'Operation polling was aborted' });
      }

      const op = await this.get(id, { signal: pollOptions?.signal });
      if (op.status === 'COMPLETED' || op.status === 'FAILED') {
        return op;
      }

      await new Promise((resolve) => setTimeout(resolve, intervalMs));
    }

    throw new AnarvaError({ code: 'WAIT_TIMEOUT', message: `Operation ${id} timed out after ${timeoutMs}ms` });
  }
}

export class FeedbackAPI {
  constructor(private client: AnarvaClient) {}

  public async create(payload: { rating: number; subject?: string; message: string; category?: string }, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('POST', '/api/v1/feedback', payload, options);
    return res.data;
  }

  public async list(query?: { page?: number; pageSize?: number; status?: string; minRating?: number }, options?: RequestOptions): Promise<any> {
    const params = new URLSearchParams();
    if (query?.page) params.set('page', query.page.toString());
    if (query?.pageSize) params.set('pageSize', query.pageSize.toString());
    if (query?.status) params.set('status', query.status);
    if (query?.minRating) params.set('minRating', query.minRating.toString());

    const qs = params.toString() ? `?${params.toString()}` : '';
    const res = await this.client.request<any>('GET', `/api/v1/feedback${qs}`, undefined, options);
    return res.data;
  }

  public async get(id: string, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('GET', `/api/v1/feedback/${id}`, undefined, options);
    return res.data;
  }

  public async updateStatus(id: string, status: string, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('PATCH', `/api/v1/feedback/${id}/status`, { status }, options);
    return res.data;
  }

  public async getAnalytics(options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('GET', '/api/v1/feedback/analytics', undefined, options);
    return res.data;
  }
}

export class NetworkingAPI {
  constructor(private client: AnarvaClient) {}

  public async listVpcs(query?: { organizationId?: string; projectId?: string }, options?: RequestOptions): Promise<any[]> {
    const params = new URLSearchParams();
    if (query?.organizationId) params.set('organizationId', query.organizationId);
    if (query?.projectId) params.set('projectId', query.projectId);
    const qs = params.toString() ? `?${params.toString()}` : '';
    const res = await this.client.request<any>('GET', `/api/v1/networks${qs}`, undefined, options);
    return res.data;
  }

  public async createVpc(payload: { name: string; cidr: string; regionId: string; organizationId?: string; projectId?: string }, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('POST', '/api/v1/networks', payload, options);
    return res.data;
  }

  public async getVpc(id: string, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('GET', `/api/v1/networks/${id}`, undefined, options);
    return res.data;
  }

  public async deleteVpc(id: string, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('DELETE', `/api/v1/networks/${id}`, undefined, options);
    return res.data;
  }

  public async listSubnets(vpcId?: string, options?: RequestOptions): Promise<any[]> {
    const qs = vpcId ? `?vpcId=${vpcId}` : '';
    const res = await this.client.request<any>('GET', `/api/v1/subnets${qs}`, undefined, options);
    return res.data;
  }

  public async createSubnet(payload: { vpcId: string; name: string; cidr: string; zone: string; type?: string; organizationId?: string; projectId?: string }, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('POST', '/api/v1/subnets', payload, options);
    return res.data;
  }

  public async listSecurityGroups(vpcId?: string, options?: RequestOptions): Promise<any[]> {
    const qs = vpcId ? `?vpcId=${vpcId}` : '';
    const res = await this.client.request<any>('GET', `/api/v1/security-groups${qs}`, undefined, options);
    return res.data;
  }

  public async createSecurityGroup(payload: { vpcId: string; name: string; description?: string; organizationId?: string; projectId?: string }, options?: RequestOptions): Promise<any> {
    const res = await this.client.request<any>('POST', '/api/v1/security-groups', payload, options);
    return res.data;
  }
}
