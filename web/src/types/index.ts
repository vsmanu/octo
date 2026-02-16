export interface Endpoint {
    id: string;
    name: string;
    url: string;
    method: string;
    interval: number;
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
    cert_expiry?: string;
    cert_issuer?: string;
    cert_subject?: string;
}
