export interface AnarvaClientConfig {
    apiKey?: string;
    apiUrl?: string;
    timeout?: number;
    maxRetries?: number;
    debug?: boolean;
}
export interface ResolvedClientConfig {
    apiKey: string;
    apiUrl: string;
    timeout: number;
    maxRetries: number;
    debug: boolean;
}
export declare function resolveConfig(config?: AnarvaClientConfig): ResolvedClientConfig;
