export interface AnarvaErrorOptions {
  code: string;
  message: string;
  requestId?: string;
  status?: number;
  details?: Record<string, unknown>;
}

export class AnarvaError extends Error {
  public readonly code: string;
  public readonly requestId?: string;
  public readonly status?: number;
  public readonly details?: Record<string, unknown>;

  constructor(options: AnarvaErrorOptions) {
    // Redact any potential API key secret if accidentally included in message
    const sanitizedMsg = options.message.replace(/anarva_(live|test)_[a-zA-Z0-9]+/g, '[REDACTED_API_KEY]');
    super(sanitizedMsg);

    this.name = 'AnarvaError';
    this.code = options.code;
    this.requestId = options.requestId;
    this.status = options.status;
    this.details = options.details;

    // Restore proper prototype chain
    Object.setPrototypeOf(this, AnarvaError.prototype);
  }

  public toJSON() {
    return {
      name: this.name,
      code: this.code,
      message: this.message,
      requestId: this.requestId,
      status: this.status,
      details: this.details,
    };
  }
}
