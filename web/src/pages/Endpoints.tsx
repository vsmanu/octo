import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import type { Config } from "../types";

export function Endpoints() {
    const [config, setConfig] = useState<Config | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        fetch("/api/v1/config")
            .then((res) => res.json())
            .then((data) => setConfig(data));
    }, []);

    if (!config) return <div className="p-8">Loading...</div>;

    return (
        <div className="space-y-6">
            <h1 className="text-3xl font-bold tracking-tight">Endpoints</h1>

            <div className="rounded-xl border bg-card text-card-foreground shadow">
                <div className="p-0">
                    <table className="w-full caption-bottom text-sm">
                        <thead className="[&_tr]:border-b">
                            <tr className="border-b transition-colors hover:bg-muted/50">
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Name</th>
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">URL</th>
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Method</th>
                            </tr>
                        </thead>
                        <tbody className="[&_tr:last-child]:border-0">
                            {config.endpoints.map((ep) => (
                                <tr
                                    key={ep.id}
                                    className="border-b transition-colors hover:bg-muted/50 cursor-pointer"
                                    onClick={() => navigate(`/endpoints/${ep.id}`)}
                                >
                                    <td className="p-4 align-middle font-medium">{ep.name}</td>
                                    <td className="p-4 align-middle">{ep.url}</td>
                                    <td className="p-4 align-middle">{ep.method}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
