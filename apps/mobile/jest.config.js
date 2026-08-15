/**
 * Jest configuration.
 *
 * `rootDir` is pinned to this app because jest-expo's preset otherwise resolves
 * the pnpm workspace root.
 */
module.exports = {
  preset: 'jest-expo',
  rootDir: __dirname,
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  testMatch: ['<rootDir>/src/**/*.test.ts', '<rootDir>/src/**/*.test.tsx'],
  transformIgnorePatterns: [
    'node_modules/(?!((jest-)?react-native|@react-native(-community)?|expo(nent)?|@expo(nent)?/.*|@expo-google-fonts/.*|react-navigation|@react-navigation/.*|native-base|react-native-svg|@supabase/.*|@meracare/.*))',
  ],
};
