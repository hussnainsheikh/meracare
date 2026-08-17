import { can } from '@meracare/contracts';
import { Stack, router, useLocalSearchParams } from 'expo-router';
import { ActivityIndicator, View } from 'react-native';

import { Button, Card, Screen, Text } from '@/components/ui';
import { useSenior } from '@/features/seniors/use-seniors';
import { ApiError } from '@/lib/api-error';
import { useTheme } from '@/theme';

/**
 * Senior dashboard.
 *
 * The same screen serves every care mode; the caller's permissions decide which
 * actions appear. Tasks, medications, appointments and activity arrive in
 * Phases 4 to 7 — this is the shell they will hang from.
 */
export default function SeniorDashboardScreen() {
  const theme = useTheme();
  const { seniorId } = useLocalSearchParams<{ seniorId: string }>();
  const senior = useSenior(seniorId ?? null);

  if (senior.isPending) {
    return (
      <Screen>
        <View style={{ alignItems: 'center', flex: 1, justifyContent: 'center' }}>
          <ActivityIndicator color={theme.colors.primary} />
        </View>
      </Screen>
    );
  }

  if (senior.isError) {
    const notFound = senior.error instanceof ApiError && senior.error.status === 404;
    return (
      <Screen scrollable>
        <Stack.Screen options={{ headerShown: true, title: '' }} />
        <Card>
          <Text variant="sectionHeading">
            {notFound ? 'This profile is not available' : 'We could not load this profile'}
          </Text>
          <Text variant="body" color="secondary">
            {senior.error instanceof ApiError
              ? senior.error.message
              : 'Something went wrong. Please try again.'}
          </Text>
          <Button
            variant="secondary"
            label="Back to Today"
            onPress={() => router.replace('/home')}
          />
        </Card>
      </Screen>
    );
  }

  const profile = senior.data;

  return (
    <Screen scrollable>
      <Stack.Screen options={{ headerShown: true, title: profile.displayName }} />

      <View style={{ gap: theme.spacing.sm }}>
        <Text variant="pageHeading">{profile.displayName}</Text>
        <Text variant="body" color="secondary">
          {profile.isSelf ? 'Your own care' : 'Care circle'}
        </Text>
      </View>

      <Card>
        <Text variant="sectionHeading">Profile</Text>
        <DetailRow label="Date of birth" value={profile.dateOfBirth} />
        <DetailRow label="Phone" value={profile.phone} />
        <DetailRow label="Address" value={profile.address} />
        <DetailRow label="Emergency contact" value={profile.emergencyContact} />

        {/* The edit action appears only when the API would allow it. */}
        {can(profile, 'senior.edit') ? (
          <Button
            variant="secondary"
            label="Edit profile"
            onPress={() =>
              router.push({
                pathname: '/seniors/[seniorId]/edit',
                params: { seniorId: profile.id },
              })
            }
          />
        ) : null}
      </Card>

      {can(profile, 'members.view') ? (
        <Card>
          <Text variant="sectionHeading">Care circle</Text>
          <Text variant="body" color="secondary">
            The family and caregivers involved in this person&apos;s care.
          </Text>
          <Button
            variant="secondary"
            label="View care circle"
            onPress={() =>
              router.push({
                pathname: '/seniors/[seniorId]/circle',
                params: { seniorId: profile.id },
              })
            }
          />
        </Card>
      ) : null}

      {can(profile, 'tasks.view') ? (
        <Card>
          <Text variant="sectionHeading">Care tasks</Text>
          <Text variant="body" color="secondary">
            What needs doing today, and what is coming up.
          </Text>
          <Button
            label="View care tasks"
            onPress={() =>
              router.push({
                pathname: '/seniors/[seniorId]/tasks',
                params: { seniorId: profile.id },
              })
            }
          />
        </Card>
      ) : null}

      <Card>
        <Text variant="sectionHeading">Care</Text>
        <Text variant="body" color="secondary">
          Medications and appointments will appear here as they are added.
        </Text>
      </Card>
    </Screen>
  );
}

function DetailRow({ label, value }: { label: string; value: string | null }) {
  const theme = useTheme();

  return (
    <View style={{ gap: theme.spacing.xs }}>
      <Text variant="secondary" color="secondary">
        {label}
      </Text>
      <Text variant="body">{value ?? 'Not added yet'}</Text>
    </View>
  );
}
