import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Settings } from "lucide-react";
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
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold tracking-tight">Endpoints</h1>
                <button
                    onClick={() => navigate("/endpoints/new")}
                    className="inline-flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
                >
                    Add Endpoint
                </button>
            </div>

            <div className="rounded-xl border bg-card text-card-foreground shadow">
                <div className="p-0">
                    <table className="w-full caption-bottom text-sm">
                        <thead className="[&_tr]:border-b">
                            <tr className="border-b transition-colors hover:bg-muted/50">
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Name</th>
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">URL</th>
                                <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Method</th>
                                <th className="h-12 px-4 text-right align-middle font-medium text-muted-foreground">Actions</th>
                            </tr>
                        </thead>
                        <tbody className="[&_tr:last-child]:border-0">
                            {config.endpoints.map((ep) => (
                                <tr
                                    key={ep.id}
                                    className="border-b transition-colors hover:bg-muted/50"
                                >
                                    <td className="p-4 align-middle font-medium cursor-pointer" onClick={() => navigate(`/endpoints/${ep.id}`)}>
                                        {ep.name}
                                    </td>
                                    <td className="p-4 align-middle cursor-pointer" onClick={() => navigate(`/endpoints/${ep.id}`)}>
                                        {ep.url}
                                    </td>
                                    <td className="p-4 align-middle cursor-pointer" onClick={() => navigate(`/endpoints/${ep.id}`)}>
                                        {ep.method}
                                    </td>
                                    <td className="p-4 align-middle text-right">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                navigate(`/endpoints/${ep.id}/edit`);
                                            }}
                                            className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors hover:bg-muted h-9 w-9"
                                            title="Edit Configuration"
                                        >
                                            <Settings className="h-4 w-4" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    );
}
