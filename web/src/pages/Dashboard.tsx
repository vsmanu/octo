import { useEffect, useState } from "react";
import { Activity, Server, AlertTriangle, CheckCircle } from "lucide-react";
import { Link } from "react-router-dom";
import type { Config } from "../types";

export function Dashboard() {
    const [config, setConfig] = useState<Config | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetch("/api/v1/config")
            .then((res) => res.json())
            .then((data) => {
                setConfig(data);
                setLoading(false);
            })
            .catch((err) => {
                console.error("Failed to fetch config:", err);
                setLoading(false);
            });
    }, []);

    if (loading) {
        return <div className="p-4">Loading...</div>;
    }

    const endpoints = config?.endpoints || [];
    const totalEndpoints = endpoints.length;

    return (
        <div className="space-y-6">
            <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>

            {/* Stats Overview */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <div className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <h3 className="tracking-tight text-sm font-medium">Total Endpoints</h3>
                        <Server className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div className="text-2xl font-bold">{totalEndpoints}</div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <div className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <h3 className="tracking-tight text-sm font-medium">Healthy</h3>
                        <CheckCircle className="h-4 w-4 text-green-500" />
                    </div>
                    <div className="text-2xl font-bold">-</div>
                    <p className="text-xs text-muted-foreground">Real-time status coming soon</p>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <div className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <h3 className="tracking-tight text-sm font-medium">Active Alerts</h3>
                        <AlertTriangle className="h-4 w-4 text-yellow-500" />
                    </div>
                    <div className="text-2xl font-bold">0</div>
                </div>
                <div className="rounded-xl border bg-card text-card-foreground shadow p-6">
                    <div className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <h3 className="tracking-tight text-sm font-medium">Uptime (24h)</h3>
                        <Activity className="h-4 w-4 text-blue-500" />
                    </div>
                    <div className="text-2xl font-bold">100%</div>
                </div>
            </div>

            {/* Recent Activity / Endpoints List */}
            <div className="rounded-xl border bg-card text-card-foreground shadow">
                <div className="p-6 flex flex-col space-y-1.5">
                    <h3 className="font-semibold leading-none tracking-tight">Monitored Endpoints</h3>
                    <p className="text-sm text-muted-foreground">Overview of all configured endpoints.</p>
                </div>
                <div className="p-6 pt-0">
                    <div className="relative w-full overflow-auto">
                        <table className="w-full caption-bottom text-sm">
                            <thead className="[&_tr]:border-b">
                                <tr className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted">
                                    <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Name</th>
                                    <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">URL</th>
                                    <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Method</th>
                                    <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Interval</th>
                                    <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Status</th>
                                </tr>
                            </thead>
                            <tbody className="[&_tr:last-child]:border-0">
                                {endpoints.map((ep) => (
                                    <tr key={ep.id} className="border-b transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted">
                                        <td className="p-4 align-middle font-medium">
                                            <Link to={`/endpoints/${ep.id}`} className="hover:underline hover:text-primary">
                                                {ep.name}
                                            </Link>
                                        </td>
                                        <td className="p-4 align-middle">
                                            <a href={ep.url} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">
                                                {ep.url}
                                            </a>
                                        </td>
                                        <td className="p-4 align-middle"><span className="inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80">{ep.method}</span></td>
                                        <td className="p-4 align-middle">{ep.interval / 1000000000}s</td>
                                        <td className="p-4 align-middle"><span className="flex h-2 w-2 rounded-full bg-green-500" /></td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>
        </div>
    );
}
