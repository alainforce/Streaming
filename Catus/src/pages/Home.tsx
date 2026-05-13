import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api } from '../api/axios';
import { MovieCard, type Movie } from '../components/shared/MovieCard';
import { MovieGridSkeleton } from '../components/shared/MovieSkeleton';
import { useMovieActions } from '../hooks/useMovieActions';
import { useAuthStore } from '../store/useAuthStore';
import { toast } from 'sonner';
import { useState } from 'react';
import { WatchedModal } from '../components/shared/WatchedModal';

export const Home = () => {
  const { user } = useAuthStore();
  const navigate = useNavigate();
  const { favorites, watched, toggleFavorite, toggleWatched, isTogglingFavorite, isTogglingWatched } = useMovieActions();
  
  const [selectedWatchedMovie, setSelectedWatchedMovie] = useState<Movie | null>(null);

  const { data, isLoading, isError } = useQuery({
    queryKey: ['trending'],
    queryFn: async () => {
      const response = await api.get('/movies/trending');
      return response.data.data as Movie[];
    },
  });

  const handleToggleFavorite = (movie: Movie) => {
    if (!user) {
      toast.info('Please log in to save favorites');
      navigate('/login');
      return;
    }
    toggleFavorite(movie);
  };

  const handleToggleWatched = (movie: Movie) => {
    if (!user) {
      toast.info('Please log in to mark movies as watched');
      navigate('/login');
      return;
    }

    const isWatched = watched.some((w: any) => w.movie_id === movie.id);
    if (isWatched) {
      // If removing from watched, just toggle it
      toggleWatched({ movie });
    } else {
      // If adding, show modal for rating and date
      setSelectedWatchedMovie(movie);
    }
  };

  const handleSaveWatched = (movie: Movie, rating?: number, date?: string) => {
    toggleWatched({ movie, personalRating: rating, watchedAt: date });
    setSelectedWatchedMovie(null);
  };

  return (
    <div className="flex flex-col gap-6 pt-4 pb-12 w-full">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight lg:text-4xl mb-2">Trending Now</h1>
          <p className="text-muted-foreground">Discover the most popular movies today.</p>
        </div>
      </div>

      {isLoading ? (
        <MovieGridSkeleton />
      ) : isError ? (
        <div className="w-full p-8 text-center border rounded-lg bg-destructive/10 text-destructive">
          Failed to load trending movies. Please try again later.
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {data?.map((movie) => {
            const isFav = favorites.some((f: any) => f.movie_id === movie.id);
            const isWtch = watched.some((w: any) => w.movie_id === movie.id);

            return (
              <MovieCard
                key={movie.id}
                movie={movie}
                isFavorite={isFav}
                isWatched={isWtch}
                onToggleFavorite={handleToggleFavorite}
                onToggleWatched={handleToggleWatched}
                isTogglingFavorite={isTogglingFavorite}
                isTogglingWatched={isTogglingWatched}
              />
            );
          })}
        </div>
      )}

      {selectedWatchedMovie && (
        <WatchedModal
          movie={selectedWatchedMovie}
          open={!!selectedWatchedMovie}
          onOpenChange={(open) => !open && setSelectedWatchedMovie(null)}
          onSave={handleSaveWatched}
        />
      )}
    </div>
  );
};
