export function resolveConfig(config) {
    const apiKey = config?.apiKey || process.env.ANARVA_API_KEY || '';
    const apiUrl = config?.apiUrl || process.env.ANARVA_API_URL || 'https://anarva-cloud-db-api.onrender.com';
    const timeout = config?.timeout ?? 30000;
    const maxRetries = config?.maxRetries ?? 3;
    const debug = config?.debug ?? false;
    return {
        apiKey,
        apiUrl: apiUrl.replace(/\/+$/, ''),
        timeout,
        maxRetries,
        debug,
    };
}
