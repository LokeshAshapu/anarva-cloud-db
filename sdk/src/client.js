"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.OperationsAPI = exports.BillingAPI = exports.MetricsAPI = exports.StorageAPI = exports.DatabaseHAAPI = exports.DatabaseBackupsAPI = exports.DatabasesAPI = exports.ComputeAPI = exports.ProjectsAPI = exports.OrganizationsAPI = exports.AnarvaClient = void 0;
const config_js_1 = require("./config.js");
const errors_js_1 = require("./errors.js");
class AnarvaClient {
    config;
    organizations;
    projects;
    compute;
    databases;
    storage;
    metrics;
    billing;
    operations;
    constructor(config) {
        this.config = (0, config_js_1.resolveConfig)(config);
        this.organizations = new OrganizationsAPI(this);
        this.projects = new ProjectsAPI(this);
        this.compute = new ComputeAPI(this);
        this.databases = new DatabasesAPI(this);
        this.storage = new StorageAPI(this);
        this.metrics = new MetricsAPI(this);
        this.billing = new BillingAPI(this);
        this.operations = new OperationsAPI(this);
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
                        const backoffMs = Math.pow(2, attempt) * 100;
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
                    throw new errors_js_1.AnarvaError({
                        code: errorObj.code || `HTTP_${response.status}`,
                        message: errorObj.message || `Request failed with status ${response.status}`,
                        requestId: errorObj.requestId || reqId,
                        status: response.status,
                        details: errorObj.details,
                    });
                }
                const resData = (await response.json());
                return {
                    data: (resData.data ?? resData),
                    requestId: resData.requestId || reqId,
                };
            }
            catch (err) {
                clearTimeout(timeoutId);
                if (err instanceof errors_js_1.AnarvaError) {
                    throw err;
                }
                if (attempt <= this.config.maxRetries && method === 'GET' && err.name !== 'AbortError') {
                    const backoffMs = Math.pow(2, attempt) * 100;
                    await new Promise((resolve) => setTimeout(resolve, backoffMs));
                    continue;
                }
                throw new errors_js_1.AnarvaError({
                    code: err.name === 'AbortError' ? 'REQUEST_ABORTED' : 'NETWORK_ERROR',
                    message: err.message || 'Network request failed',
                });
            }
        }
        throw new errors_js_1.AnarvaError({
            code: 'MAX_RETRIES_EXCEEDED',
            message: `Failed after ${this.config.maxRetries} retry attempts`,
        });
    }
}
exports.AnarvaClient = AnarvaClient;
class OrganizationsAPI {
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
exports.OrganizationsAPI = OrganizationsAPI;
class ProjectsAPI {
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
exports.ProjectsAPI = ProjectsAPI;
class ComputeAPI {
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
exports.ComputeAPI = ComputeAPI;
class DatabasesAPI {
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
exports.DatabasesAPI = DatabasesAPI;
class DatabaseBackupsAPI {
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
exports.DatabaseBackupsAPI = DatabaseBackupsAPI;
class DatabaseHAAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async get(resourceId, options) {
        const res = await this.client.request('GET', `/api/v1/ha/${resourceId}`, undefined, options);
        return res.data;
    }
}
exports.DatabaseHAAPI = DatabaseHAAPI;
class StorageAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async list(options) {
        const res = await this.client.request('GET', '/api/v1/resources?resourceType=S3', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
}
exports.StorageAPI = StorageAPI;
class MetricsAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async get(resourceId, options) {
        const res = await this.client.request('GET', `/api/v1/metrics/${resourceId}`, undefined, options);
        return res.data;
    }
}
exports.MetricsAPI = MetricsAPI;
class BillingAPI {
    client;
    constructor(client) {
        this.client = client;
    }
    async invoices(options) {
        const res = await this.client.request('GET', '/api/v1/billing/invoices', undefined, options);
        return Array.isArray(res.data) ? res.data : [res.data];
    }
}
exports.BillingAPI = BillingAPI;
class OperationsAPI {
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
                throw new errors_js_1.AnarvaError({ code: 'WAIT_ABORTED', message: 'Operation polling was aborted' });
            }
            const op = await this.get(id, { signal: pollOptions?.signal });
            if (op.status === 'COMPLETED' || op.status === 'FAILED') {
                return op;
            }
            await new Promise((resolve) => setTimeout(resolve, intervalMs));
        }
        throw new errors_js_1.AnarvaError({ code: 'WAIT_TIMEOUT', message: `Operation ${id} timed out after ${timeoutMs}ms` });
    }
}
exports.OperationsAPI = OperationsAPI;
