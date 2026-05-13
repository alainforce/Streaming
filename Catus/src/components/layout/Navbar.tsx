import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../store/useAuthStore';
import { Button } from '../ui/button';
import { Film, Search, User, LogOut, ShieldAlert, Heart, Eye } from 'lucide-react';
import { api } from '../../api/axios';

export const Navbar = () => {
  const { user, logout } = useAuthStore();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout failed on server', error);
    } finally {
      logout(false);
      navigate('/login');
    }
  };

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container mx-auto px-4 h-16 flex items-center justify-between">
        <Link to="/" className="flex items-center gap-2 group">
          <Film className="w-6 h-6 text-primary group-hover:scale-110 transition-transform" />
          <span className="font-bold text-xl tracking-tight hidden sm:inline-block">Catus</span>
        </Link>

        <div className="flex items-center gap-4">
          <Link to="/search">
            <Button variant="ghost" size="icon" title="Search">
              <Search className="w-5 h-5" />
              <span className="sr-only">Search</span>
            </Button>
          </Link>

          {!user ? (
            <div className="flex items-center gap-2">
              <Link to="/login">
                <Button variant="ghost">Log In</Button>
              </Link>
              <Link to="/signup">
                <Button>Sign Up</Button>
              </Link>
            </div>
          ) : (
            <>
              <nav className="flex items-center gap-1 sm:gap-2 mr-1 sm:mr-2">
                <Link to="/favorites">
                  <Button variant="ghost" className="hidden sm:flex px-2 sm:px-4">Favorites</Button>
                  <Button variant="ghost" size="icon" className="sm:hidden" title="Favorites">
                    <Heart className="w-5 h-5" />
                    <span className="sr-only">Favorites</span>
                  </Button>
                </Link>
                <Link to="/watched">
                  <Button variant="ghost" className="hidden sm:flex px-2 sm:px-4">Watched</Button>
                  <Button variant="ghost" size="icon" className="sm:hidden" title="Watched">
                    <Eye className="w-5 h-5" />
                    <span className="sr-only">Watched</span>
                  </Button>
                </Link>
                {user.role === 'admin' && (
                  <Link to="/admin/users">
                    <Button variant="ghost" className="hidden sm:flex px-2 sm:px-4">Admin</Button>
                    <Button variant="ghost" size="icon" className="sm:hidden" title="Admin Dashboard">
                      <ShieldAlert className="w-5 h-5" />
                      <span className="sr-only">Admin Dashboard</span>
                    </Button>
                  </Link>
                )}
              </nav>

              <div className="flex items-center gap-2">
                <Link to="/settings" className="inline-flex items-center justify-center rounded-full w-10 h-10 bg-secondary text-secondary-foreground hover:bg-secondary/80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2" title="Profile Settings">
                  <User className="w-5 h-5" />
                  <span className="sr-only">Profile Settings</span>
                </Link>

                <Button variant="ghost" size="icon" onClick={handleLogout} className="text-destructive hover:bg-destructive/10 hover:text-destructive" title="Log Out">
                  <LogOut className="w-5 h-5" />
                  <span className="sr-only">Log Out</span>
                </Button>
              </div>
            </>
          )}
        </div>
      </div>
    </header>
  );
};
