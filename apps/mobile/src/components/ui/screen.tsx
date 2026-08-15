import { ScrollView, StyleSheet, View, type ViewProps } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';

import { useTheme } from '@/theme';

export interface ScreenProps extends ViewProps {
  /** Wraps the content in a ScrollView. Use for forms and long content. */
  scrollable?: boolean;
}

/**
 * Page container: applies the themed background and the safe-area insets that
 * every screen needs.
 */
export function Screen({ scrollable = false, style, children, ...rest }: ScreenProps) {
  const theme = useTheme();
  const insets = useSafeAreaInsets();

  const contentStyle = [
    {
      padding: theme.spacing.lg,
      paddingTop: insets.top + theme.spacing.lg,
      paddingBottom: insets.bottom + theme.spacing.lg,
      gap: theme.spacing.lg,
    },
    style,
  ];

  if (scrollable) {
    return (
      <ScrollView
        style={[styles.fill, { backgroundColor: theme.colors.background }]}
        contentContainerStyle={contentStyle}
        keyboardShouldPersistTaps="handled"
      >
        {children}
      </ScrollView>
    );
  }

  return (
    <View
      style={[styles.fill, { backgroundColor: theme.colors.background }, contentStyle]}
      {...rest}
    >
      {children}
    </View>
  );
}

const styles = StyleSheet.create({
  fill: {
    flex: 1,
  },
});
