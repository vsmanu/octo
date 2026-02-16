import { Sidebar } from "./Sidebar";
import { Outlet } from "react-router-dom";

export function Layout() {
    return (
        <div className="flex h-screen bg-background">
            <Sidebar />
            <main className="flex-1 overflow-y-auto p-8">
                <Outlet />
            </main>
        </div>
    );
}
