import axios from 'axios';
import { useAuthStore } from '../store/useAuthStore';

// Retrieve base URL from environment
const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

export const api = axios.create({
  baseURL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const tmdbImageURL = (path: string, size: 'w185' | 'w342' | 'w500' | 'w780' | 'original' = 'w500') => {
  if (!path) return '/placeholder-poster.png'; // Add a fallback image in public folder later
  const baseUrl = import.meta.env.VITE_TMDB_IMAGE_BASE_URL || 'https://image.tmdb.org/t/p';
  return `${baseUrl}/${size}${path}`;
};

// Request Interceptor: Attach Auth Token
api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().token;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response Interceptor: Handle Global Errors like 401
api.interceptors.response.use(
  (response) => response,
  (error) => {
    // Check if the error is 401
    if (error.response?.status === 401) {
      useAuthStore.getState().logout(false);
      // Optional: Since React Router v6 is declarative, we can handle navigation
      // in the router setup or a small wrapper component. 
      // But clearing state here automatically kicks them out of protected routes.
    }
    return Promise.reject(error);
  }
);
