import { useTranslation } from 'react-i18next';
import AuthLayout from '../auth-layout';
import TwoColumnAuth from '../components/two-column-auth';
import AnimatedLineBackground from '../sign-in/components/animated-line-background';
import { ForgotPasswordForm } from './components/forgot-password-form';
import '../sign-in/login-styles.css';

export default function ForgotPassword() {
  const { t } = useTranslation();

  return (
    <AuthLayout>
      <AnimatedLineBackground key='optimized-layout' />
      <TwoColumnAuth
        title={t('auth.forgotPassword.title') || '忘记密码'}
        description={t('auth.forgotPassword.subtitle') || '输入您的注册邮箱，我们将发送重置密码链接'}
        rightFooter={null}
      >
        <ForgotPasswordForm />
      </TwoColumnAuth>
    </AuthLayout>
  );
}
