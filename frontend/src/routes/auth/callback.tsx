import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';

export const Route = createFileRoute('/auth/callback')({
  component: AuthCallback,
});

function AuthCallback() {
  const navigate = useNavigate();
  const { setUser, setAccessToken } = useAuthStore((state) => state.auth);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for auth success flag
        const urlParams = new URLSearchParams(window.location.search);
        const authSuccess = urlParams.get('auth');

        if (authSuccess === 'success') {
          // Fetch user info from verify endpoint (uses cookie)
          const response = await fetch('/auth/verify', {
            credentials: 'include',
          });

          if (response.ok) {
            const userData = await response.json();

            // Store user info
            const user = {
              id: userData.user_id,
              email: userData.email,
              firstName: userData.name?.split(' ')[0] || '',
              lastName: userData.name?.split(' ').slice(1).join(' ') || '',
              isOwner: false,
              preferLanguage: 'zh',
              avatar: userData.avatar,
              scopes: [],
              roles: [],
              projects: [],
            };

            setUser(user);

            // Get token from cookie and store in localStorage
            const token = getCookie('auth_token');
            if (token) {
              setAccessToken(token);
            }

            // Redirect to dashboard
            navigate({ to: '/dashboard' });
          } else {
            setError('验证登录状态失败');
          }
        } else {
          setError('登录失败');
        }
      } catch (err) {
        setError('处理登录回调时出错');
        console.error('Auth callback error:', err);
      }
    };

    handleCallback();
  }, [navigate, setUser, setAccessToken]);

  if (error) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <div className="text-center">
          <div className="text-red-500 mb-4">{error}</div>
          <button
            onClick={() => navigate({ to: '/sign-in' })}
            className="text-indigo-600 hover:underline"
          >
            返回登录
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="text-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-indigo-600 border-t-transparent mx-auto mb-4"></div>
        <div className="text-slate-600">正在处理登录...</div>
      </div>
    </div>
  );
}

function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift() || null;
  }
  return null;
}
