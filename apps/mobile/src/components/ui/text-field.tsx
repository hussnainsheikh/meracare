import { StyleSheet, TextInput, View, type TextInputProps } from 'react-native';

import { useTheme } from '@/theme';

import { Text } from './text';

export interface TextFieldProps extends TextInputProps {
  label: string;
  /** Validation message shown below the field. */
  error?: string;
}

/** Labelled text input sized for the accessibility targets in docs/18. */
export function TextField({ label, error, style, ...rest }: TextFieldProps) {
  const theme = useTheme();

  return (
    <View style={{ gap: theme.spacing.xs }}>
      <Text variant="bodyStrong">{label}</Text>
      <TextInput
        accessibilityLabel={label}
        placeholderTextColor={theme.colors.textMuted}
        style={[
          styles.input,
          theme.typography.body,
          {
            minHeight: theme.minTouchTarget,
            paddingHorizontal: theme.spacing.md,
            borderRadius: theme.radii.md,
            backgroundColor: theme.colors.surface,
            borderColor: error ? theme.colors.danger : theme.colors.border,
            color: theme.colors.textPrimary,
          },
          style,
        ]}
        {...rest}
      />
      {/* Errors are announced in words, never signalled by colour alone. */}
      {error ? (
        <Text variant="secondary" color="danger">
          {error}
        </Text>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  input: {
    borderWidth: 1,
  },
});
