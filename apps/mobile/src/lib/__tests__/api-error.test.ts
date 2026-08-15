import { ApiError } from '../api-error';

describe('ApiError.fromResponse', () => {
  it('reads the API error envelope', () => {
    const error = ApiError.fromResponse(403, {
      error: { code: 'FORBIDDEN', message: 'You do not have access to this senior.' },
    });

    expect(error.status).toBe(403);
    expect(error.code).toBe('FORBIDDEN');
    expect(error.message).toBe('You do not have access to this senior.');
    expect(error.isForbidden).toBe(true);
    expect(error.isRetryable).toBe(false);
  });

  it('keeps validation details', () => {
    const error = ApiError.fromResponse(422, {
      error: {
        code: 'VALIDATION_FAILED',
        message: 'Please check the highlighted fields.',
        details: { displayName: 'This field is required.' },
      },
    });

    expect(error.details).toEqual({ displayName: 'This field is required.' });
  });

  it('falls back to a safe message for an unrecognised body', () => {
    for (const body of [null, undefined, '<html>502 Bad Gateway</html>', { oops: true }]) {
      const error = ApiError.fromResponse(502, body);

      expect(error.code).toBe('INTERNAL');
      expect(error.message).toBe('Something went wrong. Please try again.');
      expect(error.isRetryable).toBe(true);
    }
  });

  it('flags 401 as unauthenticated', () => {
    const error = ApiError.fromResponse(401, {
      error: { code: 'UNAUTHENTICATED', message: 'Sign in to continue.' },
    });

    expect(error.isUnauthenticated).toBe(true);
    expect(error.isRetryable).toBe(false);
  });
});

describe('ApiError.network', () => {
  it('is retryable and explains the offline state in plain language', () => {
    const cause = new TypeError('Network request failed');
    const error = ApiError.network(cause);

    expect(error.status).toBe(0);
    expect(error.isRetryable).toBe(true);
    expect(error.cause).toBe(cause);
    expect(error.message).toContain('offline');
  });
});
