import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/(auth)/auth/callback')({
  component: AuthCallback,
});

function AuthCallback() {
  const { isLoading, error } = (function() {
    // Simple OAuth callback handler
    // In production, this should properly exchange the code for tokens
    const searchParams = new URLSearchParams(window.location.search);
    const code = searchParams.get('code');
    const state = searchParams.get('state');
    const errorParam = searchParams.get('error');
    
    if (errorParam) {
      return { isLoading: false, error: errorParam };
    }
    
    if (code) {
      // TODO: Exchange code for tokens via backend API
      // For now, just redirect to dashboard
      setTimeout(() => {
        window.location.href = '/';
      }, 1000);
      return { isLoading: true, error: null };
    }
    
    return { isLoading: false, error: 'Invalid callback' };
  })();

  if (isLoading) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="text-center">
          <div className="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent" />
          <p className="mt-4 text-gray-600">正在完成登录...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-screen w-full items-center justify-center">
        <div className="text-center text-red-600">
          <p className="text-xl font-semibold">登录失败</p>
          <p className="mt-2 text-gray-600">{error}</p>
          <button 
            onClick={() => window.location.href = '/sign-in'}
            className="mt-4 text-indigo-600 hover:underline"
          >
            返回登录
          </button>
        </div>
      </div>
    );
  }

  return null;
}
