import { useQuery } from '@tanstack/react-query';
import { api } from '../../api/axios';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../../components/ui/card';
import { Users, Activity, Ban, Film, Heart } from 'lucide-react';

export const AdminStats = () => {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['admin', 'stats'],
    queryFn: async () => {
      const response = await api.get('/admin/stats');
      return response.data.data;
    },
  });

  if (isLoading) return <div className="p-8 text-center text-muted-foreground">Loading dashboard...</div>;
  if (isError) return <div className="p-8 text-center text-destructive">Failed to load dashboard.</div>;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Stats & Analytics</h2>
        <p className="text-muted-foreground">Platform-wide statistics and usage data.</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.total_users || 0}</div>
            <p className="text-xs text-muted-foreground">
              +{data?.new_users_last_7_days || 0} in last 7 days
            </p>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Users</CardTitle>
            <Activity className="h-4 w-4 text-green-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.active_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Banned Users</CardTitle>
            <Ban className="h-4 w-4 text-destructive" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.banned_users || 0}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Favorites</CardTitle>
            <Heart className="h-4 w-4 text-pink-500 fill-pink-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{data?.total_favorites_saved || 0}</div>
            <p className="text-xs text-muted-foreground">
              +{data?.new_favorites_last_7_days || 0} in last 7 days
            </p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-1 lg:grid-cols-2">
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Top 10 Saved Movies</CardTitle>
            <CardDescription>Most frequently favorited content across all users.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {data?.top_saved_movies?.map((movie: any, idx: number) => (
                <div key={movie.movie_id} className="flex items-center">
                  <span className="w-6 text-center text-sm font-medium text-muted-foreground mr-4">
                    {idx + 1}
                  </span>
                  <Film className="h-4 w-4 text-muted-foreground mr-3" />
                  <div className="ml-0 space-y-1 overflow-hidden flex-1">
                    <p className="text-sm font-medium leading-none truncate pr-4" title={movie.title}>{movie.title}</p>
                  </div>
                  <div className="ml-auto font-medium tabular-nums px-2 py-0.5 bg-muted rounded text-sm">
                    {movie.save_count}
                  </div>
                </div>
              ))}
              {!data?.top_saved_movies?.length && (
                <div className="text-center text-sm text-muted-foreground py-4">No data available yet</div>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
};
