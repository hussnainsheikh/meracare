import { Text as RNText, type TextProps as RNTextProps } from 'react-native';

import { useTheme, type TypographyVariant } from '@/theme';

type ColorRole = 'primary' | 'secondary' | 'muted' | 'brand' | 'success' | 'warning' | 'danger';

export interface TextProps extends RNTextProps {
  variant?: TypographyVariant;
  color?: ColorRole;
}

/**
 * Typography primitive. Every piece of text in the app goes through this so the
 * scale and colour roles stay consistent (docs/18).
 */
export function Text({ variant = 'body', color = 'primary', style, ...rest }: TextProps) {
  const theme = useTheme();

  const colorValue = {
    primary: theme.colors.textPrimary,
    secondary: theme.colors.textSecondary,
    muted: theme.colors.textMuted,
    brand: theme.colors.primary,
    success: theme.colors.success,
    warning: theme.colors.warning,
    danger: theme.colors.danger,
  }[color];

  return <RNText style={[theme.typography[variant], { color: colorValue }, style]} {...rest} />;
}
