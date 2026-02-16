import { LayoutDashboard, Settings, Activity } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { cn } from "@/lib/utils";

const navigation = [
    { name: "Dashboard", href: "/", icon: LayoutDashboard },
    { name: "Endpoints", href: "/endpoints", icon: Activity },
    { name: "Configuration", href: "/config", icon: Settings },
];

export function Sidebar() {
    const location = useLocation();

    return (
        <div className="flex h-full w-64 flex-col bg-card border-r">
            <div className="flex h-16 items-center px-6">
                <Activity className="h-8 w-8 text-primary" />
                <span className="ml-2 text-xl font-bold">Octo</span>
            </div>
            <nav className="flex-1 space-y-1 px-4 py-4">
                {navigation.map((item) => {
                    const isActive = location.pathname === item.href;
                    return (
                        <Link
                            key={item.name}
                            to={item.href}
                            className={cn(
                                "group flex items-center px-2 py-2 text-sm font-medium rounded-md",
                                isActive
                                    ? "bg-primary/10 text-primary"
                                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                            )}
                        >
                            <item.icon
                                className={cn(
                                    "mr-3 h-5 w-5 flex-shrink-0",
                                    isActive ? "text-primary" : "text-muted-foreground group-hover:text-accent-foreground"
                                )}
                                aria-hidden="true"
                            />
                            {item.name}
                        </Link>
                    );
                })}
            </nav>
        </div>
    );
}
