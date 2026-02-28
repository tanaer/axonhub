import { createFileRoute } from '@tanstack/react-router';
import { APIKeysPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/api-keys')({
  component: APIKeysPage,
});
