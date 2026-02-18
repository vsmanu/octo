import { useAuth } from "@/context/AuthContext";
import { User, Shield, Key } from "lucide-react";

export function Profile() {
    const { user } = useAuth();

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Profile</h1>
                <p className="text-muted-foreground">Manage your account settings.</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* User Info Card */}
                <div className="rounded-lg border bg-card text-card-foreground shadow-sm">
                    <div className="flex flex-col space-y-1.5 p-6">
                        <h3 className="font-semibold leading-none tracking-tight">Personal Information</h3>
                        <p className="text-sm text-muted-foreground">Your account details.</p>
                    </div>
                    <div className="p-6 pt-0 space-y-4">
                        <div className="flex items-center gap-4 rounded-md border p-4">
                            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                                <User className="h-6 w-6" />
                            </div>
                            <div>
                                <p className="text-sm font-medium leading-none">Username</p>
                                <p className="text-lg font-bold">{user?.username}</p>
                            </div>
                        </div>

                        <div className="flex items-center gap-4 rounded-md border p-4">
                            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 text-primary">
                                <Shield className="h-6 w-6" />
                            </div>
                            <div>
                                <p className="text-sm font-medium leading-none">Role</p>
                                <p className="text-lg font-bold capitalize">{user?.role}</p>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Security Card */}
                <div className="rounded-lg border bg-card text-card-foreground shadow-sm">
                    <div className="flex flex-col space-y-1.5 p-6">
                        <h3 className="font-semibold leading-none tracking-tight">Security</h3>
                        <p className="text-sm text-muted-foreground">Manage your password and security settings.</p>
                    </div>
                    <div className="p-6 pt-0 space-y-4">
                        <div className="rounded-md bg-muted p-4">
                            <div className="flex items-center gap-2 mb-2">
                                <Key className="h-5 w-5 text-muted-foreground" />
                                <span className="font-medium">Password</span>
                            </div>
                            <p className="text-sm text-muted-foreground mb-4">
                                Changing password is not yet supported in this version. Please contact your administrator.
                            </p>
                            <button className="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 cursor-not-allowed opacity-50" disabled>
                                Change Password
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
