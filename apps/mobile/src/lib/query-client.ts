import { QueryCache, QueryClient } from '@tanstack/react-query';

import { ApiError } from './api-error';

/**
 * TanStack Query is the only cache for server state (docs/06-mobile-architecture.md).
 *
 * Defaults are tuned for care data on a phone: no polling, retry only the
 * failures that are worth retrying, and keep data usable while it refreshes.
 */
export function createQueryClient(): QueryClient {
  // Losing access to a senior is not an error the screen that noticed should
  // handle alone. A caregiver revoked while the app is open keeps whatever was
  // already fetched — their name in the list on Today, yesterday's medication
  // — until something asks again. This watches every query for the API saying
  // no, and drops that senior from the cache when it does
  // (plans/phase9.md §14).
  const queryCache = new QueryCache({
    onError: (error, query) => {
      if (!(error instanceof ApiError)) return;
      // 404 is what a denial looks like: the API answers the same for a senior
      // that does not exist and one the caller may not see, deliberately
      // (docs/02-permissions-and-authorization.md). 401 is an expired session.
      if (error.status !== 404 && error.status !== 401) return;

      const [root, seniorId] = query.queryKey;
      if (root !== 'seniors' || typeof seniorId !== 'string') return;

      // Only what nothing is currently watching. A mounted screen keeps its own
      // query so it can render its own "not available" message; removing it
      // would make its observer refetch, fail again, and arrive back here —
      // a loop that would hammer the API for as long as the screen stayed open.
      client.removeQueries({ queryKey: ['seniors', seniorId], type: 'inactive' });

      // The list is now wrong too, and refetching it is safe: it answers 200
      // with the senior simply absent, so it cannot re-enter this handler.
      void client.invalidateQueries({ queryKey: ['seniors'], exact: true });
    },
  });

  const client = new QueryClient({
    queryCache,
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
        // Mutations are retried deliberately by the offline sync queue, not
        // blindly here.
        retry: false,
      },
    },
  });

  return client;
}
