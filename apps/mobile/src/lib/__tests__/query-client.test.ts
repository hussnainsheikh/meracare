import { ApiError } from '@/lib/api-error';
import { createQueryClient } from '@/lib/query-client';

/**
 * What happens to cached care data when access to a senior goes away while the
 * app is open (plans/phase9.md §14).
 *
 * Revocation happens on somebody else's phone. This one finds out only when it
 * next asks, and until then it is holding a family's medication list.
 */

function seed(client: ReturnType<typeof createQueryClient>) {
  client.setQueryData(['seniors'], [{ id: 'senior-1' }, { id: 'senior-2' }]);
  client.setQueryData(['seniors', 'senior-1'], { id: 'senior-1', displayName: 'Amma' });
  client.setQueryData(['seniors', 'senior-1', 'medications', 'today'], [{ id: 'dose-1' }]);
  client.setQueryData(['seniors', 'senior-2', 'medications', 'today'], [{ id: 'dose-2' }]);
}

/** Runs a query that fails, so the cache's error handler sees it. */
async function failWith(
  client: ReturnType<typeof createQueryClient>,
  queryKey: unknown[],
  error: unknown,
) {
  await client
    .fetchQuery({ queryKey, queryFn: () => Promise.reject(error), retry: false })
    .catch(() => {});
}

it('drops a senior’s cached care when the API stops recognising the caller', async () => {
  const client = createQueryClient();
  seed(client);

  await failWith(
    client,
    ['seniors', 'senior-1', 'tasks', 'today'],
    new ApiError(404, 'NOT_FOUND', 'That senior does not exist.'),
  );

  expect(client.getQueryData(['seniors', 'senior-1', 'medications', 'today'])).toBeUndefined();
  expect(client.getQueryData(['seniors', 'senior-1'])).toBeUndefined();

  client.clear();
});

it('leaves every other senior alone', async () => {
  // A caregiver with several clients loses one, not all of them.
  const client = createQueryClient();
  seed(client);

  await failWith(
    client,
    ['seniors', 'senior-1', 'tasks', 'today'],
    new ApiError(404, 'NOT_FOUND', 'That senior does not exist.'),
  );

  expect(client.getQueryData(['seniors', 'senior-2', 'medications', 'today'])).toEqual([
    { id: 'dose-2' },
  ]);

  client.clear();
});

it('does the same for an expired session', async () => {
  const client = createQueryClient();
  seed(client);

  await failWith(
    client,
    ['seniors', 'senior-1', 'activity'],
    new ApiError(401, 'UNAUTHENTICATED', 'Sign in to continue.'),
  );

  expect(client.getQueryData(['seniors', 'senior-1'])).toBeUndefined();

  client.clear();
});

it('ignores failures that say nothing about access', async () => {
  // A server error or a dropped connection is not a revocation, and throwing
  // the cache away for one would empty the app every time the signal drops.
  const client = createQueryClient();

  for (const error of [
    new ApiError(500, 'INTERNAL', 'Something went wrong. Please try again.'),
    ApiError.network(new Error('offline')),
    new Error('not an ApiError at all'),
  ]) {
    seed(client);
    await failWith(client, ['seniors', 'senior-1', 'tasks', 'today'], error);

    expect(client.getQueryData(['seniors', 'senior-1'])).toEqual({
      id: 'senior-1',
      displayName: 'Amma',
    });
  }

  client.clear();
});

it('ignores failures that are not about a senior at all', async () => {
  const client = createQueryClient();
  seed(client);

  await failWith(
    client,
    ['medications', 'med-1'],
    new ApiError(404, 'NOT_FOUND', 'That medication does not exist.'),
  );

  expect(client.getQueryData(['seniors', 'senior-1'])).toEqual({
    id: 'senior-1',
    displayName: 'Amma',
  });

  client.clear();
});
