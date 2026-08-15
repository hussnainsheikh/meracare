import { useState } from 'react';

import { supabase } from '@/lib/supabase';

/**
 * Email sign-in, sign-up and sign-out.
 *
 * Apple and Google are the documented launch providers (docs/12-tech-stack.md);
 * they are added once the Supabase project has those providers configured.
 * Email keeps the foundation verifiable end to end in the meantime.
 */
export function useAuthActions() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function run(action: () => Promise<{ error: { message: string } | null }>) {
    setIsSubmitting(true);
    setError(null);
    try {
      const result = await action();
      if (result.error) {
        setError(result.error.message);
        return false;
      }
      return true;
    } finally {
      setIsSubmitting(false);
    }
  }

  return {
    isSubmitting,
    error,
    clearError: () => setError(null),

    signIn: (email: string, password: string) =>
      run(() => supabase.auth.signInWithPassword({ email: email.trim(), password })),

    signUp: (email: string, password: string) =>
      run(() => supabase.auth.signUp({ email: email.trim(), password })),

    signOut: () => run(() => supabase.auth.signOut()),
  };
}
