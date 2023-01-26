module.exports = {
  extends: [
    // By extending from a plugin config, we can get recommended rules without having to add them manually.
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:import/recommended",
    "plugin:jsx-a11y/recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    // This disables the formatting rules in ESLint that Prettier is going to be responsible for handling.
    // Make sure it's always the last config, so it gets the chance to override other configs.
    "eslint-config-prettier",
  ],
  settings: {
    react: {
      // Tells eslint-plugin-react to automatically detect the version of React to use.
      version: "detect",
    },
    // Tells eslint how to resolve imports
    "import/resolver": {
      node: {
        paths: ["src"],
        extensions: [".js", ".jsx", ".ts", ".tsx"],
      },
    },
  },
  env: {
    browser: true,
  },

  rules: {
    // PLEASE ALWAYS PROVIDE A REASON FOR DISABLING/OVERWRITING A RULE
    // IT'S HARD TO EVALUATE THIS SECTION AT A LATER POINT IN TIME

    // It's save to import React when using vite
    "react/react-in-jsx-scope": "off",

    // console logs are fine in development, but eslint can help us
    // remember to remove them. console.error and console.warn are
    // allowed for now. They should only be placed in code that should
    // not be reached and can provide a helpful hint to the developer.
    "no-console": ["error", { allow: ["error", "warn"] }],

    // REMOVE WHEN 100 % TYPESCRIPT IS ACHIEVED

    // we will use TypeScript's types for component props instead)
    "react/prop-types": "off",
    // this is the default in typescript and we want to enforce it in JavaScript as well
    "prefer-const": "error",
  },
};
