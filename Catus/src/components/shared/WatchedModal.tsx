import { useState } from 'react';
import type { Movie } from './MovieCard';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';

interface WatchedModalProps {
  movie: Movie | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (movie: Movie, rating?: number, date?: string) => void;
}

export const WatchedModal = ({ movie, open, onOpenChange, onSave }: WatchedModalProps) => {
  const [rating, setRating] = useState<string>('');
  
  // Default to current year
  const currentYear = new Date().getFullYear().toString();
  const [year, setYear] = useState<string>(currentYear);

  const handleSave = () => {
    if (!movie) return;

    let finalRating: number | undefined;
    if (rating) {
      const pRating = parseFloat(rating);
      if (pRating >= 1 && pRating <= 10) {
        finalRating = pRating;
      }
    }

    // Convert year string to ISO representation of Jan 1st of that year
    let finalDate: string | undefined;
    if (year && year.length === 4) {
      finalDate = new Date(`${year}-01-01T00:00:00.000Z`).toISOString();
    }

    onSave(movie, finalRating, finalDate);
  };

  if (!movie) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Mark "{movie.title}" as Watched</DialogTitle>
          <DialogDescription>
            You can optionally add a personal rating (1-10) and the year you watched it.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="rating" className="text-right">
              Rating
            </Label>
            <Input
              id="rating"
              type="number"
              min="1"
              max="10"
              step="0.5"
              placeholder="1-10"
              className="col-span-3"
              value={rating}
              onChange={(e) => setRating(e.target.value)}
            />
          </div>
          <div className="grid grid-cols-4 items-center gap-4">
            <Label htmlFor="year" className="text-right">
              Year
            </Label>
            <Input
              id="year"
              type="number"
              min="1900"
              max={currentYear}
              className="col-span-3"
              value={year}
              onChange={(e) => setYear(e.target.value)}
              placeholder="e.g. 2024"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button>
          <Button onClick={handleSave}>Save</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
