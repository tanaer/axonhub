import { createFileRoute } from '@tanstack/react-router';
import { ModelsPage } from '@/features/user-center/pages';

export const Route = createFileRoute('/_authenticated/user-center/models')({
  component: ModelsPage,
});
