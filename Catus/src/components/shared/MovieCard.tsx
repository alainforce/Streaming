import { Star, Eye, Heart, Plus, Info } from 'lucide-react';
import { Card, CardContent } from '../ui/card';
import { Button } from '../ui/button';
import { Badge } from '../ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '../ui/dialog';
import { tmdbImageURL } from '../../api/axios';
export interface Movie {
  id: number;
  title: string;
  overview: string;
  release_date: string;
  poster_path: string;
  vote_average: number;
  vote_count: number;
  popularity: number;
  original_language: string;
  genre_ids: number[];
}

interface MovieCardProps {
  movie: Movie;
  isFavorite?: boolean;
  isWatched?: boolean;
  onToggleFavorite?: (movie: Movie) => void;
  onToggleWatched?: (movie: Movie) => void;
  // If loading, optimistic UI could set these
  isTogglingFavorite?: boolean;
  isTogglingWatched?: boolean;
}

export const MovieCard = ({
  movie,
  isFavorite = false,
  isWatched = false,
  onToggleFavorite,
  onToggleWatched,
  isTogglingFavorite = false,
  isTogglingWatched = false,
}: MovieCardProps) => {
  const year = movie.release_date?.substring(0, 4) || 'Unknown';

  return (
    <Dialog>
      <Card className="overflow-hidden flex flex-col h-full bg-card group relative">
        <DialogTrigger className="relative aspect-[2/3] overflow-hidden bg-muted w-full block text-left outline-none cursor-pointer focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2">
          {movie.poster_path ? (
            <img
              src={tmdbImageURL(movie.poster_path, 'w342')}
              alt={movie.title}
              className="w-full h-full object-cover transition-transform duration-300 group-hover:scale-105"
              loading="lazy"
            />
          ) : (
            <div className="w-full h-full flex items-center justify-center text-muted-foreground p-4 text-center">
              No Poster
            </div>
          )}
          <div className="absolute inset-0 bg-black/60 opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex items-center justify-center pointer-events-none">
            <Info className="w-12 h-12 text-white/90" />
            <span className="sr-only">View Overview</span>
          </div>
          <div className="absolute top-2 right-2 flex flex-col gap-2 pointer-events-none">
            <Badge variant="secondary" className="font-semibold px-2 py-1 shadow-md bg-background/80 backdrop-blur-md">
              <Star className="w-3.5 h-3.5 mr-1 text-yellow-500 fill-yellow-500" />
              {(movie.vote_average || 0).toFixed(1)}
            </Badge>
          </div>
        </DialogTrigger>
      <CardContent className="p-4 flex flex-col flex-1">
        <div className="flex-1">
          <h3 className="font-bold text-lg leading-tight mb-1 line-clamp-1" title={movie.title}>
            {movie.title}
          </h3>
          <p className="text-sm text-muted-foreground mb-4">{year}</p>
        </div>
        <div className="flex flex-col gap-2 mt-auto pt-4">
          <Button
            variant={isFavorite ? "secondary" : "outline"}
            className="w-full transition-colors"
            size="sm"
            disabled={isTogglingFavorite}
            onClick={() => onToggleFavorite?.(movie)}
            title={isFavorite ? "Remove from favorites" : "Add to favorites"}
          >
            <Heart className={`w-4 h-4 mr-1.5 ${isFavorite ? "fill-destructive text-destructive" : ""}`} />
            {isFavorite ? "Saved to Favorites" : "Add to Favorites"}
          </Button>
          <Button
            variant={isWatched ? "secondary" : "default"}
            className={`w-full transition-colors ${isWatched ? "bg-primary/20 text-primary hover:bg-primary/30" : ""}`}
            size="sm"
            disabled={isTogglingWatched}
            onClick={() => onToggleWatched?.(movie)}
            title={isWatched ? "Already watched" : "Mark as watched"}
          >
            {isWatched ? (
              <>
                <Eye className="w-4 h-4 mr-1.5 fill-primary text-primary" />
                Watched
              </>
            ) : (
              <>
                <Plus className="w-4 h-4 mr-1.5" />
                Add to Watched
              </>
            )}
          </Button>
        </div>
      </CardContent>
      <DialogContent className="sm:max-w-[700px] flex flex-col sm:flex-row gap-6 p-6 max-h-[85vh] overflow-y-auto">
        {movie.poster_path && (
          <div className="hidden sm:block sm:w-[240px] flex-shrink-0">
            <img 
              src={tmdbImageURL(movie.poster_path, 'w500')} 
              alt={movie.title} 
              className="w-full rounded-md object-cover shadow-sm"
            />
          </div>
        )}
        <div className="flex-1 flex flex-col mt-2 sm:mt-0">
          <DialogHeader className="mb-4 text-left">
            <DialogTitle className="text-3xl font-bold leading-tight">
              {movie.title} <span className="text-muted-foreground font-normal">({year})</span>
            </DialogTitle>
            <DialogDescription className="flex items-center gap-4 mt-2 text-foreground">
              <span className="flex items-center font-semibold bg-secondary px-2 py-0.5 rounded-md">
                <Star className="w-4 h-4 mr-1 text-yellow-500 fill-yellow-500" />
                {(movie.vote_average || 0).toFixed(1)}
              </span>
              <span className="text-sm text-muted-foreground">{movie.vote_count} votes</span>
            </DialogDescription>
          </DialogHeader>
          <div className="text-base text-foreground/90 leading-relaxed text-left">
            <h4 className="font-semibold mb-2">Overview</h4>
            {movie.overview || "No overview available for this movie."}
          </div>
        </div>
      </DialogContent>
    </Card>
    </Dialog>
  );
};
