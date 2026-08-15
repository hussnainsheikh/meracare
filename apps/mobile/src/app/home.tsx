import { Redirect } from 'expo-router';
import { View } from 'react-native';

import { Button, Card, Screen, Text } from '@/components/ui';
import { useSession } from '@/features/auth/session-provider';
import { useAuthActions } from '@/features/auth/use-auth-actions';
import { useMe } from '@/features/profile/use-me';
import { ApiError } from '@/lib/api-error';
import { useTheme } from '@/theme';

/**
 * Placeholder home screen.
 *
 * It exists to prove the foundation end to end: Supabase session → Go API →
 * PostgreSQL. The real Today/Home screens arrive with Phase 2 and Phase 9.
 */
export default function HomeScreen() {
  const theme = useTheme();
  const { isSignedIn, isRestoring } = useSession();
  const { signOut, isSubmitting } = useAuthActions();
  const me = useMe(isSignedIn);

  if (!isRestoring && !isSignedIn) {
    return <Redirect href="/sign-in" />;
  }

  return (
    <Screen scrollable>
      <Text variant="pageHeading">Today</Text>

      <Card>
        <Text variant="sectionHeading">Your account</Text>

        {me.isPending ? (
          <Text variant="body" color="secondary">
            Loading your profile…
          </Text>
        ) : me.isError ? (
          <View style={{ gap: theme.spacing.sm }}>
            <Text variant="body" color="danger">
              {me.error instanceof ApiError
                ? me.error.message
                : 'We could not load your profile just now.'}
            </Text>
            <Button variant="secondary" label="Try again" onPress={() => void me.refetch()} />
          </View>
        ) : (
          <View style={{ gap: theme.spacing.xs }}>
            <Text variant="body">{me.data.displayName}</Text>
            <Text variant="secondary" color="secondary">
              Connected to the MeraCare API.
            </Text>
          </View>
        )}
      </Card>

      <Button variant="ghost" label="Sign out" onPress={signOut} loading={isSubmitting} />
    </Screen>
  );
}
