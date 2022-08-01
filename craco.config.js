const BundleAnalyzerPlugin =
  require("webpack-bundle-analyzer").BundleAnalyzerPlugin;

module.exports = function () {
  return {
    webpack: {
      plugins: [new BundleAnalyzerPlugin({ analyzerMode: "server" })],
    },
  };
};
