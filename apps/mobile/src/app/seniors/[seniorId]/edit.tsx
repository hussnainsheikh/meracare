import { Stack, router, useLocalSearchParams } from 'expo-router';
import { useEffect, useState } from 'react';
import { ActivityIndicator, View } from 'react-native';

import { Button, Card, Screen, Text, TextField } from '@/components/ui';
import { useSenior, useUpdateSenior } from '@/features/seniors/use-seniors';
import { ApiError } from '@/lib/api-error';
import { useTheme } from '@/theme';

/**
 * Edit a senior profile.
 *
 * The API enforces `senior.edit`; a caller without it receives 404 here just as
 * they would anywhere else, so hiding the entry point is convenience, not
 * security (docs/02).
 */
export default function EditSeniorScreen() {
  const theme = useTheme();
  const { seniorId } = useLocalSearchParams<{ seniorId: string }>();
  const senior = useSenior(seniorId ?? null);
  const updateSenior = useUpdateSenior(seniorId ?? '');

  const [form, setForm] = useState({
    displayName: '',
    dateOfBirth: '',
    phone: '',
    address: '',
    emergencyContact: '',
  });
  const [loaded, setLoaded] = useState(false);

  // Seed the form once, so a background refetch cannot overwrite typing.
  useEffect(() => {
    if (loaded || !senior.data) return;
    setForm({
      displayName: senior.data.displayName,
      dateOfBirth: senior.data.dateOfBirth ?? '',
      phone: senior.data.phone ?? '',
      address: senior.data.address ?? '',
      emergencyContact: senior.data.emergencyContact ?? '',
    });
    setLoaded(true);
  }, [loaded, senior.data]);

  const fieldErrors =
    updateSenior.error instanceof ApiError ? (updateSenior.error.details ?? {}) : {};

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
    return (
      <Screen scrollable>
        <Card>
          <Text variant="sectionHeading">This profile is not available</Text>
          <Button
            variant="secondary"
            label="Back to Today"
            onPress={() => router.replace('/home')}
          />
        </Card>
      </Screen>
    );
  }

  async function handleSave() {
    await updateSenior.mutateAsync({
      displayName: form.displayName.trim(),
      // An empty string clears the field; the API treats it as "remove".
      dateOfBirth: form.dateOfBirth.trim() || null,
      phone: form.phone.trim() || null,
      address: form.address.trim() || null,
      emergencyContact: form.emergencyContact.trim() || null,
    });
    router.back();
  }

  const update = (field: keyof typeof form) => (value: string) =>
    setForm((current) => ({ ...current, [field]: value }));

  return (
    <Screen scrollable>
      <Stack.Screen options={{ headerShown: true, title: 'Edit profile' }} />

      <Text variant="pageHeading">Edit profile</Text>

      <Card>
        <TextField
          label="Name"
          value={form.displayName}
          onChangeText={update('displayName')}
          autoCapitalize="words"
          error={fieldErrors.displayName}
        />
        <TextField
          label="Date of birth"
          value={form.dateOfBirth}
          onChangeText={update('dateOfBirth')}
          placeholder="YYYY-MM-DD"
          autoCapitalize="none"
          inputMode="numeric"
          error={fieldErrors.dateOfBirth}
        />
        <TextField
          label="Phone"
          value={form.phone}
          onChangeText={update('phone')}
          inputMode="tel"
          error={fieldErrors.phone}
        />
        <TextField
          label="Address"
          value={form.address}
          onChangeText={update('address')}
          multiline
          error={fieldErrors.address}
        />
        <TextField
          label="Emergency contact"
          value={form.emergencyContact}
          onChangeText={update('emergencyContact')}
          placeholder="Name and phone number"
          error={fieldErrors.emergencyContact}
        />

        {updateSenior.isError && Object.keys(fieldErrors).length === 0 ? (
          <Text variant="secondary" color="danger">
            {updateSenior.error instanceof ApiError
              ? updateSenior.error.message
              : 'We could not save those changes. Please try again.'}
          </Text>
        ) : null}

        <Button
          label="Save changes"
          onPress={handleSave}
          disabled={form.displayName.trim().length === 0}
          loading={updateSenior.isPending}
        />
        <Button variant="ghost" label="Cancel" onPress={() => router.back()} />
      </Card>
    </Screen>
  );
}
