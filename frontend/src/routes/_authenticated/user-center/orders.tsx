import { createFileRoute } from '@tanstack/react-router';
import { OrdersPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/orders')({
  component: OrdersPage,
});
