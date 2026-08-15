import { AnarvaClientConfig } from './config.js';
import * as Types from './types/index.js';
export interface RequestOptions {
    signal?: AbortSignal;
    idempotencyKey?: string;
}
export declare class AnarvaClient {
    private readonly config;
    readonly organizations: OrganizationsAPI;
    readonly projects: ProjectsAPI;
    readonly compute: ComputeAPI;
    readonly databases: DatabasesAPI;
    readonly storage: StorageAPI;
    readonly metrics: MetricsAPI;
    readonly billing: BillingAPI;
    readonly operations: OperationsAPI;
    constructor(config?: AnarvaClientConfig);
    request<T>(method: string, path: string, body?: unknown, options?: RequestOptions): Promise<{
        data: T;
        requestId?: string;
    }>;
}
export declare class OrganizationsAPI {
    private client;
    constructor(client: AnarvaClient);
    list(options?: RequestOptions): Promise<Types.Organization[]>;
    get(id: string, options?: RequestOptions): Promise<Types.Organization>;
}
export declare class ProjectsAPI {
    private client;
    constructor(client: AnarvaClient);
    list(options?: RequestOptions): Promise<Types.Project[]>;
    get(id: string, options?: RequestOptions): Promise<Types.Project>;
}
export declare class ComputeAPI {
    private client;
    constructor(client: AnarvaClient);
    list(options?: RequestOptions): Promise<Types.ComputeInstance[]>;
    create(params: Types.CreateComputeParams, options?: RequestOptions): Promise<Types.ComputeInstance>;
}
export declare class DatabasesAPI {
    private client;
    readonly backups: DatabaseBackupsAPI;
    readonly ha: DatabaseHAAPI;
    constructor(client: AnarvaClient);
    list(options?: RequestOptions): Promise<Types.DatabaseInstance[]>;
    get(id: string, options?: RequestOptions): Promise<Types.DatabaseInstance>;
    create(params: Types.CreateDatabaseParams, options?: RequestOptions): Promise<Types.DatabaseInstance>;
    failover(id: string, options?: RequestOptions): Promise<Types.FailoverJob>;
}
export declare class DatabaseBackupsAPI {
    private client;
    constructor(client: AnarvaClient);
    list(resourceId: string, options?: RequestOptions): Promise<Types.DatabaseBackup[]>;
    create(resourceId: string, snapshotName: string, options?: RequestOptions): Promise<Types.DatabaseBackup>;
}
export declare class DatabaseHAAPI {
    private client;
    constructor(client: AnarvaClient);
    get(resourceId: string, options?: RequestOptions): Promise<Types.MultiAZConfig>;
}
export declare class StorageAPI {
    private client;
    constructor(client: AnarvaClient);
    list(options?: RequestOptions): Promise<Types.StorageBucket[]>;
}
export declare class MetricsAPI {
    private client;
    constructor(client: AnarvaClient);
    get(resourceId: string, options?: RequestOptions): Promise<Types.CloudWatchMetrics>;
}
export declare class BillingAPI {
    private client;
    constructor(client: AnarvaClient);
    invoices(options?: RequestOptions): Promise<Types.BillingInvoice[]>;
}
export declare class OperationsAPI {
    private client;
    constructor(client: AnarvaClient);
    get(id: string, options?: RequestOptions): Promise<Types.ControlPlaneOperation>;
    wait(id: string, pollOptions?: {
        timeoutMs?: number;
        intervalMs?: number;
        signal?: AbortSignal;
    }): Promise<Types.ControlPlaneOperation>;
}
