import { createFileRoute } from '@tanstack/react-router';
import { AccountPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/account')({
  component: AccountPage,
});
