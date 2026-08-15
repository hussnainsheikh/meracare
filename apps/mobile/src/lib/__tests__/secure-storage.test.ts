import * as SecureStore from 'expo-secure-store';

import { secureStorage } from '../secure-storage';

jest.mock('expo-secure-store');

// In-memory stand-in for the Keychain/Keystore.
const store = new Map<string, string>();

beforeEach(() => {
  store.clear();
  jest.mocked(SecureStore.getItemAsync).mockImplementation(async (key) => store.get(key) ?? null);
  jest.mocked(SecureStore.setItemAsync).mockImplementation(async (key, value) => {
    store.set(key, value);
  });
  jest.mocked(SecureStore.deleteItemAsync).mockImplementation(async (key) => {
    store.delete(key);
  });
});

describe('secureStorage', () => {
  it('returns null for a key that was never written', async () => {
    await expect(secureStorage.getItem('session')).resolves.toBeNull();
  });

  it('round-trips a short value', async () => {
    await secureStorage.setItem('session', 'a-short-token');

    await expect(secureStorage.getItem('session')).resolves.toBe('a-short-token');
  });

  // A Supabase session with a long JWT exceeds SecureStore's ~2KB advisory
  // limit, so long values must be chunked.
  it('round-trips a value larger than one chunk', async () => {
    const value = 'x'.repeat(5000);

    await secureStorage.setItem('session', value);

    expect(store.size).toBeGreaterThan(2);
    await expect(secureStorage.getItem('session')).resolves.toBe(value);
  });

  it('drops chunks left over when a value shrinks', async () => {
    await secureStorage.setItem('session', 'y'.repeat(5000));
    await secureStorage.setItem('session', 'small');

    await expect(secureStorage.getItem('session')).resolves.toBe('small');
    expect([...store.keys()].filter((key) => key.startsWith('session.'))).toHaveLength(2);
  });

  it('removes every chunk', async () => {
    await secureStorage.setItem('session', 'z'.repeat(5000));

    await secureStorage.removeItem('session');

    expect(store.size).toBe(0);
    await expect(secureStorage.getItem('session')).resolves.toBeNull();
  });

  // A half-written value must not be handed back as a usable session.
  it('discards a partially written value', async () => {
    await secureStorage.setItem('session', 'w'.repeat(5000));
    store.delete('session.1');

    await expect(secureStorage.getItem('session')).resolves.toBeNull();
    expect(store.size).toBe(0);
  });

  it('round-trips an empty string as an empty string, not null', async () => {
    await secureStorage.setItem('session', '');

    await expect(secureStorage.getItem('session')).resolves.toBe('');
  });
});
