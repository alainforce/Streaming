import { useQuery } from '@tanstack/react-query';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useState, useEffect } from 'react';
import { api } from '../api/axios';
import { MovieCard, type Movie } from '../components/shared/MovieCard';
import { MovieGridSkeleton } from '../components/shared/MovieSkeleton';
import { PaginationControl } from '../components/shared/PaginationControl';
import { useMovieActions } from '../hooks/useMovieActions';
import { useDebounce } from '../hooks/useDebounce';
import { useAuthStore } from '../store/useAuthStore';
import { Input } from '../components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Alert, AlertDescription, AlertTitle } from '../components/ui/alert';
import { AlertCircle } from 'lucide-react';
import { toast } from 'sonner';
import { WatchedModal } from '../components/shared/WatchedModal';

const validSorts = [
  { value: 'popularity.desc', label: 'Popularity Desc' },
  { value: 'popularity.asc', label: 'Popularity Asc' },
  { value: 'vote_average.desc', label: 'Rating Desc' },
  { value: 'vote_average.asc', label: 'Rating Asc' },
  { value: 'primary_release_date.desc', label: 'Date Desc' },
  { value: 'primary_release_date.asc', label: 'Date Asc' },
];

export const Search = () => {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { favorites, watched, toggleFavorite, toggleWatched, isTogglingFavorite, isTogglingWatched } = useMovieActions();
  const [selectedWatchedMovie, setSelectedWatchedMovie] = useState<Movie | null>(null);

  const initialQ = searchParams.get('q') || '';
  const [searchTerm, setSearchTerm] = useState(initialQ);
  const debouncedSearch = useDebounce(searchTerm, 500);

  const genre = searchParams.get('genre') || '';
  const year = searchParams.get('year') || '';
  const sort_by = searchParams.get('sort_by') || 'popularity.desc';
  const page = parseInt(searchParams.get('page') || '1', 10);

  useEffect(() => {
    if (debouncedSearch !== searchParams.get('q')) {
      const newParams = new URLSearchParams(searchParams);
      if (debouncedSearch) {
        newParams.set('q', debouncedSearch);
      } else {
        newParams.delete('q');
      }
      newParams.set('page', '1');
      setSearchParams(newParams, { replace: true });
    }
  }, [debouncedSearch, searchParams, setSearchParams]);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'instant' });
  }, [page, debouncedSearch, genre, year, sort_by]);

  const { data: genresData } = useQuery({
    queryKey: ['genres'],
    queryFn: async () => {
      const { data } = await api.get('/movies/genres');
      return data.data as { id: number; name: string }[];
    },
    staleTime: 1000 * 60 * 60 * 24, // 24 hours
  });

  const { data: searchData, isLoading, isError, error } = useQuery({
    queryKey: ['search', debouncedSearch, genre, year, sort_by, page],
    queryFn: async () => {
      if (!debouncedSearch && !genre && !year) {
        // Return mostly empty if nothing is selected yet, or force fetch without q
        // Actually, if browse mode, at least one is required by backend if not handled
        const { data } = await api.get('/movies/search', {
          params: { genre, year, sort_by, page }
        });
        return data.data;
      }
      const { data } = await api.get('/movies/search', {
        params: { q: debouncedSearch, genre, year, sort_by, page }
      });
      return data.data;
    },
    enabled: !!debouncedSearch || !!genre || !!year,
  });

  const handleFilterChange = (key: string, value: string) => {
    const newParams = new URLSearchParams(searchParams);
    if (value && value !== 'all') {
      newParams.set(key, value);
    } else {
      newParams.delete(key);
    }
    newParams.set('page', key === 'page' ? value : '1');
    setSearchParams(newParams);
  };

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
      toggleWatched({ movie });
    } else {
      setSelectedWatchedMovie(movie);
    }
  };

  return (
    <div className="flex flex-col gap-6 pt-4 pb-12 w-full">
      <div className="flex flex-col gap-4 bg-muted/30 p-4 rounded-xl border border-border">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Input 
            type="search" 
            placeholder="Search keywords..." 
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="md:col-span-4"
          />
          <Select value={genre || 'all'} onValueChange={(val) => handleFilterChange('genre', val || '')}>
            <SelectTrigger>
              <SelectValue placeholder="All Genres" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Genres</SelectItem>
              {genresData?.map((g) => (
                <SelectItem key={g.id} value={g.id.toString()}>{g.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input 
            type="number" 
            placeholder="Year (e.g. 2023)" 
            value={year}
            onChange={(e) => handleFilterChange('year', e.target.value)}
          />
          <Select value={sort_by} onValueChange={(val) => handleFilterChange('sort_by', val || '')}>
            <SelectTrigger className="md:col-span-2">
              <SelectValue placeholder="Sort By" />
            </SelectTrigger>
            <SelectContent>
              {validSorts.map((s) => (
                <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {searchData?.warning && (
        <Alert variant="default" className="bg-yellow-500/20 text-yellow-600 border-yellow-500/50">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Warning</AlertTitle>
          <AlertDescription>{searchData.warning}</AlertDescription>
        </Alert>
      )}

      {(!debouncedSearch && !genre && !year) ? (
        <div className="w-full p-12 text-center border rounded-lg bg-card text-muted-foreground">
          Enter a search term or select filters to discover movies.
        </div>
      ) : isLoading ? (
        <MovieGridSkeleton />
      ) : isError ? (
        <div className="w-full p-8 text-center border rounded-lg bg-destructive/10 text-destructive">
          Failed to fetch results. {(error as any)?.response?.data?.error || 'Unknown error.'}
        </div>
      ) : searchData?.results?.length === 0 ? (
        <div className="w-full p-12 text-center border rounded-lg bg-card text-muted-foreground">
          No results found for your search.
        </div>
      ) : (
        <>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
            {searchData?.results?.map((movie: Movie) => {
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

          <PaginationControl
            currentPage={page}
            totalPages={searchData?.total_pages || 1}
            onPageChange={(p) => handleFilterChange('page', p.toString())}
          />
        </>
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
