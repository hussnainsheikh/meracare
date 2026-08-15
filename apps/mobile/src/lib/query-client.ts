import { QueryClient } from '@tanstack/react-query';

import { ApiError } from './api-error';

/**
 * TanStack Query is the only cache for server state (docs/06-mobile-architecture.md).
 *
 * Defaults are tuned for care data on a phone: no polling, retry only the
 * failures that are worth retrying, and keep data usable while it refreshes.
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        gcTime: 24 * 60 * 60 * 1000,
        refetchOnWindowFocus: false,
        // No polling — the server pushes what matters (docs/08).
        refetchInterval: false,
        retry: (failureCount, error) => {
          if (error instanceof ApiError && !error.isRetryable) return false;
          return failureCount < 2;
        },
      },
      mutations: {
        // Mutations are retried deliberately by the sync queue in Phase 11,
        // not blindly here.
        retry: false,
      },
    },
  });
}
