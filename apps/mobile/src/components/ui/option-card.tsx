import { Pressable, StyleSheet, View } from 'react-native';

import { useTheme } from '@/theme';

import { Text } from './text';

export interface OptionCardProps {
  title: string;
  description: string;
  selected: boolean;
  onPress: () => void;
}

/**
 * A large, selectable choice.
 *
 * Selection is shown by a border, a tinted surface, and a checkmark together —
 * docs/18 requires that information never depend on colour alone.
 */
export function OptionCard({ title, description, selected, onPress }: OptionCardProps) {
  const theme = useTheme();

  return (
    <Pressable
      accessibilityRole="radio"
      accessibilityState={{ selected }}
      onPress={onPress}
      style={({ pressed }) => [
        styles.base,
        {
          minHeight: theme.minTouchTarget,
          padding: theme.spacing.lg,
          borderRadius: theme.radii.lg,
          gap: theme.spacing.xs,
          backgroundColor: selected ? theme.colors.primaryLight : theme.colors.surface,
          borderColor: selected ? theme.colors.primary : theme.colors.border,
          borderWidth: selected ? 2 : StyleSheet.hairlineWidth,
          opacity: pressed ? 0.85 : 1,
        },
      ]}
    >
      <View style={styles.header}>
        <Text variant="bodyStrong">{title}</Text>
        {selected ? (
          <Text variant="bodyStrong" color="brand">
            ✓
          </Text>
        ) : null}
      </View>
      <Text variant="secondary" color="secondary">
        {description}
      </Text>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    justifyContent: 'center',
  },
  header: {
    alignItems: 'center',
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
});
