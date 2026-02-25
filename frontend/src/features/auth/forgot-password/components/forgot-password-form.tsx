import { HTMLAttributes, useState } from 'react';
import { z } from 'zod';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { Link } from '@tanstack/react-router';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

type ForgotFormProps = HTMLAttributes<HTMLFormElement>;

const createFormSchema = (t: (key: string) => string) =>
  z.object({
    email: z.string().min(1, { message: t('auth.forgotPassword.validation.emailRequired') }).email({ message: t('auth.forgotPassword.validation.emailInvalid') }),
  });

export function ForgotPasswordForm({ className, ...props }: ForgotFormProps) {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const formSchema = createFormSchema(t);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: { email: '' },
  });

  async function onSubmit(data: z.infer<typeof formSchema>) {
    setIsLoading(true);
    try {
      const response = await fetch('/admin/auth/forgot-password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email: data.email }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || '发送失败');
      }

      setIsSuccess(true);
      toast.success(t('auth.forgotPassword.success') || '重置链接已发送到您的邮箱');
    } catch (error: any) {
      toast.error(error.message || '发送失败');
    } finally {
      setIsLoading(false);
    }
  }

  if (isSuccess) {
    return (
      <div className='text-center space-y-4'>
        <div className='text-green-600 text-5xl mb-4'>✓</div>
        <p className='text-slate-700'>{t('auth.forgotPassword.checkEmail') || '请检查您的邮箱，按照邮件中的说明重置密码'}</p>
        <Link
          to='/sign-in'
          className='inline-block font-medium text-indigo-600 transition-colors hover:text-indigo-500 hover:underline'
        >
          {t('auth.forgotPassword.backToSignIn') || '返回登录'}
        </Link>
      </div>
    );
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className={cn('grid gap-6', className)} {...props}>
        <FormField
          control={form.control}
          name='email'
          render={({ field }) => (
            <FormItem>
              <FormLabel className='text-sm font-medium text-slate-700'>{t('auth.forgotPassword.form.email.label') || '邮箱地址'}</FormLabel>
              <FormControl>
                <Input
                  type='email'
                  placeholder={t('auth.forgotPassword.form.email.placeholder') || 'name@example.com'}
                  className='border-slate-300 !bg-white text-slate-800 transition-all duration-300 placeholder:text-slate-400 focus:border-slate-500 focus:!bg-white focus:ring-2 focus:ring-slate-200'
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
              {t('auth.forgotPassword.form.sending') || '发送中...'}
            </div>
          ) : (
            t('auth.forgotPassword.form.submit') || '发送重置链接'
          )}
        </Button>

        {/* Back to Sign In Link */}
        <div className="text-center text-sm text-slate-600">
          <Link
            to='/sign-in'
            className='font-medium text-indigo-600 transition-colors hover:text-indigo-500 hover:underline'
          >
            {t('auth.forgotPassword.backToSignIn') || '返回登录'}
          </Link>
        </div>
      </form>
    </Form>
  );
}
