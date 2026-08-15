// Metro applies babel-preset-expo automatically; Jest needs it declared here so
// test files go through the same transform.
module.exports = function (api) {
  api.cache(true);
  return {
    presets: [['babel-preset-expo', { reactCompiler: true }]],
  };
};
