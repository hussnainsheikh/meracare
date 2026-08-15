import { createClient } from '@supabase/supabase-js';
import { AppState, Platform } from 'react-native';

import { env } from './env';
import { sessionStorage } from './secure-storage';

/**
 * Supabase client — authentication only.
 *
 * The mobile app must not perform business-critical database writes through
 * Supabase's data APIs; those go through the Go API (docs/07-database-and-sync.md).
 */
export const supabase = createClient(env.supabaseUrl, env.supabaseAnonKey, {
  auth: {
    storage: sessionStorage,
    autoRefreshToken: true,
    persistSession: true,
    // There is no URL to parse on native; deep-link sessions are handled
    // explicitly when social sign-in is added.
    detectSessionInUrl: Platform.OS === 'web',
  },
});

/**
 * Refreshes the session while the app is in the foreground and stops the timer
 * in the background — no continuous JS timer runs while backgrounded
 * (docs/08-notifications-and-background.md).
 */
export function startAuthAutoRefresh(): () => void {
  if (Platform.OS === 'web') return () => {};

  const apply = (state: string) => {
    if (state === 'active') {
      void supabase.auth.startAutoRefresh();
    } else {
      void supabase.auth.stopAutoRefresh();
    }
  };

  apply(AppState.currentState);
  const subscription = AppState.addEventListener('change', apply);

  return () => {
    subscription.remove();
    void supabase.auth.stopAutoRefresh();
  };
}
