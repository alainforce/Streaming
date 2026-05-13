import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '../store/useAuthStore';
import { api } from '../api/axios';
import { Card, CardContent, CardDescription, CardHeader, CardTitle, CardFooter } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../components/ui/dialog';
import { toast } from 'sonner';
import { useNavigate } from 'react-router-dom';
import { format } from 'date-fns';
import { Loader2 } from 'lucide-react';

export const Settings = () => {
  const { user, setUser, logout } = useAuthStore();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const [isDeleting, setIsDeleting] = useState(false);
  const [open, setOpen] = useState(false);
  
  // Email Form State
  const [newEmail, setNewEmail] = useState(user?.email || '');
  const [emailError, setEmailError] = useState('');

  // Password Form State
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');

  if (!user) return null;

  const { data: profileData, isLoading: isProfileLoading } = useQuery({
    queryKey: ['settings', 'profile'],
    queryFn: async () => {
      const response = await api.get('/settings/profile');
      return response.data.data;
    },
  });

  const updateEmailMutation = useMutation({
    mutationFn: async (email: string) => {
      const response = await api.patch('/settings/email', { email });
      return response.data.data;
    },
    onSuccess: (updatedUser) => {
      // Update the global auth store securely
      setUser({ ...user, email: updatedUser.email });
      queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] });
      toast.success('Email updated successfully');
      setEmailError('');
    },
    onError: (error: any) => {
      if (error.response?.status === 409) {
        setEmailError('This email is already in use.');
      } else if (error.response?.status === 400) {
        setEmailError('Invalid email format.');
      } else {
        toast.error('Failed to update email. Please try again.');
      }
    }
  });

  const handleUpdateEmail = (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    if (!newEmail || newEmail === user.email) return;
    updateEmailMutation.mutate(newEmail);
  };

  const updatePasswordMutation = useMutation({
    mutationFn: async (password: string) => {
      const response = await api.patch('/settings/password', { new_password: password });
      return response.data;
    },
    onSuccess: () => {
      toast.success('Password updated successfully');
      setNewPassword('');
      setConfirmPassword('');
      setPasswordError('');
    },
    onError: (error: any) => {
      if (error.response?.status === 400) {
        setPasswordError('Password too short (minimum 8 characters).');
      } else {
        toast.error('Failed to update password. Please try again.');
      }
    }
  });

  const handleUpdatePassword = (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    if (!newPassword || newPassword.length < 8) {
      setPasswordError('Password must be at least 8 characters long.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError('Passwords do not match.');
      return;
    }
    updatePasswordMutation.mutate(newPassword);
  };

  const handleSecurityLogout = async () => {
    try {
      await api.post('/auth/logout');
    } catch (error) {
      console.error('Logout failed on server', error);
    } finally {
      logout(false);
      navigate('/login');
    }
  };

  const handleDeleteAccount = async () => {
    setIsDeleting(true);
    try {
      await api.delete('/auth/account');
      toast.success('Account deleted successfully');
      setOpen(false);
      logout(false);
      navigate('/');
    } catch (error) {
      toast.error('Failed to delete account');
    } finally {
      setIsDeleting(false);
    }
  };

  const formattedDate = profileData?.profile?.created_at 
    ? format(new Date(profileData.profile.created_at), 'MMMM dd, yyyy') 
    : user.created_at ? format(new Date(user.created_at), 'MMMM dd, yyyy') : 'Unknown';
    
  const lastActivity = profileData?.stats?.most_recent_activity
    ? format(new Date(profileData.stats.most_recent_activity), 'MMMM dd, yyyy')
    : 'No activity yet';

  return (
    <div className="flex flex-col gap-6 pt-4 pb-12 w-full max-w-2xl mx-auto">
      <div className="flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-extrabold tracking-tight lg:text-4xl mb-2">Account Settings</h1>
          <p className="text-muted-foreground">Manage your account preferences.</p>
        </div>
      </div>

      {/* SECTION 1 - Profile Overview */}
      <Card>
        <CardHeader>
          <CardTitle>Profile Details</CardTitle>
          <CardDescription>Your personal information and activity statistics.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {isProfileLoading ? (
            <div className="flex justify-center py-4">
              <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
            </div>
          ) : (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Email</p>
                  <p className="text-lg">{profileData?.profile?.email || user.email}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Role</p>
                  <p className="text-lg capitalize">{profileData?.profile?.role || user.role}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Status</p>
                  <p className="text-lg capitalize">{profileData?.profile?.status || user.status}</p>
                </div>
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Member Since</p>
                  <p className="text-lg">{formattedDate}</p>
                </div>
              </div>

              <div className="pt-4 border-t">
                <h4 className="font-medium mb-4 text-sm text-muted-foreground uppercase tracking-wider">Activity Stats</h4>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div className="bg-muted/50 p-4 rounded-lg">
                    <p className="text-sm text-muted-foreground mb-1">Total Favorites Saved</p>
                    <p className="text-2xl font-bold">{profileData?.stats?.total_favorites || 0}</p>
                  </div>
                  <div className="bg-muted/50 p-4 rounded-lg">
                    <p className="text-sm text-muted-foreground mb-1">Total Movies Watched</p>
                    <p className="text-2xl font-bold">{profileData?.stats?.total_watched || 0}</p>
                  </div>
                  <div className="bg-muted/50 p-4 rounded-lg">
                    <p className="text-sm text-muted-foreground mb-1">Last Activity</p>
                    <p className="text-lg font-medium">{lastActivity}</p>
                  </div>
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* SECTION 2 - Update Email */}
      <Card>
        <CardHeader>
          <CardTitle>Email Address</CardTitle>
          <CardDescription>Update the email address associated with your account.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleUpdateEmail} className="space-y-4 max-w-md">
            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium">New Email</label>
              <Input
                id="email"
                type="email"
                value={newEmail}
                onChange={(e) => {
                  setNewEmail(e.target.value);
                  if (emailError) setEmailError('');
                }}
                disabled={updateEmailMutation.isPending}
              />
              {emailError && <p className="text-sm text-destructive">{emailError}</p>}
            </div>
            <Button 
              type="submit" 
              disabled={updateEmailMutation.isPending || !newEmail || newEmail === user.email}
            >
              {updateEmailMutation.isPending ? 'Saving...' : 'Save Changes'}
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* SECTION 3 - Update Password */}
      <Card>
        <CardHeader>
          <CardTitle>Password</CardTitle>
          <CardDescription>Update your password to keep your account secure.</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleUpdatePassword} className="space-y-4 max-w-md">
            <div className="space-y-2">
              <label htmlFor="newPassword" className="text-sm font-medium">New Password</label>
              <Input
                id="newPassword"
                type="password"
                value={newPassword}
                onChange={(e) => {
                  setNewPassword(e.target.value);
                  if (passwordError) setPasswordError('');
                }}
                disabled={updatePasswordMutation.isPending}
              />
            </div>
            <div className="space-y-2">
              <label htmlFor="confirmPassword" className="text-sm font-medium">Confirm New Password</label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value);
                  if (passwordError) setPasswordError('');
                }}
                disabled={updatePasswordMutation.isPending}
              />
              {passwordError && <p className="text-sm text-destructive">{passwordError}</p>}
            </div>
            <div className="flex flex-col sm:flex-row gap-4 pt-2">
              <Button 
                type="submit" 
                disabled={updatePasswordMutation.isPending || !newPassword || !confirmPassword}
              >
                {updatePasswordMutation.isPending ? 'Updating...' : 'Update Password'}
              </Button>
              {updatePasswordMutation.isSuccess && (
                <Button 
                  type="button" 
                  variant="outline" 
                  onClick={handleSecurityLogout}
                >
                  Log out for security
                </Button>
              )}
            </div>
          </form>
        </CardContent>
      </Card>

      {/* SECTION 4 - Danger Zone */}
      <Card className="border-destructive/50 mt-8">
        <CardHeader>
          <CardTitle className="text-destructive">Danger Zone</CardTitle>
          <CardDescription>Permanently delete your account and all saved data.</CardDescription>
        </CardHeader>
        <CardFooter>
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger className="inline-flex items-center justify-center rounded-md text-sm font-medium bg-destructive text-destructive-foreground hover:bg-destructive/90 h-10 px-4 py-2">
              Delete Account
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Are you absolutely sure?</DialogTitle>
                <DialogDescription>
                  This action cannot be undone. This will permanently delete your account,
                  including your favorites and watched movies list.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
                <Button variant="destructive" onClick={handleDeleteAccount} disabled={isDeleting}>
                  {isDeleting ? 'Deleting...' : 'Yes, delete my account'}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </CardFooter>
      </Card>
    </div>
  );
};
