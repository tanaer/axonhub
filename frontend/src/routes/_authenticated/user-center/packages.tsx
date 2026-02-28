import { createFileRoute } from '@tanstack/react-router';
import { PackagesPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/packages')({
  component: PackagesPage,
});
