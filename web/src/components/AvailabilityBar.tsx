
import type { Metric } from "../types";

interface AvailabilityBarProps {
    metrics: Metric[];
}

export function AvailabilityBar({ metrics }: AvailabilityBarProps) {
    // Determine the number of bars to show (e.g., last 50 checks)
    const limit = 60;
    const recentMetrics = metrics.slice(-limit);

    // Fill with placeholders if less than limit? Or just show what we have.
    // Let's show what we have, aligned to the right.

    return (
        <div className="w-full">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">Availability History</h3>
            <div className="flex w-full h-8 gap-0.5">
                {recentMetrics.map((m, i) => (
                    <div
                        key={i}
                        className={`flex-1 rounded-sm cursor-help transition-colors ${m.success ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'
                            }`}
                        title={`${new Date(m.timestamp).toLocaleString()} - ${m.success ? 'OK' : 'Error: ' + (m.error || 'Unknown')}`}
                        style={{ minWidth: '4px' }}
                    />
                ))}
                {/* Fill empty space if needed? */}
                {Array.from({ length: Math.max(0, limit - recentMetrics.length) }).map((_, i) => (
                    <div key={`empty-${i}`} className="flex-1 bg-muted/20 rounded-sm" />
                ))}
            </div>
            <div className="flex justify-between text-xs text-muted-foreground mt-1">
                <span>{limit} checks ago</span>
                <span>Now</span>
            </div>
        </div>
    );
}
