import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { AlertTriangle, ArrowLeft, CheckCircle, Clock, Settings, Shield } from "lucide-react";
import { Bar, BarChart, CartesianGrid, Cell, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import type { Config, Endpoint, Metric } from "../types";

export function EndpointDetails() {
    const { id } = useParams();
    const navigate = useNavigate();
    const [endpoint, setEndpoint] = useState<Endpoint | null>(null);
    const [metrics, setMetrics] = useState<Metric[]>([]);
    const [loading, setLoading] = useState(true);

    const [timeRange, setTimeRange] = useState("1h");
    const [customStart, setCustomStart] = useState("");
    const [customEnd, setCustomEnd] = useState("");
    const [isCustom, setIsCustom] = useState(false);

    const ranges = [
        { label: 'Last 1 Hour', value: '1h' },
        { label: 'Last 6 Hours', value: '6h' },
        { label: 'Last 12 Hours', value: '12h' },
        { label: 'Last 24 Hours', value: '24h' },
        { label: 'Last 7 Days', value: '168h' },
        { label: 'Last 30 Days', value: '720h' },
        { label: 'Last Quarter', value: '2160h' },
        { label: 'Last Year', value: '8760h' },
        { label: 'Custom', value: 'custom' },
    ];

    useEffect(() => {
        // Fetch config to get endpoint details
        fetch("/api/v1/config")
            .then((res) => res.json())
            .then((data: Config) => {
                const ep = data.endpoints.find((e) => e.id === id);
                if (ep) {
                    setEndpoint(ep);
                }
            });
    }, [id]);

    useEffect(() => {
        // Fetch history
        if (!id) return;

        let url = `/api/v1/endpoints/${id}/history`;
        if (isCustom) {
            if (!customStart || !customEnd) return; // Wait for both inputs
            const from = new Date(customStart).toISOString();
            const to = new Date(customEnd).toISOString();
            url += `?from=${from}&to=${to}`;
        } else {
            url += `?duration=${timeRange}`;
        }

        setLoading(true);
        fetch(url)
            .then((res) => res.json())
            .then((data: Metric[]) => {
                setMetrics(data || []);
                setLoading(false);
            })
            .catch(err => {
                console.error(err);
                setLoading(false);
            });
    }, [id, timeRange, isCustom, customStart, customEnd]); // Trigger on any change

    if (loading) return <div className="p-8">Loading details...</div>;
    if (!endpoint) return <div className="p-8">Endpoint not found</div>;

    // Process data for charts
    const chartData = metrics.map((m: Metric) => ({
        timestamp: new Date(m.timestamp).toISOString(),
        time: new Date(m.timestamp).toLocaleTimeString(),
        duration: m.duration_ns / 1_000_000, // ms
        success: m.success ? 1 : 0, // For bar chart
        status: m.success ? 'success' : 'failure',
        error: m.error || 'Unknown error',
        color: m.success ? '#22c55e' : '#ef4444' // green-500 : red-500
    }));

    const lastMetric = metrics.length > 0 ? metrics[metrics.length - 1] : null;
    const isHealthy = lastMetric?.success;

    const totalRequests = metrics.length;
    const successfulRequests = metrics.filter(m => m.success).length;
    const availability = totalRequests > 0 ? (successfulRequests / totalRequests) * 100 : 0;

    const durations = metrics.map(m => m.duration_ns / 1_000_000);
    const avgDuration = durations.length > 0 ? durations.reduce((a, b) => a + b, 0) / durations.length : 0;
    const minDuration = durations.length > 0 ? Math.min(...durations) : 0;
    const maxDuration = durations.length > 0 ? Math.max(...durations) : 0;

    return (
        <div className="space-y-6">
            <div className="flex flex-col space-y-4 md:flex-row md:items-center md:justify-between md:space-y-0">
                <div className="flex items-center space-x-4">
                    <button onClick={() => navigate(-1)} className="p-2 hover:bg-muted rounded-full">
                        <ArrowLeft className="h-5 w-5" />
                    </button>
                    <div>
                        <div className="flex items-center gap-3">
                            <h1 className="text-2xl font-bold tracking-tight">{endpoint.name}</h1>
                            <button
                                onClick={() => navigate(`/endpoints/${id}/edit`)}
                                className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors hover:bg-muted h-8 w-8 text-muted-foreground"
                                title="Edit Configuration"
                            >
                                <Settings className="h-4 w-4" />
                            </button>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                            <a href={endpoint.url} target="_blank" rel="noopener noreferrer" className="hover:underline hover:text-primary">
                                {endpoint.url}
                            </a>
                        </p>
                    </div>
                    {isHealthy ? (
                        <span className="inline-flex items-center rounded-full bg-green-100 px-3 py-1 text-sm font-medium text-green-800">
                            <CheckCircle className="mr-1.5 h-4 w-4" /> Healthy
                        </span>
                    ) : (metrics.length > 0 ? (
                        <span className="inline-flex items-center rounded-full bg-red-100 px-3 py-1 text-sm font-medium text-red-800">
                            <AlertTriangle className="mr-1.5 h-4 w-4" /> Unhealthy
                        </span>
                    ) : (
                        <span className="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-sm font-medium text-gray-800">
                            Unknown
                        </span>
                    )
                    )}
                </div>

                <div className="flex flex-col items-end gap-2">
                    <div className="flex items-center gap-2">
                        <div className="flex items-center gap-2 bg-card p-1 rounded-md border shadow-sm">
                            <Clock className="h-4 w-4 text-muted-foreground ml-2" />
                            <select
                                value={isCustom ? 'custom' : timeRange}
                                onChange={(e) => {
                                    const val = e.target.value;
                                    if (val === 'custom') {
                                        setIsCustom(true);
                                    } else {
                                        setIsCustom(false);
                                        setTimeRange(val);
                                    }
                                }}
                                className="h-8 rounded-md bg-transparent px-2 text-sm focus-visible:outline-none"
                            >
                                {ranges.map((r) => (
                                    <option key={r.value} value={r.value}>{r.label}</option>
                                ))}
                            </select>
                        </div>
                    </div>
                    {isCustom && (
                        <div className="flex flex-wrap items-center gap-2 text-sm bg-card p-2 rounded-md border shadow-sm">
                            <span className="text-muted-foreground">From:</span>
                            <input
                                type="datetime-local"
                                value={customStart}
                                onChange={(e) => setCustomStart(e.target.value)}
                                className="h-8 rounded-md border border-input bg-transparent px-2"
                            />
                            <span className="text-muted-foreground">To:</span>
                            <input
                                type="datetime-local"
                                value={customEnd}
                                onChange={(e) => setCustomEnd(e.target.value)}
                                className="h-8 rounded-md border border-input bg-transparent px-2"
                            />
                        </div>
                    )}
                </div>
            </div>

            <div className="grid gap-4 md:grid-cols-4">
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Availability</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {availability.toFixed(2)}%
                        </span>
                    </div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Avg Latency</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {avgDuration.toFixed(2)}
                        </span>
                        <span className="ml-1 text-sm text-muted-foreground">ms</span>
                    </div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Min Latency</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {minDuration.toFixed(2)}
                        </span>
                        <span className="ml-1 text-sm text-muted-foreground">ms</span>
                    </div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Max Latency</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {maxDuration.toFixed(2)}
                        </span>
                        <span className="ml-1 text-sm text-muted-foreground">ms</span>
                    </div>
                </div>
            </div>

            {(lastMetric?.cert_expiry) && (
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <div className="flex items-center gap-2 mb-4">
                        <Shield className="h-5 w-5 text-primary" />
                        <h3 className="font-semibold">SSL/TLS Certificate</h3>
                    </div>
                    <div className="grid gap-6 md:grid-cols-3">
                        <div>
                            <p className="text-sm font-medium text-muted-foreground mb-1">Expiration</p>
                            <div className="flex items-center gap-2">
                                <Clock className="h-4 w-4 text-muted-foreground" />
                                <span className="font-medium">
                                    {new Date(lastMetric.cert_expiry!).toLocaleDateString()}
                                </span>
                            </div>
                            <p className="text-xs text-muted-foreground mt-1">
                                Expires in {Math.ceil((new Date(lastMetric.cert_expiry!).getTime() - Date.now()) / (1000 * 60 * 60 * 24))} days
                            </p>
                        </div>
                        <div className="md:col-span-2">
                            <div className="grid gap-4 md:grid-cols-2">
                                <div>
                                    <p className="text-sm font-medium text-muted-foreground mb-1">Issuer</p>
                                    <p className="text-sm font-medium truncate" title={lastMetric.cert_issuer}>
                                        {lastMetric.cert_issuer}
                                    </p>
                                </div>
                                <div>
                                    <p className="text-sm font-medium text-muted-foreground mb-1">Subject</p>
                                    <p className="text-sm font-medium truncate" title={lastMetric.cert_subject}>
                                        {lastMetric.cert_subject}
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            <div className="space-y-4">
                {/* Availability Chart */}
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="font-semibold mb-4">Availability History</h3>
                    <div className="h-[60px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={chartData} syncId="endpointMetrics">
                                <Tooltip
                                    trigger="hover"
                                    content={({ active, payload }) => {
                                        if (active && payload && payload.length) {
                                            const data = payload[0].payload;
                                            return (
                                                <div className="rounded-lg border bg-background p-2 shadow-sm text-xs">
                                                    <div className="font-bold">{data.time}</div>
                                                    <div className={data.success ? "text-green-500" : "text-red-500"}>
                                                        {data.success ? "Success" : `Error: ${data.error}`}
                                                    </div>
                                                </div>
                                            );
                                        }
                                        return null;
                                    }}
                                />
                                <Bar dataKey="success" maxBarSize={10} minPointSize={2}>
                                    {chartData.map((entry: any, index: number) => (
                                        <Cell key={`cell-${index}`} fill={entry.color} />
                                    ))}
                                </Bar>
                            </BarChart>
                        </ResponsiveContainer>
                    </div>
                </div>

                {/* Response Time Chart */}
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="font-semibold mb-4">Response Time History (1h)</h3>
                    <div className="h-[300px] w-full">
                        <ResponsiveContainer width="100%" height="100%">
                            <LineChart data={chartData} syncId="endpointMetrics">
                                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                <XAxis dataKey="time" hide />
                                <YAxis
                                    label={{ value: 'ms', angle: -90, position: 'insideLeft' }}
                                    tickFormatter={(val) => val.toFixed(0)}
                                />
                                <Tooltip
                                    contentStyle={{ backgroundColor: 'var(--color-card)', borderColor: 'var(--color-border)' }}
                                    itemStyle={{ color: 'var(--color-foreground)' }}
                                    labelStyle={{ color: 'var(--color-foreground)' }}
                                    formatter={(value: number | undefined) => [value ? `${value.toFixed(2)} ms` : '0 ms', 'Duration']}
                                />
                                <Line
                                    type="monotone"
                                    dataKey="duration"
                                    stroke="var(--color-primary)"
                                    strokeWidth={2}
                                    dot={false}
                                    activeDot={{ r: 4 }}
                                />
                            </LineChart>
                        </ResponsiveContainer>
                    </div>
                </div>
            </div>
        </div>
    );
}
