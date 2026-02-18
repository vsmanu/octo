import { useState, useRef, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { User, LogOut, ChevronDown } from "lucide-react";
import { useAuth } from "@/context/AuthContext";
import { cn } from "@/lib/utils";

export function Header() {
    const { user, logout } = useAuth();
    const navigate = useNavigate();
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);

    const handleLogout = () => {
        logout();
        navigate("/login");
    };

    // Close dropdown when clicking outside
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        }
        document.addEventListener("mousedown", handleClickOutside);
        return () => {
            document.removeEventListener("mousedown", handleClickOutside);
        };
    }, []);

    return (
        <header className="flex h-16 items-center justify-between border-b bg-card px-6">
            <div className="flex items-center gap-2">
                {/* Placeholder for page title or breadcrumbs if needed later */}
            </div>

            <div className="relative" ref={dropdownRef}>
                <button
                    onClick={() => setIsOpen(!isOpen)}
                    className="flex items-center gap-3 rounded-full border bg-background px-3 py-1.5 hover:bg-accent transition-colors focus:outline-none focus:ring-2 focus:ring-primary/50"
                >
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/10 text-primary">
                        <User className="h-4 w-4" />
                    </div>
                    <div className="hidden text-left text-sm md:block">
                        <p className="font-medium">{user?.username || "User"}</p>
                    </div>
                    <ChevronDown className={cn("h-4 w-4 text-muted-foreground transition-transform duration-200", isOpen && "rotate-180")} />
                </button>

                {isOpen && (
                    <div className="absolute right-0 mt-2 w-48 origin-top-right rounded-md border bg-popover p-1 shadow-md ring-1 ring-black ring-opacity-5 focus:outline-none animate-in fade-in zoom-in-95 duration-100">
                        <div className="px-2 py-1.5 text-sm font-semibold text-muted-foreground border-b mb-1">
                            My Account
                        </div>
                        <Link
                            to="/profile"
                            className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-accent hover:text-accent-foreground cursor-pointer"
                            onClick={() => setIsOpen(false)}
                        >
                            <User className="mr-2 h-4 w-4" />
                            <span>Profile</span>
                        </Link>
                        <button
                            onClick={handleLogout}
                            className="flex w-full items-center rounded-sm px-2 py-1.5 text-sm outline-none hover:bg-destructive hover:text-destructive-foreground cursor-pointer text-destructive"
                        >
                            <LogOut className="mr-2 h-4 w-4" />
                            <span>Log out</span>
                        </button>
                    </div>
                )}
            </div>
        </header>
    );
}
