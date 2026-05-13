import { useMovieActions } from '../hooks/useMovieActions';
import { MovieCard, type Movie } from '../components/shared/MovieCard';
import { MovieGridSkeleton } from '../components/shared/MovieSkeleton';
import { WatchedModal } from '../components/shared/WatchedModal';
import { useState } from 'react';

export const Favorites = () => {
  // const navigate = useNavigate();
  const { favorites, watched, toggleFavorite, toggleWatched, isTogglingFavorite, isTogglingWatched } = useMovieActions();
  const [selectedWatchedMovie, setSelectedWatchedMovie] = useState<Movie | null>(null);

  // Consider it "loading" if we don't have the data structure yet, 
  // though react-query caches it. We can rely on favorites length for empty state.
  const isLoading = false; // We can add a robust loading state derived from useQuery if needed

  const handleToggleFavorite = (movie: Movie) => {
    toggleFavorite(movie);
  };

  const handleToggleWatched = (movie: Movie) => {
    const isWatched = watched.some((w: any) => w.movie_id === movie.id);
    if (isWatched) {
      toggleWatched({ movie });
    } else {
      setSelectedWatchedMovie(movie);
    }
  };

  return (
    <div className="flex flex-col gap-6 pt-4 pb-12 w-full">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight lg:text-4xl mb-2">My Favorites</h1>
          <p className="text-muted-foreground">Movies you've saved for later.</p>
        </div>
      </div>

      {isLoading ? (
        <MovieGridSkeleton count={10} />
      ) : favorites.length === 0 ? (
        <div className="w-full p-12 text-center border rounded-lg bg-card text-muted-foreground">
          You have no favorites yet. Start exploring movies.
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {favorites.map((fav: any) => {
            // Reconstruct movie object from favorite
            const movie: Movie = {
              id: fav.movie_id,
              title: fav.title,
              overview: fav.overview,
              poster_path: fav.poster_path,
              vote_average: fav.vote_average,
              release_date: '', // Provide dummy data if not saved
              vote_count: 0,
              popularity: 0,
              original_language: 'en',
              genre_ids: []
            };

            const isFav = true; // By definition
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
          onSave={(movie, rating, date) => {
            toggleWatched({ movie, personalRating: rating, watchedAt: date });
            setSelectedWatchedMovie(null);
          }}
        />
      )}
    </div>
  );
};
