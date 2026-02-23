import { useState, useEffect } from 'react';
import { useAuth } from '../context/AuthContext';

export const Config = () => {
    const [config, setConfig] = useState('');
    const [status, setStatus] = useState<'idle' | 'loading' | 'saving' | 'success' | 'error'>('idle');
    const [message, setMessage] = useState('');
    const { user } = useAuth();
    const isAdmin = user?.role === "admin";

    useEffect(() => {
        fetchConfig();
    }, []);

    const fetchConfig = async () => {
        setStatus('loading');
        try {
            const response = await fetch('/api/v1/config');
            if (!response.ok) throw new Error('Failed to fetch config');
            const data = await response.json();
            setConfig(JSON.stringify(data, null, 2));
            setStatus('idle');
        } catch (err) {
            setStatus('error');
            setMessage(err instanceof Error ? err.message : 'An error occurred');
        }
    };

    const handleSave = async () => {
        setStatus('saving');
        setMessage('');
        try {
            // Validate JSON
            let parsed;
            try {
                parsed = JSON.parse(config);
            } catch (e) {
                throw new Error('Invalid JSON format');
            }

            const response = await fetch('/api/v1/config', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(parsed),
            });

            if (!response.ok) {
                const text = await response.text();
                throw new Error(text || 'Failed to update config');
            }

            setStatus('success');
            setMessage('Configuration saved successfully!');

            // Clear success message after 3 seconds
            setTimeout(() => {
                setStatus('idle');
                setMessage('');
            }, 3000);

        } catch (err) {
            setStatus('error');
            setMessage(err instanceof Error ? err.message : 'An error occurred');
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold text-gray-900 dark:text-white">Configuration</h1>
                {isAdmin && (
                    <button
                        onClick={handleSave}
                        disabled={status === 'saving' || status === 'loading'}
                        className={`px-4 py-2 rounded-md text-white font-medium transition-colors ${status === 'saving' || status === 'loading'
                            ? 'bg-blue-400 cursor-not-allowed'
                            : 'bg-blue-600 hover:bg-blue-700'
                            }`}
                    >
                        {status === 'saving' ? 'Saving...' : 'Save Changes'}
                    </button>
                )}
            </div>

            {message && (
                <div className={`p-4 rounded-md ${status === 'error' ? 'bg-red-100 text-red-700' : 'bg-green-100 text-green-700'
                    }`}>
                    {message}
                </div>
            )}

            <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
                <p className="mb-4 text-sm text-gray-600 dark:text-gray-400">
                    {isAdmin ? "Edit your configuration below. Ensure valid JSON syntax." : "Current configuration schema (Read-Only)."}
                </p>
                <textarea
                    value={config}
                    onChange={(e) => setConfig(e.target.value)}
                    disabled={!isAdmin}
                    className="w-full h-96 font-mono text-sm p-4 border border-gray-300 dark:border-gray-700 rounded-md bg-gray-50 dark:bg-gray-900 text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none resize-y disabled:opacity-75 disabled:cursor-not-allowed"
                    placeholder="{ ... }"
                />
            </div>
        </div>
    );
};
