import { useMutation, useQueryClient, useQuery } from '@tanstack/react-query';
import { api } from '../api/axios';
import type { Movie } from '../components/shared/MovieCard';
import { useAuthStore } from '../store/useAuthStore';
import { toast } from 'sonner';

export const useMovieActions = () => {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();

  const { data: favorites = [] } = useQuery({
    queryKey: ['favorites'],
    queryFn: async () => {
      const { data } = await api.get('/favorites');
      return data.data; // Assumes response is { status: 'success', data: [...] }
    },
    enabled: !!user,
  });

  const { data: watched = [] } = useQuery({
    queryKey: ['watched'],
    queryFn: async () => {
      const { data } = await api.get('/watched');
      return data.data;
    },
    enabled: !!user,
  });

  const toggleFavoriteMutation = useMutation({
    mutationFn: async (movie: Movie) => {
      const isFav = favorites.some((f: any) => f.movie_id === movie.id);
      if (isFav) {
        await api.delete(`/favorites/${movie.id}`);
        return { movie, action: 'removed' };
      } else {
        await api.post('/favorites', {
          movie_id: movie.id,
          title: movie.title,
          overview: movie.overview,
          poster_path: movie.poster_path,
          vote_average: movie.vote_average,
        });
        return { movie, action: 'added' };
      }
    },
    onMutate: async (movie) => {
      await queryClient.cancelQueries({ queryKey: ['favorites'] });
      const previousFavorites = queryClient.getQueryData(['favorites']);
      const isFav = favorites.some((f: any) => f.movie_id === movie.id);

      queryClient.setQueryData(['favorites'], (old: any) => {
        if (!old) return old;
        if (isFav) {
          return old.filter((f: any) => f.movie_id !== movie.id);
        } else {
          return [...old, { 
            movie_id: movie.id, 
            title: movie.title, 
            overview: movie.overview, 
            poster_path: movie.poster_path, 
            vote_average: movie.vote_average,
            added_at: new Date().toISOString()
          }];
        }
      });
      return { previousFavorites };
    },
    onError: (err, _newMovie, context) => {
      console.error('Favorites mutation failed:', err);
      queryClient.setQueryData(['favorites'], context?.previousFavorites);
      toast.error('Failed to update favorites. Please try again.');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['favorites'] });
    },
    onSuccess: (data) => {
      toast.success(data.action === 'added' ? 'Added to favorites' : 'Removed from favorites');
    }
  });

  const toggleWatchedMutation = useMutation({
    mutationFn: async ({ movie, personalRating, watchedAt }: { movie: Movie, personalRating?: number | null, watchedAt?: string | null }) => {
      const isWatched = watched.some((w: any) => w.movie_id === movie.id);
      if (isWatched) {
        await api.delete(`/watched/${movie.id}`);
        return { movie, action: 'removed' };
      } else {
        await api.post('/watched', {
          movie_id: movie.id,
          title: movie.title,
          overview: movie.overview,
          poster_path: movie.poster_path,
          vote_average: movie.vote_average,
          personal_rating: personalRating ?? null,
          watched_at: watchedAt ?? new Date().toISOString(),
        });
        return { movie, action: 'added' };
      }
    },
    onMutate: async ({ movie }) => {
      await queryClient.cancelQueries({ queryKey: ['watched'] });
      const previousWatched = queryClient.getQueryData(['watched']);
      const isWatched = watched.some((w: any) => w.movie_id === movie.id);

      queryClient.setQueryData(['watched'], (old: any) => {
        if (!old) return old;
        if (isWatched) {
          return old.filter((w: any) => w.movie_id !== movie.id);
        } else {
          return [...old, { 
            movie_id: movie.id, 
            title: movie.title, 
            overview: movie.overview, 
            poster_path: movie.poster_path, 
            vote_average: movie.vote_average 
          }];
        }
      });
      return { previousWatched };
    },
    onError: (err, _vars, context) => {
      console.error('Watched mutation failed:', err);
      queryClient.setQueryData(['watched'], context?.previousWatched);
      toast.error('Failed to update watched list. Please try again.');
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['watched'] });
    },
    onSuccess: (data) => {
      toast.success(data.action === 'added' ? 'Marked as watched' : 'Removed from watched list');
    }
  });

  return {
    favorites,
    watched,
    toggleFavorite: toggleFavoriteMutation.mutate,
    isTogglingFavorite: toggleFavoriteMutation.isPending,
    toggleWatched: toggleWatchedMutation.mutate,
    isTogglingWatched: toggleWatchedMutation.isPending,
  };
};
