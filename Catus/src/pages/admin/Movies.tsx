import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { api, tmdbImageURL } from '../../api/axios';
import { format } from 'date-fns';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../../components/ui/table';
import { Badge } from '../../components/ui/badge';
import { PaginationControl } from '../../components/shared/PaginationControl';
import { Star } from 'lucide-react';

export const AdminMovies = () => {
  const [page, setPage] = useState(1);
  const pageSize = 20;

  const { data, isLoading, isError } = useQuery({
    queryKey: ['admin', 'movies', { page, pageSize }],
    queryFn: async () => {
      const { data } = await api.get('/admin/movies', { params: { page, page_size: pageSize } });
      return data;
    },
  });

  if (isLoading) return <div className="p-8 text-center text-muted-foreground">Loading movies...</div>;
  if (isError) return <div className="p-8 text-center text-destructive">Failed to load movies.</div>;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Content Overview</h2>
        <p className="text-muted-foreground">View all saved movies across all users.</p>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Movie</TableHead>
              <TableHead>Saved By</TableHead>
              <TableHead className="text-right">Rating</TableHead>
              <TableHead className="hidden md:table-cell text-right">Date Saved</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.data?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} className="h-24 text-center">
                  No movies saved yet.
                </TableCell>
              </TableRow>
            ) : (
              data?.data?.map((m: any) => (
                <TableRow key={m.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-14 bg-muted overflow-hidden rounded">
                        {m.poster_path && (
                          <img 
                            src={tmdbImageURL(m.poster_path, 'w185')} 
                            alt={m.title}
                            className="object-cover w-full h-full"
                          />
                        )}
                      </div>
                      <span className="font-medium line-clamp-2" title={m.title}>{m.title}</span>
                    </div>
                  </TableCell>
                  <TableCell className="text-muted-foreground">{m.user_email}</TableCell>
                  <TableCell className="text-right">
                    <Badge variant="secondary" className="font-semibold shadow-sm">
                      <Star className="w-3.5 h-3.5 mr-1 text-yellow-500 fill-yellow-500" />
                      {(m.vote_average || 0).toFixed(1)}
                    </Badge>
                  </TableCell>
                  <TableCell className="hidden md:table-cell text-right text-muted-foreground">
                    {format(new Date(m.added_at), 'MMM dd, yyyy')}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {data?.pagination && (
        <PaginationControl
          currentPage={data.pagination.page}
          totalPages={data.pagination.total_pages}
          onPageChange={setPage}
        />
      )}
    </div>
  );
};
