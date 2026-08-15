// Metro configuration for the pnpm workspace.
// https://docs.expo.dev/guides/monorepos/
const { getDefaultConfig } = require('expo/metro-config');
const path = require('node:path');

const projectRoot = __dirname;
const workspaceRoot = path.resolve(projectRoot, '../..');

const config = getDefaultConfig(projectRoot);

// Watch the workspace so `@meracare/contracts` source is transformed by Metro.
config.watchFolders = [workspaceRoot];

config.resolver.nodeModulesPaths = [
  path.resolve(projectRoot, 'node_modules'),
  path.resolve(workspaceRoot, 'node_modules'),
];
// Resolve every package from the app's own tree first, so a single React copy
// is used across the workspace.
config.resolver.disableHierarchicalLookup = true;

module.exports = config;
