import { NavLink, Outlet } from 'react-router-dom';
import { Users, Film, BarChart } from 'lucide-react';
import { cn } from '../../lib/utils';

export const AdminLayout = () => {
  return (
    <div className="flex flex-col md:flex-row gap-6 pt-4 pb-12 w-full flex-1">
      <aside className="w-full md:w-64 shrink-0">
        <nav className="flex flex-col gap-2">
          <div className="font-semibold text-lg mb-2 px-3">Admin Dashboard</div>
          <NavLink
            to="/admin/users"
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2 px-3 py-2 rounded-md transition-colors",
                isActive ? "bg-primary text-primary-foreground" : "hover:bg-muted"
              )
            }
          >
            <Users className="w-4 h-4" />
            Users
          </NavLink>
          <NavLink
            to="/admin/movies"
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2 px-3 py-2 rounded-md transition-colors",
                isActive ? "bg-primary text-primary-foreground" : "hover:bg-muted"
              )
            }
          >
            <Film className="w-4 h-4" />
            Content Overview
          </NavLink>
          <NavLink
            to="/admin/stats"
            className={({ isActive }) =>
              cn(
                "flex items-center gap-2 px-3 py-2 rounded-md transition-colors",
                isActive ? "bg-primary text-primary-foreground" : "hover:bg-muted"
              )
            }
          >
            <BarChart className="w-4 h-4" />
            Stats & Analytics
          </NavLink>
        </nav>
      </aside>
      
      <main className="flex-1 min-w-0">
        <Outlet />
      </main>
    </div>
  );
};
