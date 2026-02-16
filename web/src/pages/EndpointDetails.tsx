import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, AlertTriangle, CheckCircle, Shield, Clock } from "lucide-react";
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import type { Config, Endpoint, Metric } from "../types";

export function EndpointDetails() {
    const { id } = useParams();
    const navigate = useNavigate();
    const [endpoint, setEndpoint] = useState<Endpoint | null>(null);
    const [metrics, setMetrics] = useState<Metric[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Fetch config to get endpoint details
        fetch("/api/v1/config")
            .then((res) => res.json())
            .then((data: Config) => {
                const ep = data.endpoints.find((e) => e.id === id);
                if (ep) {
                    setEndpoint(ep);
                } else {
                    // Handle not found?
                }
            });

        // Fetch history
        if (id) {
            fetch(`/api/v1/endpoints/${id}/history?duration=1h`)
                .then((res) => res.json())
                .then((data: Metric[]) => {
                    // Transform data for chart if needed. 
                    // Recharts needs array of objects.
                    // Timestamps are ISO strings, might need formatting.
                    setMetrics(data || []);
                    setLoading(false);
                })
                .catch(err => {
                    console.error(err);
                    setLoading(false);
                });
        }
    }, [id]);

    if (loading) return <div className="p-8">Loading details...</div>;
    if (!endpoint) return <div className="p-8">Endpoint not found</div>;

    // Process data for charts
    const chartData = metrics.map(m => ({
        time: new Date(m.timestamp).toLocaleTimeString(),
        duration: m.duration_ns / 1_000_000, // ms
        success: m.success ? 1 : 0
    }));

    const lastMetric = metrics.length > 0 ? metrics[metrics.length - 1] : null;
    const isHealthy = lastMetric?.success;

    return (
        <div className="space-y-6">
            <div className="flex items-center space-x-4">
                <button onClick={() => navigate(-1)} className="p-2 hover:bg-muted rounded-full">
                    <ArrowLeft className="h-5 w-5" />
                </button>
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">{endpoint.name}</h1>
                    <p className="text-sm text-muted-foreground">
                        <a href={endpoint.url} target="_blank" rel="noopener noreferrer" className="hover:underline hover:text-primary">
                            {endpoint.url}
                        </a>
                    </p>
                </div>
                <div className="ml-auto">
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
            </div>

            <div className="grid gap-4 md:grid-cols-3">
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Last Response Time</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {lastMetric ? (lastMetric.duration_ns / 1_000_000).toFixed(2) : '-'}
                        </span>
                        <span className="ml-1 text-sm text-muted-foreground">ms</span>
                    </div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Check Interval</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {endpoint.interval / 1_000_000_000}
                        </span>
                        <span className="ml-1 text-sm text-muted-foreground">s</span>
                    </div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <h3 className="text-sm font-medium text-muted-foreground">Method</h3>
                    <div className="mt-2 flex items-baseline">
                        <span className="text-2xl font-semibold">
                            {endpoint.method}
                        </span>
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

            <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                <div className="mb-4">
                    <h3 className="font-semibold">Response Time History (1h)</h3>
                </div>
                <div className="h-[300px] w-full">
                    <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={chartData}>
                            <CartesianGrid strokeDasharray="3 3" vertical={false} />
                            <XAxis dataKey="time" hide />
                            <YAxis />
                            <Tooltip
                                contentStyle={{ backgroundColor: 'var(--color-card)', borderColor: 'var(--color-border)' }}
                                itemStyle={{ color: 'var(--color-foreground)' }}
                            />
                            <Line type="monotone" dataKey="duration" stroke="var(--color-primary)" strokeWidth={2} dot={false} />
                        </LineChart>
                    </ResponsiveContainer>
                </div>
            </div>
        </div>
    );
}
