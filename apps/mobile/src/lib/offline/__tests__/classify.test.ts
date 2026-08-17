import { ApiError } from '@/lib/api-error';

import { classify, readNotes } from '../classify';

/**
 * Which failures are worth retrying.
 *
 * This is the decision the offline queue turns on. Treating a refusal as
 * temporary would retry it forever; treating a lost connection as final would
 * throw away care that really happened.
 */

describe('classify', () => {
  it('retries a request that never reached the server', () => {
    expect(classify(ApiError.network(new Error('offline')))).toMatchObject({ kind: 'transient' });
  });

  it('retries a server-side failure', () => {
    const outcome = classify(new ApiError(503, 'UNAVAILABLE', 'The service is not ready.'));

    expect(outcome.kind).toBe('transient');
  });

  // The task already has a different outcome and the server is authoritative.
  // Sending the same action again could only be refused again.
  it('gives up when the record has already been decided', () => {
    const outcome = classify(
      new ApiError(409, 'CONFLICT', 'Somebody has already recorded a different outcome.'),
    );

    expect(outcome).toMatchObject({
      kind: 'permanent',
      message: 'Somebody has already recorded a different outcome.',
    });
  });

  // The task is gone, or access was revoked while the action sat in the queue.
  it('gives up when the record is no longer reachable', () => {
    expect(classify(new ApiError(404, 'NOT_FOUND', 'That task does not exist.'))).toMatchObject({
      kind: 'permanent',
    });
  });

  it('gives up when the caller is not allowed', () => {
    expect(classify(new ApiError(403, 'FORBIDDEN', 'Not allowed.'))).toMatchObject({
      kind: 'permanent',
    });
  });

  it('gives up on a request the server will always reject', () => {
    expect(
      classify(new ApiError(422, 'VALIDATION_FAILED', 'Please check the fields.')),
    ).toMatchObject({ kind: 'permanent' });
  });

  // An expired session is not fixed by resending the same request unchanged;
  // the user has to sign in, and the queue should not spin meanwhile.
  it('gives up when the session has expired', () => {
    expect(classify(new ApiError(401, 'UNAUTHENTICATED', 'Sign in to continue.'))).toMatchObject({
      kind: 'permanent',
    });
  });

  // Something unexpected is worth one more attempt rather than discarding a
  // record of care on a failure we do not understand.
  it('retries an error it does not recognise', () => {
    expect(classify(new Error('something odd'))).toMatchObject({ kind: 'transient' });
  });
});

describe('readNotes', () => {
  it('reads the note a queued operation carries', () => {
    expect(readNotes(JSON.stringify({ notes: 'She was asleep' }))).toBe('She was asleep');
  });

  it('has nothing to read when no note was recorded', () => {
    expect(readNotes(null)).toBeNull();
    expect(readNotes(JSON.stringify({}))).toBeNull();
  });

  // A payload that will not parse must not take the whole pass down with it:
  // the operation still deserves to be sent, just without a note.
  it('survives a payload it cannot read', () => {
    expect(readNotes('not json')).toBeNull();
  });
});
