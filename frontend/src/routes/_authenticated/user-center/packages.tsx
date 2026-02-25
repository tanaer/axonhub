import { createFileRoute } from '@tanstack/react-router';
import UserCenter from '@/features/user-center';

export const Route = createFileRoute('/_authenticated/user-center/packages')({
  component: UserCenter,
});
