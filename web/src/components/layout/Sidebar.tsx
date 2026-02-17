import { LayoutDashboard, Settings, Activity, ChevronLeft, ChevronRight } from "lucide-react";
import { Link, useLocation } from "react-router-dom";
import { cn } from "@/lib/utils";
import { useState, useEffect } from "react";

const navigation = [
    { name: "Dashboard", href: "/", icon: LayoutDashboard },
    { name: "Endpoints", href: "/endpoints", icon: Activity },
    { name: "Configuration", href: "/config", icon: Settings },
];

export function Sidebar() {
    const location = useLocation();
    const [isCollapsed, setIsCollapsed] = useState(false);

    // Auto-collapse on mobile/narrow screens
    useEffect(() => {
        const handleResize = () => {
            if (window.innerWidth < 1024) {
                setIsCollapsed(true);
            } else {
                setIsCollapsed(false);
            }
        };

        // Initial check
        handleResize();

        window.addEventListener("resize", handleResize);
        return () => window.removeEventListener("resize", handleResize);
    }, []);

    return (
        <div
            className={cn(
                "flex h-full flex-col bg-card border-r transition-all duration-300 ease-in-out",
                isCollapsed ? "w-20" : "w-64"
            )}
        >
            <div className={cn("flex h-16 items-center px-6", isCollapsed ? "justify-center px-0" : "justify-between")}>
                <div className="flex items-center overflow-hidden">
                    <Activity className="h-8 w-8 text-primary flex-shrink-0" />
                    <span className={cn("ml-2 text-xl font-bold transition-all duration-300", isCollapsed ? "opacity-0 w-0" : "opacity-100")}>
                        Octo
                    </span>
                </div>
                {!isCollapsed && (
                    <button
                        onClick={() => setIsCollapsed(!isCollapsed)}
                        className="p-1 hover:bg-accent rounded-md lg:hidden"
                    >
                        <ChevronLeft className="h-4 w-4" />
                    </button>
                )}
            </div>

            <nav className="flex-1 space-y-1 px-4 py-4">
                {navigation.map((item) => {
                    const isActive = location.pathname === item.href;
                    return (
                        <Link
                            key={item.name}
                            to={item.href}
                            title={isCollapsed ? item.name : undefined}
                            className={cn(
                                "group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-all duration-300",
                                isActive
                                    ? "bg-primary/10 text-primary"
                                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
                                isCollapsed ? "justify-center" : ""
                            )}
                        >
                            <item.icon
                                className={cn(
                                    "h-5 w-5 flex-shrink-0",
                                    isActive ? "text-primary" : "text-muted-foreground group-hover:text-accent-foreground",
                                    !isCollapsed && "mr-3"
                                )}
                                aria-hidden="true"
                            />
                            <span className={cn("whitespace-nowrap transition-all duration-300 overflow-hidden", isCollapsed ? "w-0 opacity-0" : "w-auto opacity-100")}>
                                {item.name}
                            </span>
                        </Link>
                    );
                })}
            </nav>

            <div className={cn("p-4 border-t", isCollapsed ? "flex justify-center" : "flex justify-end")}>
                <button
                    onClick={() => setIsCollapsed(!isCollapsed)}
                    className="p-2 hover:bg-accent rounded-md text-muted-foreground"
                >
                    {isCollapsed ? <ChevronRight className="h-5 w-5" /> : <ChevronLeft className="h-5 w-5" />}
                </button>
            </div>
        </div>
    );
}
