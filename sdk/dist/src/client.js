import { resolveConfig } from './config.js';
import { AnarvaError } from './errors.js';
export class AnarvaClient {
    config;
    organizations;
    projects;
    compute;
    databases;
    storage;
    metrics;
    billing;
    operations;
    feedback;
    constructor(config) {
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
    }
    async request(method, path, body, options) {
        const url = `${this.config.apiUrl}${path}`;
        const headers = {
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
                    let errData = {};
                    try {
                        errData = await response.json();
                    }
                    catch {
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
                let resData = {};
                try {
                    resData = await response.json();
                }
                catch {
                    throw new AnarvaError({
                        code: 'INVALID_API_RESPONSE',
                        message: 'Failed to parse API response JSON payload',
                        status: response.status,
                        requestId: reqId,
                    });
                }
                return {
                    data: (resData.data ?? resData),
                    requestId: resData.requestId || reqId,
                };
            }
            catch (err) {
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
    client;
    constructor(client) {
        this.client = client;
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/organizations', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
    async get(id, options) {
        const res = await this.client.request('GET', `/api/v1/organizations/${id}`, undefined, options);
        return res.data;
    }
}
export class ProjectsAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/projects', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
    async get(id, options) {
        const res = await this.client.request('GET', `/api/v1/projects/${id}`, undefined, options);
        return res.data;
    }
}
export class ComputeAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/resources?resourceType=EC2', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
    async create(params, options) {
        const res = await this.client.request('POST', '/api/v1/compute/instances', params, options);
        return res.data;
    }
}
export class DatabasesAPI {
    client;
    backups;
    ha;
    constructor(client) {
        this.client = client;
        this.backups = new DatabaseBackupsAPI(client);
        this.ha = new DatabaseHAAPI(client);
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/resources?resourceType=RDS', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
    async get(id, options) {
        const res = await this.client.request('GET', `/api/v1/databases/${id}`, undefined, options);
        return res.data;
    }
    async create(params, options) {
        const res = await this.client.request('POST', '/api/v1/databases', params, options);
        return res.data;
    }
    async failover(id, options) {
        const res = await this.client.request('POST', `/api/v1/databases/${id}/failover`, {}, options);
        return res.data;
    }
}
export class DatabaseBackupsAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async list(resourceId, options) {
        const res = await this.client.request('GET', `/api/v1/backups?resourceId=${resourceId}`, undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
    async create(resourceId, snapshotName, options) {
        const res = await this.client.request('POST', '/api/v1/backups', { resourceId, snapshotName }, options);
        return res.data;
    }
}
export class DatabaseHAAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async get(resourceId, options) {
        const res = await this.client.request('GET', `/api/v1/ha/${resourceId}`, undefined, options);
        return res.data;
    }
}
export class StorageAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/resources?resourceType=S3', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
}
export class MetricsAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async get(resourceId, options) {
        const res = await this.client.request('GET', `/api/v1/metrics/${resourceId}`, undefined, options);
        return res.data;
    }
}
export class BillingAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async invoices(options) {
        const res = await this.client.request('GET', '/api/v1/billing/invoices', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
}
export class OperationsAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async get(id, options) {
        const res = await this.client.request('GET', `/api/v1/operations/${id}`, undefined, options);
        return res.data;
    }
    async wait(id, pollOptions) {
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
    client;
    constructor(client) {
        this.client = client;
    }
    async create(payload, options) {
        const res = await this.client.request('POST', '/api/v1/feedback', payload, options);
        return res.data;
    }
    async list(query, options) {
        const params = new URLSearchParams();
        if (query?.page)
            params.set('page', query.page.toString());
        if (query?.pageSize)
            params.set('pageSize', query.pageSize.toString());
        if (query?.status)
            params.set('status', query.status);
        if (query?.minRating)
            params.set('minRating', query.minRating.toString());
        const qs = params.toString() ? `?${params.toString()}` : '';
        const res = await this.client.request('GET', `/api/v1/feedback${qs}`, undefined, options);
        return res.data;
    }
    async get(id, options) {
        const res = await this.client.request('GET', `/api/v1/feedback/${id}`, undefined, options);
        return res.data;
    }
    async updateStatus(id, status, options) {
        const res = await this.client.request('PATCH', `/api/v1/feedback/${id}/status`, { status }, options);
        return res.data;
    }
    async getAnalytics(options) {
        const res = await this.client.request('GET', '/api/v1/feedback/analytics', undefined, options);
        return res.data;
    }
}
