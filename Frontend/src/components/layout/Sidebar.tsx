import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Gamepad2, LogOut, User, BarChart2 } from 'lucide-react';
import { authService } from '../../services/auth.service';
import toast from 'react-hot-toast';

export const Sidebar: React.FC = () => {
  const navigate = useNavigate();

  const handleLogout = () => {
    authService.logout();
    toast.success('Logged out successfully');
    navigate('/login');
  };

  const navItems = [
    { name: 'Dashboard', path: '/dashboard', icon: LayoutDashboard },
    { name: 'Games', path: '/games', icon: Gamepad2 },
    { name: 'Analytics', path: '/analytics', icon: BarChart2 },
    { name: 'Profile', path: '/profile', icon: User },
  ];

  return (
    <aside className="w-64 bg-white border-r border-border h-screen flex flex-col fixed left-0 top-0">
      <div className="h-16 flex items-center px-6 border-b border-border">
        <h1 className="text-lg font-semibold tracking-tight text-foreground">DataPlay</h1>
      </div>
      
      <nav className="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
        {navItems.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive 
                    ? 'bg-primary text-foreground' 
                    : 'text-accent hover:bg-secondary hover:text-foreground'
                }`
              }
            >
              <Icon className="w-4 h-4" strokeWidth={2} />
              {item.name}
            </NavLink>
          );
        })}
      </nav>

      <div className="p-4 border-t border-border">
        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium text-accent hover:bg-red-50 hover:text-red-600 transition-colors"
        >
          <LogOut className="w-4 h-4" strokeWidth={2} />
          Logout
        </button>
      </div>
    </aside>
  );
};
