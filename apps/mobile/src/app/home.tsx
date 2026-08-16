import type { Senior } from '@meracare/contracts';
import { Link, Redirect, router } from 'expo-router';
import { ActivityIndicator, Pressable, StyleSheet, View } from 'react-native';

import { Button, Card, Screen, Text } from '@/components/ui';
import { useSession } from '@/features/auth/session-provider';
import { useAuthActions } from '@/features/auth/use-auth-actions';
import { useSeniors } from '@/features/seniors/use-seniors';
import { ApiError } from '@/lib/api-error';
import { useTheme } from '@/theme';

/**
 * Home / Today.
 *
 * One screen for every care mode: what the user sees is decided by their
 * relationships, not by a separate app (docs/13-mvp-screen-map.md). The daily
 * task, medication and appointment content arrives in later phases.
 */
export default function HomeScreen() {
  const theme = useTheme();
  const { isSignedIn, isRestoring } = useSession();
  const { signOut, isSubmitting } = useAuthActions();
  const seniors = useSeniors(isSignedIn);

  if (!isRestoring && !isSignedIn) {
    return <Redirect href="/sign-in" />;
  }

  return (
    <Screen scrollable>
      <Text variant="pageHeading">Today</Text>

      {seniors.isPending ? (
        <View
          style={{
            alignItems: 'center',
            gap: theme.spacing.md,
            paddingVertical: theme.spacing.xxl,
          }}
        >
          <ActivityIndicator color={theme.colors.primary} />
          <Text variant="secondary" color="secondary">
            Loading your care circle…
          </Text>
        </View>
      ) : seniors.isError ? (
        <Card>
          <Text variant="bodyStrong">We could not load your seniors</Text>
          <Text variant="secondary" color="secondary">
            {seniors.error instanceof ApiError
              ? seniors.error.message
              : 'Something went wrong. Please try again.'}
          </Text>
          <Button variant="secondary" label="Try again" onPress={() => void seniors.refetch()} />
        </Card>
      ) : seniors.data.length === 0 ? (
        <EmptyState />
      ) : (
        <View style={{ gap: theme.spacing.md }}>
          {seniors.data.map((senior) => (
            <SeniorRow key={senior.id} senior={senior} />
          ))}

          <Button
            variant="secondary"
            label="Add another person"
            onPress={() => router.push('/onboarding')}
          />
        </View>
      )}

      <Button variant="ghost" label="Sign out" onPress={signOut} loading={isSubmitting} />
    </Screen>
  );
}

/** Shown before the first profile exists — Solo Mode is one tap away. */
function EmptyState() {
  const theme = useTheme();

  return (
    <Card>
      <Text variant="sectionHeading">Let&apos;s get set up</Text>
      <Text variant="body" color="secondary">
        Add the person you are caring for — or yourself. You can invite family and caregivers
        whenever you need help, and never before.
      </Text>
      <View style={{ marginTop: theme.spacing.sm }}>
        <Button label="Get started" onPress={() => router.push('/onboarding')} />
      </View>
    </Card>
  );
}

/** One senior in the list, labelled with the reader's own relationship. */
function SeniorRow({ senior }: { senior: Senior }) {
  const theme = useTheme();

  return (
    <Link href={{ pathname: '/seniors/[seniorId]', params: { seniorId: senior.id } }} asChild>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`${senior.displayName}, ${describeRole(senior)}`}
        style={({ pressed }) => [
          styles.row,
          {
            minHeight: theme.minTouchTarget,
            padding: theme.spacing.lg,
            borderRadius: theme.radii.lg,
            backgroundColor: theme.colors.surface,
            borderColor: theme.colors.border,
            opacity: pressed ? 0.85 : 1,
          },
        ]}
      >
        <View style={{ flex: 1, gap: theme.spacing.xs }}>
          <Text variant="bodyStrong">{senior.displayName}</Text>
          <Text variant="secondary" color="secondary">
            {describeRole(senior)}
          </Text>
        </View>
        <Text variant="body" color="secondary">
          ›
        </Text>
      </Pressable>
    </Link>
  );
}

/** Plain-language description of the reader's relationship to this senior. */
function describeRole(senior: Senior): string {
  if (senior.isSelf) return 'Your own care';

  switch (senior.role) {
    case 'family_member':
      return 'Family care';
    case 'professional_caregiver':
      return 'In your professional care';
    default:
      return 'Care circle';
  }
}

const styles = StyleSheet.create({
  row: {
    alignItems: 'center',
    borderWidth: StyleSheet.hairlineWidth,
    flexDirection: 'row',
    gap: 12,
  },
});
