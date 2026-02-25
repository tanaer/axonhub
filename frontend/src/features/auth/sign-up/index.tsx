import { useTranslation } from 'react-i18next';
import AuthLayout from '../auth-layout';
import TwoColumnAuth from '../components/two-column-auth';
import AnimatedLineBackground from '../sign-in/components/animated-line-background';
import { SignUpForm } from './components/sign-up-form';
import '../sign-in/login-styles.css';

export default function SignUp() {
  const { t } = useTranslation();

  return (
    <AuthLayout>
      <AnimatedLineBackground key='optimized-layout' />
      <TwoColumnAuth
        title={t('auth.signUp.title') || '创建账号'}
        description={t('auth.signUp.subtitle') || '填写以下信息注册新账号'}
        rightFooter={<p className='text-xs leading-relaxed text-slate-500 sm:text-sm'>{t('auth.signUp.footer.agreement') || '注册即表示您同意我们的服务条款和隐私政策'}</p>}
      >
        <SignUpForm />
      </TwoColumnAuth>
    </AuthLayout>
  );
}
