/**
 * Jest setup.
 *
 * Environment variables are set here because `src/lib/env.ts` throws when they
 * are missing — that check is deliberate, so tests supply values rather than
 * weaken it.
 */
process.env.EXPO_PUBLIC_SUPABASE_URL ??= 'https://test.supabase.co';
process.env.EXPO_PUBLIC_SUPABASE_ANON_KEY ??= 'test-anon-key';
process.env.EXPO_PUBLIC_API_URL ??= 'http://localhost:8080';
