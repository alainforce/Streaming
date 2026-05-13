import { useMovieActions } from '../hooks/useMovieActions';
import { MovieCard, type Movie } from '../components/shared/MovieCard';
// import { MovieGridSkeleton } from '../components/shared/MovieSkeleton';
import { Badge } from '../components/ui/badge';

export const Watched = () => {
  const { favorites, watched, toggleFavorite, toggleWatched, isTogglingFavorite, isTogglingWatched } = useMovieActions();

  const handleToggleFavorite = (movie: Movie) => {
    toggleFavorite(movie);
  };

  const handleToggleWatched = (movie: Movie) => {
    toggleWatched({ movie }); // If we are here, it's already watched, so this un-watches it
  };

  return (
    <div className="flex flex-col gap-6 pt-4 pb-12 w-full">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight lg:text-4xl mb-2">My Watched List</h1>
          <p className="text-muted-foreground">Movies you've already seen.</p>
        </div>
      </div>

      {watched.length === 0 ? (
        <div className="w-full p-12 text-center border rounded-lg bg-card text-muted-foreground">
          Your watched list is empty. Log a movie you've seen.
        </div>
      ) : (
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {watched.map((w: any) => {
            const movie: Movie = {
              id: w.movie_id,
              title: w.title,
              overview: w.overview,
              poster_path: w.poster_path,
              vote_average: w.vote_average,
              release_date: '',
              vote_count: 0,
              popularity: 0,
              original_language: '',
              genre_ids: []
            };

            const isFav = favorites.some((f: any) => f.movie_id === movie.id);
            const isWtch = true;

            // Optional: You could create a specialized WatchedCard if you want to display the personal rating
            // For now, we'll overlay the personal rating if it exists
            return (
              <div key={movie.id} className="relative group">
                <MovieCard
                  movie={movie}
                  isFavorite={isFav}
                  isWatched={isWtch}
                  onToggleFavorite={handleToggleFavorite}
                  onToggleWatched={handleToggleWatched}
                  isTogglingFavorite={isTogglingFavorite}
                  isTogglingWatched={isTogglingWatched}
                />
                {w.personal_rating && (
                  <Badge className="absolute top-2 left-2 bg-primary text-primary-foreground font-bold z-20 shadow-md">
                    ★ {w.personal_rating}/10
                  </Badge>
                )}
                {w.watched_at && (
                  <Badge variant="secondary" className="absolute top-10 right-2 bg-black/80 backdrop-blur-sm text-white font-semibold z-20 shadow-md border-none">
                    Watched: {new Date(w.watched_at).getFullYear()}
                  </Badge>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};
