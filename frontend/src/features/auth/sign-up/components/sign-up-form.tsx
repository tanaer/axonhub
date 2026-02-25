import { HTMLAttributes, useState } from 'react';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { passwordSchema } from '@/lib/validation';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { PasswordInput } from '@/components/password-input';
import { SocialLoginButtons } from '@/features/auth/components/social-login-buttons';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

type SignUpFormProps = HTMLAttributes<HTMLFormElement>;

const createFormSchema = (t: (key: string) => string) =>
  z
    .object({
      email: z.string().min(1, { message: t('auth.signUp.validation.emailRequired') }).email({ message: t('auth.signUp.validation.emailInvalid') }),
      password: passwordSchema(t),
      confirmPassword: z.string().min(1, { message: t('auth.signUp.validation.confirmPasswordRequired') }),
    })
    .refine((data) => data.password === data.confirmPassword, {
      message: t('auth.signUp.validation.passwordMismatch'),
      path: ['confirmPassword'],
    });

export function SignUpForm({ className, ...props }: SignUpFormProps) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  const formSchema = createFormSchema(t);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: '',
    },
  });

  async function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true);
    try {
      const response = await fetch('/admin/auth/signup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: data.email,
          password: data.password,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || '注册失败');
      }

      toast.success(t('auth.signUp.success') || '注册成功！请登录');
      // Redirect to sign-in
      window.location.href = '/sign-in';
    } catch (error: any) {
      toast.error(error.message || '注册失败');
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className={cn('grid gap-6', className)} {...props}>
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel className='text-sm font-medium text-slate-700'>{t('auth.signUp.form.email.label') || '邮箱'}</FormLabel>
              <FormControl>
                <Input
                  type='email'
                  placeholder={t('auth.signUp.form.email.placeholder') || 'name@example.com'}
                  className='border-slate-300 !bg-white text-slate-800 transition-all duration-300 placeholder:text-slate-400 focus:border-slate-500 focus:!bg-white focus:ring-2 focus:ring-slate-200'
                  {...field}
                />
              </FormControl>
              <FormMessage className='text-red-600' />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='password'
          render={({ field }) => (
            <FormItem>
              <FormLabel className='text-sm font-medium text-slate-700'>{t('auth.signUp.form.password.label') || '密码'}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('auth.signUp.form.password.placeholder') || '至少7个字符'}
                  className='border-slate-300 bg-white text-slate-800 backdrop-blur-sm transition-all duration-300 placeholder:text-slate-400 focus:border-slate-500 focus:bg-white focus:ring-2 focus:ring-slate-200'
                  {...field}
                />
              </FormControl>
              <FormMessage className='text-red-600' />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name='confirmPassword'
          render={({ field }) => (
            <FormItem>
              <FormLabel className='text-sm font-medium text-slate-700'>{t('auth.signUp.form.confirmPassword.label') || '确认密码'}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('auth.signUp.form.confirmPassword.placeholder') || '再次输入密码'}
                  className='border-slate-300 bg-white text-slate-800 backdrop-blur-sm transition-all duration-300 placeholder:text-slate-400 focus:border-slate-500 focus:bg-white focus:ring-2 focus:ring-slate-200'
                  {...field}
                />
              </FormControl>
              <FormMessage className='text-red-600' />
            </FormItem>
          )}
        />
        <Button
          type='submit'
          className='mt-6 w-full rounded-lg bg-slate-800 px-6 py-3 font-medium text-white shadow-lg transition-all duration-300 hover:bg-slate-700 hover:shadow-xl focus:ring-2 focus:ring-slate-500 focus:ring-offset-2 disabled:opacity-50'
          disabled={isLoading}
        >
          {isLoading ? (
            <div className='flex items-center justify-center gap-2'>
              <div className='h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white'></div>
              {t('auth.signUp.form.creating') || '注册中...'}
            </div>
          ) : (
            t('auth.signUp.form.submit') || '创建账号'
          )}
        </Button>

        {/* Social Login Buttons */}
        <SocialLoginButtons />

        {/* Sign In Link */}
        <div className="text-center text-sm text-slate-600">
          已有账号？{' '}
          <Link
            to='/sign-in'
            className='font-medium text-indigo-600 transition-colors hover:text-indigo-500 hover:underline'
          >
            立即登录
          </Link>
        </div>
      </form>
    </Form>
  );
}
