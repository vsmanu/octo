import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { ArrowLeft, Save, Trash, Plus, X } from "lucide-react";
import type { Config, Endpoint } from "../types";

// Helper component for Key-Value pairs (Headers, Tags)
function KeyValueEditor({
    items,
    onChange,
    title
}: {
    items: Record<string, string>;
    onChange: (items: Record<string, string>) => void;
    title: string;
}) {
    const entries = Object.entries(items);

    const updateItem = (oldKey: string, newKey: string, newValue: string) => {
        const newItems = { ...items };
        if (oldKey !== newKey) {
            delete newItems[oldKey];
        }
        newItems[newKey] = newValue;
        onChange(newItems);
    };

    const deleteItem = (keyToDelete: string) => {
        const newItems = { ...items };
        delete newItems[keyToDelete];
        onChange(newItems);
    };

    const addItem = () => {
        const newItems = { ...items, "": "" };
        onChange(newItems);
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                    {title}
                </label>
                <button
                    type="button"
                    onClick={addItem}
                    className="inline-flex items-center text-xs text-primary hover:underline"
                >
                    <Plus className="mr-1 h-3 w-3" /> Add {title.slice(0, -1)}
                </button>
            </div>
            <div className="space-y-2">
                {entries.map(([key, value], index) => (
                    <div key={index} className="flex items-center gap-2">
                        <input
                            placeholder="Key"
                            value={key}
                            onChange={(e) => updateItem(key, e.target.value, value)}
                            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        />
                        <input
                            placeholder="Value"
                            value={value}
                            onChange={(e) => updateItem(key, key, e.target.value)}
                            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        />
                        <button
                            type="button"
                            onClick={() => deleteItem(key)}
                            className="p-2 text-muted-foreground hover:text-red-600"
                        >
                            <X className="h-4 w-4" />
                        </button>
                    </div>
                ))}
                {entries.length === 0 && (
                    <div className="text-sm text-muted-foreground italic">No {title.toLowerCase()} defined.</div>
                )}
            </div>
        </div>
    );
}

export function EndpointForm() {
    const { id } = useParams();
    const navigate = useNavigate();
    const isEditMode = Boolean(id);

    const [formData, setFormData] = useState<Partial<Endpoint>>({
        name: "",
        url: "",
        method: "GET",
        interval: 60,
        timeout: 10,
        headers: {},
        validation: {
            status_codes: [200],
            content_match: {
                type: "",
                pattern: ""
            }
        },
        ssl: {
            expiration_alert_days: [30, 7, 1]
        },
        tags: {}
    });
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    useEffect(() => {
        if (isEditMode && id) {
            setLoading(true);
            fetch("/api/v1/config")
                .then((res) => res.json())
                .then((data: Config) => {
                    const ep = data.endpoints.find((e) => e.id === id);
                    if (ep) {
                        // Convert nanoseconds to seconds for display
                        setFormData({
                            ...ep,
                            interval: ep.interval / 1_000_000_000,
                            timeout: ep.timeout / 1_000_000_000,
                            validation: {
                                ...ep.validation,
                                content_match: ep.validation.content_match || { type: "", pattern: "" }
                            }
                        });
                    } else {
                        setError("Endpoint not found");
                    }
                    setLoading(false);
                })
                .catch((err) => {
                    setError("Failed to load endpoint");
                    console.error(err);
                    setLoading(false);
                });
        }
    }, [id, isEditMode]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value } = e.target;
        setFormData((prev) => ({
            ...prev,
            [name]: (name === "interval" || name === "timeout") ? Number(value) : value,
        }));
    };

    // Deep merge helper for nested state updates would be nice, but simple handlers work too
    const handleValidationChange = (field: string, value: any) => {
        setFormData(prev => ({
            ...prev,
            validation: {
                ...prev.validation!,
                [field]: value
            }
        }));
    };

    const handleSSLChange = (days: string) => {
        const daysArray = days.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n));
        setFormData(prev => ({
            ...prev,
            ssl: {
                expiration_alert_days: daysArray
            }
        }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setError("");

        try {
            const url = isEditMode
                ? `/api/v1/config/endpoints/${id}`
                : "/api/v1/config/endpoints";
            const method = isEditMode ? "PUT" : "POST";

            // Prepare payload
            const payload = {
                ...formData,
                interval: (formData.interval || 60) * 1_000_000_000,
                timeout: (formData.timeout || 10) * 1_000_000_000,
                // Clean up empty optional fields
                validation: {
                    ...formData.validation,
                    content_match: (formData.validation?.content_match?.pattern)
                        ? formData.validation.content_match
                        : undefined
                }
            };

            const res = await fetch(url, {
                method,
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(payload),
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || "Failed to save endpoint");
            }

            navigate("/endpoints");
        } catch (err: any) {
            setError(err.message);
            setLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!confirm("Are you sure you want to delete this endpoint?")) return;

        setLoading(true);
        try {
            const res = await fetch(`/api/v1/config/endpoints/${id}`, {
                method: "DELETE",
            });

            if (!res.ok) {
                const errText = await res.text();
                throw new Error(errText || "Failed to delete endpoint");
            }

            navigate("/endpoints");
        } catch (err: any) {
            setError(err.message);
            setLoading(false);
        }
    };

    if (loading && isEditMode && !formData.id) return <div className="p-8">Loading...</div>;

    return (
        <div className="max-w-3xl mx-auto space-y-6 pb-12">
            <div className="flex items-center space-x-4">
                <button onClick={() => navigate(-1)} className="p-2 hover:bg-muted rounded-full">
                    <ArrowLeft className="h-5 w-5" />
                </button>
                <h1 className="text-2xl font-bold tracking-tight">
                    {isEditMode ? "Edit Endpoint" : "Add Endpoint"}
                </h1>
                {isEditMode && (
                    <button
                        onClick={handleDelete}
                        className="ml-auto inline-flex items-center rounded-md bg-red-100 px-3 py-2 text-sm font-medium text-red-800 hover:bg-red-200"
                    >
                        <Trash className="mr-2 h-4 w-4" /> Delete
                    </button>
                )}
            </div>

            {error && (
                <div className="rounded-md bg-red-50 p-4 text-sm text-red-700">
                    {error}
                </div>
            )}

            <form onSubmit={handleSubmit} className="space-y-8">
                {/* Basic Info */}
                <div className="space-y-4 bg-card p-6 rounded-xl border shadow">
                    <h2 className="text-lg font-semibold">Basic Information</h2>
                    <div className="space-y-2">
                        <label className="text-sm font-medium leading-none">Name</label>
                        <input
                            name="name"
                            value={formData.name}
                            onChange={handleChange}
                            required
                            className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            placeholder="My Service"
                        />
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium leading-none">URL</label>
                        <input
                            name="url"
                            value={formData.url}
                            onChange={handleChange}
                            required
                            type="url"
                            className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            placeholder="https://example.com"
                        />
                    </div>
                </div>

                {/* Request Settings */}
                <div className="space-y-4 bg-card p-6 rounded-xl border shadow">
                    <h2 className="text-lg font-semibold">Request Settings</h2>
                    <div className="grid grid-cols-3 gap-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium leading-none">Method</label>
                            <select
                                name="method"
                                value={formData.method}
                                onChange={handleChange}
                                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            >
                                <option value="GET">GET</option>
                                <option value="POST">POST</option>
                                <option value="PUT">PUT</option>
                                <option value="HEAD">HEAD</option>
                                <option value="DELETE">DELETE</option>
                                <option value="PATCH">PATCH</option>
                            </select>
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium leading-none">Interval (sec)</label>
                            <input
                                name="interval"
                                type="number"
                                min="1"
                                value={formData.interval}
                                onChange={handleChange}
                                required
                                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium leading-none">Timeout (sec)</label>
                            <input
                                name="timeout"
                                type="number"
                                min="1"
                                value={formData.timeout}
                                onChange={handleChange}
                                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            />
                        </div>
                    </div>

                    <div className="pt-2">
                        <KeyValueEditor
                            title="Request Headers"
                            items={formData.headers || {}}
                            onChange={(headers) => setFormData(prev => ({ ...prev, headers }))}
                        />
                    </div>
                </div>

                {/* Validation */}
                <div className="space-y-4 bg-card p-6 rounded-xl border shadow">
                    <h2 className="text-lg font-semibold">Validation</h2>
                    <div className="space-y-2">
                        <label className="text-sm font-medium leading-none">Expected Status Codes (comma separated)</label>
                        <input
                            value={formData.validation?.status_codes?.join(", ") || ""}
                            onChange={(e) => handleValidationChange("status_codes", e.target.value.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n)))}
                            className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            placeholder="200, 201"
                        />
                    </div>

                    <div className="grid grid-cols-3 gap-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium leading-none">Content Match Type</label>
                            <select
                                value={formData.validation?.content_match?.type || ""}
                                onChange={(e) => handleValidationChange("content_match", { ...formData.validation?.content_match, type: e.target.value })}
                                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            >
                                <option value="">None</option>
                                <option value="exact">Exact Match</option>
                                <option value="regex">Regular Expression</option>
                            </select>
                        </div>
                        <div className="col-span-2 space-y-2">
                            <label className="text-sm font-medium leading-none">Content Match Pattern</label>
                            <input
                                value={formData.validation?.content_match?.pattern || ""}
                                onChange={(e) => handleValidationChange("content_match", { ...formData.validation?.content_match, pattern: e.target.value })}
                                disabled={!formData.validation?.content_match?.type}
                                className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50"
                                placeholder={formData.validation?.content_match?.type === 'regex' ? 'Regex pattern' : 'Exact string'}
                            />
                        </div>
                    </div>
                </div>

                {/* Advanced & Tags */}
                <div className="space-y-4 bg-card p-6 rounded-xl border shadow">
                    <h2 className="text-lg font-semibold">Advanced & Metadata</h2>

                    <div className="space-y-2">
                        <label className="text-sm font-medium leading-none">SSL Expiration Alerts (days before expiry, comma separated)</label>
                        <input
                            value={formData.ssl?.expiration_alert_days?.join(", ") || ""}
                            onChange={(e) => handleSSLChange(e.target.value)}
                            className="flex h-10 w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                            placeholder="30, 7, 1"
                        />
                    </div>

                    <div className="pt-2">
                        <KeyValueEditor
                            title="Tags"
                            items={formData.tags || {}}
                            onChange={(tags) => setFormData(prev => ({ ...prev, tags }))}
                        />
                    </div>
                </div>

                <div className="flex justify-end pt-4">
                    <button
                        type="submit"
                        disabled={loading}
                        className="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-11 px-8 py-2"
                    >
                        {loading ? "Saving..." : <><Save className="mr-2 h-4 w-4" /> Save Configuration</>}
                    </button>
                </div>
            </form>
        </div>
    );
}
