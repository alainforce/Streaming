import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export interface User {
  id: string;
  email: string;
  role: 'user' | 'admin';
  status: 'active' | 'banned';
  created_at: string;
}

interface AuthState {
  token: string | null;
  user: User | null;
  login: (token: string, user: User) => void;
  logout: (serverCall?: boolean) => void;
  setUser: (user: User) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,

      login: (token, user) => set({ token, user }),

      logout: (_serverCall = true) => {
        // If needed, we can call server logout here or in the component.
        // It's cleaner to handle the API call in the component or a side effect, 
        // then call `logout()` here to clear state.
        set({ token: null, user: null });
      },

      setUser: (user) => set({ user }),
    }),
    {
      name: 'auth-storage', // Key for localStorage
    }
  )
);
