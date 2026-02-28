import { createFileRoute } from '@tanstack/react-router';
import { ReferralPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/referral')({
  component: ReferralPage,
});
