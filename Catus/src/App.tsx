import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './store/useAuthStore';
import { Toaster } from 'sonner';

// Pages (to be implemented)
import { Home } from './pages/Home';
import { Search } from './pages/Search';
import { Login } from './pages/Login';
import { Signup } from './pages/Signup';
import { Favorites } from './pages/Favorites';
import { Watched } from './pages/Watched';
import { Settings } from './pages/Settings';
import { AdminLayout } from './pages/admin/AdminLayout';
import { AdminUsers } from './pages/admin/Users';
import { AdminMovies } from './pages/admin/Movies';
import { AdminStats } from './pages/admin/Stats';

const ProtectedRoute = ({ children, requireAdmin = false }: { children: React.ReactNode, requireAdmin?: boolean }) => {
  const { token, user } = useAuthStore();
  
  if (!token) return <Navigate to="/login" replace />;
  
  if (requireAdmin && user?.role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  return <>{children}</>;
};

import { Navbar } from './components/layout/Navbar';

function App() {
  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col items-center w-full">
      <Navbar />

      <main className="flex-1 w-full max-w-7xl mx-auto p-4 flex flex-col">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/search" element={<Search />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<Signup />} />
          
          <Route path="/favorites" element={<ProtectedRoute><Favorites /></ProtectedRoute>} />
          <Route path="/watched" element={<ProtectedRoute><Watched /></ProtectedRoute>} />
          <Route path="/settings" element={<ProtectedRoute><Settings /></ProtectedRoute>} />
          
          <Route path="/admin" element={<ProtectedRoute requireAdmin><AdminLayout /></ProtectedRoute>}>
            <Route index element={<Navigate to="stats" replace />} />
            <Route path="users" element={<AdminUsers />} />
            <Route path="movies" element={<AdminMovies />} />
            <Route path="stats" element={<AdminStats />} />
          </Route>
        </Routes>
      </main>

      <Toaster position="bottom-right" />
    </div>
  );
}

export default App;
