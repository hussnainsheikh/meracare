import { darkColors, lightColors, minTouchTarget, typography } from '../tokens';

describe('theme tokens', () => {
  it('uses the locked Deep Teal brand colour in light mode', () => {
    expect(lightColors.primary).toBe('#0F766E');
  });

  it('defines every semantic role in both themes', () => {
    expect(Object.keys(darkColors).sort()).toEqual(Object.keys(lightColors).sort());

    for (const [role, value] of Object.entries(darkColors)) {
      expect(value).toMatch(/^#[0-9A-Fa-f]{6}$/);
      expect(lightColors[role as keyof typeof lightColors]).toMatch(/^#[0-9A-Fa-f]{6}$/);
    }
  });

  it('keeps touch targets at the 48dp accessibility minimum', () => {
    expect(minTouchTarget).toBeGreaterThanOrEqual(48);
  });

  it('keeps body and action text large enough for older adults', () => {
    expect(typography.body.fontSize).toBeGreaterThanOrEqual(16);
    expect(typography.action.fontSize).toBeGreaterThanOrEqual(16);
    expect(typography.pageHeading.fontSize).toBeGreaterThanOrEqual(28);
  });
});
