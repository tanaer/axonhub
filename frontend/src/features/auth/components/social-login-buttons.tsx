import { Button } from '@/components/ui/button';
import { IconBrandGoogle, IconBrandGithub } from '@tabler/icons-react';

interface SocialLoginButtonsProps {
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export function SocialLoginButtons({ onError }: SocialLoginButtonsProps) {
  const handleSocialLogin = (provider: 'google' | 'github') => {
    const redirectUri = encodeURIComponent(window.location.origin + '/auth/callback');
    window.location.href = `/api/auth/oauth/${provider}?redirect_uri=${redirectUri}`;
  };

  return (
    <div className="space-y-3">
      <div className="relative">
        <div className="absolute inset-0 flex items-center">
          <span className="w-full border-t border-slate-300" />
        </div>
        <div className="relative flex justify-center text-xs uppercase">
          <span className="bg-white px-2 text-slate-500">或使用社交账号登录</span>
        </div>
      </div>
      
      <div className="grid grid-cols-2 gap-3">
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={() => handleSocialLogin('google')}
        >
          <IconBrandGoogle className="mr-2 h-4 w-4" />
          Google
        </Button>
        <Button
          type="button"
          variant="outline"
          className="w-full"
          onClick={() => handleSocialLogin('github')}
        >
          <IconBrandGithub className="mr-2 h-4 w-4" />
          GitHub
        </Button>
      </div>
    </div>
  );
}
