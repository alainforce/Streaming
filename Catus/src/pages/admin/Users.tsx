import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { api } from '../../api/axios';
import { format } from 'date-fns';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../../components/ui/table';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogClose,
} from '../../components/ui/dialog';
import { Button } from '../../components/ui/button';
import { Badge } from '../../components/ui/badge';
import { PaginationControl } from '../../components/shared/PaginationControl';
import { toast } from 'sonner';

export const AdminUsers = () => {
  const [page, setPage] = useState(1);
  const pageSize = 20;
  const queryClient = useQueryClient();

  const { data, isLoading, isError } = useQuery({
    queryKey: ['admin', 'users', { page, pageSize }],
    queryFn: async () => {
      const { data } = await api.get('/admin/users', { params: { page, page_size: pageSize } });
      return data;
    },
  });

  const banMutation = useMutation({
    mutationFn: async ({ id, isBanned }: { id: string, isBanned: boolean }) => {
      if (isBanned) {
        await api.patch(`/admin/users/${id}/unban`);
        return { id, status: 'active' };
      } else {
        await api.patch(`/admin/users/${id}/ban`);
        return { id, status: 'banned' };
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
      toast.success('User status updated');
    },
    onError: () => toast.error('Failed to update user status')
  });

  const deleteMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/admin/users/${id}`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] });
      toast.success('User deleted successfully');
    },
    onError: () => toast.error('Failed to delete user')
  });

  const handleBanToggle = (id: string, currentStatus: string) => {
    banMutation.mutate({ id, isBanned: currentStatus === 'banned' });
  };

  if (isLoading) return <div className="p-8 text-center text-muted-foreground">Loading users...</div>;
  if (isError) return <div className="p-8 text-center text-destructive">Failed to load users.</div>;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">User Management</h2>
        <p className="text-muted-foreground">View and manage all users on the platform.</p>
      </div>

      <div className="rounded-md border bg-card">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="hidden md:table-cell">Joined</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {data?.data?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  No users found.
                </TableCell>
              </TableRow>
            ) : (
              data?.data?.map((u: any) => (
                <TableRow key={u.id}>
                  <TableCell className="font-medium">{u.email}</TableCell>
                  <TableCell className="capitalize">{u.role}</TableCell>
                  <TableCell>
                    <Badge variant={u.status === 'banned' ? 'destructive' : 'default'} className={u.status === 'active' ? 'bg-green-500 hover:bg-green-600 outline-none border-none' : ''}>
                      {u.status}
                    </Badge>
                  </TableCell>
                  <TableCell className="hidden md:table-cell text-muted-foreground">
                    {format(new Date(u.created_at), 'MMM dd, yyyy')}
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    {u.role !== 'admin' && ( // Prevent admin self-ban visually
                      <>
                        <Button 
                          variant="outline" 
                          size="sm"
                          onClick={() => handleBanToggle(u.id, u.status)}
                          disabled={banMutation.isPending}
                        >
                          {u.status === 'banned' ? 'Unban' : 'Ban'}
                        </Button>
                        <Dialog>
                          <DialogTrigger className="inline-flex items-center justify-center rounded-md text-sm font-medium bg-destructive text-destructive-foreground hover:bg-destructive/90 h-8 px-3 ml-2">
                            Delete
                          </DialogTrigger>
                          <DialogContent>
                            <DialogHeader>
                              <DialogTitle>Are you absolutely sure?</DialogTitle>
                              <DialogDescription>
                                This action cannot be undone. This will permanently delete user {u.email} and all their saved data.
                              </DialogDescription>
                            </DialogHeader>
                            <DialogFooter>
                              <DialogClose render={<Button variant="outline" />}>
                                Cancel
                              </DialogClose>
                              <Button 
                                variant="destructive" 
                                onClick={() => deleteMutation.mutate(u.id)}
                                disabled={deleteMutation.isPending}
                              >
                                {deleteMutation.isPending ? 'Deleting...' : 'Yes, delete user'}
                              </Button>
                            </DialogFooter>
                          </DialogContent>
                        </Dialog>
                      </>
                    )}
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
