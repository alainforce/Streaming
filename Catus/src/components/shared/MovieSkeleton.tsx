import { Card, CardContent } from '../ui/card';
import { Skeleton } from '../ui/skeleton';

export const MovieSkeleton = () => {
  return (
    <Card className="overflow-hidden flex flex-col h-full bg-card">
      <Skeleton className="w-full aspect-[2/3] rounded-none" />
      <CardContent className="p-4 flex flex-col flex-1 gap-2">
        <Skeleton className="h-5 w-4/5" />
        <Skeleton className="h-4 w-1/4 mb-4" />
        
        <div className="flex items-center gap-2 mt-auto pt-2">
          <Skeleton className="h-8 flex-1" />
          <Skeleton className="h-8 flex-1" />
        </div>
      </CardContent>
    </Card>
  );
};

export const MovieGridSkeleton = ({ count = 20 }: { count?: number }) => {
  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <MovieSkeleton key={i} />
      ))}
    </div>
  );
};
