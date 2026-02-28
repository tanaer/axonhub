import { createFileRoute } from '@tanstack/react-router';
import { BillingPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/billing')({
  component: BillingPage,
});
