module.exports = {
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
    "../src/componentsNext/**/*.mdx",
    "../src/componentsNext/**/*.stories.@(js|jsx|ts|tsx)",
    "../src/hooksNext/**/*.mdx",
    "../src/hooksNext/**/*.stories.@(js|jsx|ts|tsx)",
  ],

  addons: [
    "@storybook/addon-essentials",
    "@storybook/addon-interactions",
    "storybook-addon-react-router-v6",
  ],
  framework: {
    name: "@storybook/react-vite",
    options: {},
  },
  docs: {
    autodocs: true,
  },
};
