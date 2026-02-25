import { useEffect } from 'react';
import { useRouter } from '@tanstack/react-router';
import { isAuthError } from '@/gql/graphql';
import { useAuthStore, getTokenFromStorage } from '@/stores/authStore';
import { Skeleton } from '@/components/ui/skeleton';
import { useMe } from '@/features/auth/data/auth';

interface AuthGuardProps {
  children: React.ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const { accessToken } = useAuthStore((state) => state.auth);
  
  // Also check localStorage directly for initial load
  const storedToken = getTokenFromStorage();
  const hasToken = !!accessToken || !!storedToken;

  // Automatically fetch user info when token is available
  const { isLoading: isMeLoading, error: meError } = useMe();

  useEffect(() => {
    // If no token, redirect to sign-in
    if (!hasToken) {
      const currentPath = window.location.pathname;
      // Don't redirect if already on auth pages
      if (
        !currentPath.startsWith('/sign-in') &&
        !currentPath.startsWith('/sign-up') &&
        !currentPath.startsWith('/initialization') &&
        !currentPath.startsWith('/forgot-password') &&
        !currentPath.startsWith('/otp') &&
        !currentPath.startsWith('/auth/')
      ) {
        router.navigate({ to: '/sign-in' });
      }
    }
  }, [hasToken, router]);

  // Handle me query error (e.g., token expired)
  useEffect(() => {
    if (meError && hasToken && isAuthError(meError)) {
      // Token might be expired, redirect to sign-in
      router.navigate({ to: '/sign-in' });
    }
  }, [meError, hasToken, router]);

  // Show loading while checking auth
  if (!hasToken) {
    const currentPath = window.location.pathname;
    // Don't show loading on auth pages
    if (
      currentPath.startsWith('/sign-in') ||
      currentPath.startsWith('/sign-up') ||
      currentPath.startsWith('/initialization') ||
      currentPath.startsWith('/forgot-password') ||
      currentPath.startsWith('/otp') ||
      currentPath.startsWith('/auth/')
    ) {
      return <>{children}</>;
    }

    return (
      <div className='flex h-screen items-center justify-center'>
        <div className='space-y-4'>
          <Skeleton className='h-8 w-48' />
          <Skeleton className='h-4 w-32' />
        </div>
      </div>
    );
  }

  // Show loading while fetching user info
  if (hasToken && isMeLoading) {
    return (
      <div className='flex h-screen items-center justify-center'>
        <div className='space-y-4'>
          <Skeleton className='h-8 w-48' />
          <Skeleton className='h-4 w-32' />
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
