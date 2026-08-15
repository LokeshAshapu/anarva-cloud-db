export interface AnarvaErrorOptions {
    code: string;
    message: string;
    requestId?: string;
    status?: number;
    details?: Record<string, unknown>;
}
export declare class AnarvaError extends Error {
    readonly code: string;
    readonly requestId?: string;
    readonly status?: number;
    readonly details?: Record<string, unknown>;
    constructor(options: AnarvaErrorOptions);
    toJSON(): {
        name: string;
        code: string;
        message: string;
        requestId: string | undefined;
        status: number | undefined;
        details: Record<string, unknown> | undefined;
    };
}
