export interface Endpoint {
    id: string;
    name: string;
    url: string;
    method: string;
    interval: number;
    timeout: number;
    headers: Record<string, string>;
    validation: {
        status_codes: number[];
        content_match?: {
            type: string;
            pattern: string;
        };
    };
    ssl: {
        expiration_alert_days: number[];
    };
    tags: Record<string, string>;
}

export interface GlobalConfig {
    check_interval: number;
    request_timeout: number;
}

export interface Config {
    global: GlobalConfig;
    endpoints: Endpoint[];
}

export interface Metric {
    endpoint_id: string;
    timestamp: string;
    duration_ns: number;
    status_code: number;
    success: boolean;
    error?: string;
    cert_expiry?: string;
    cert_issuer?: string;
    cert_subject?: string;
}

export interface User {
    username: string;
    role: string;
}
