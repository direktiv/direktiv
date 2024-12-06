export default {
  typescript: {
    check: false,
    checkOptions: {},
    reactDocgen: "react-docgen-typescript",
    reactDocgenTypescriptOptions: {
      shouldExtractLiteralValuesFromEnum: true,
      propFilter: (prop) =>
        prop.parent ? !/node_modules/.test(prop.parent.fileName) : true,
    },
  },
  stories: [
    "../src/design/**/*.mdx",
    "../src/design/**/*.stories.@(js|jsx|ts|tsx)",
  ],

  addons: [
    "@storybook/addon-essentials",
    "@storybook/addon-interactions",
    "@chromatic-com/storybook",
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  docs: {},
  // https://github.com/chromaui/chromatic-cli/issues/550#issuecomment-1326856720
  viteFinal: (config) => ({
    ...config,
    build: {
      ...config.build,
      sourcemap: false,
      target: ["es2020"],
    },
  }),
};
