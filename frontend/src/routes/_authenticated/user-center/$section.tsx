import { createFileRoute } from '@tanstack/react-router';
import UserCenter from '@/features/user-center';

export const Route = createFileRoute('/_authenticated/user-center/$section')({
  component: UserCenter,
});
