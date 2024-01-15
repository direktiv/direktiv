module.exports = {
  extends: [
    // By extending from a plugin config, we can get recommended rules without having to add them manually.
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:import/recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended",
    "plugin:storybook/recommended",
    "plugin:@tanstack/eslint-plugin-query/recommended",
    "plugin:tailwindcss/recommended",
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
      typescript: {},
      node: {
        paths: ["src"],
        extensions: [".js", ".jsx", ".ts", ".tsx"],
      },
    },
  },
  env: {
    browser: true,
    node: true,
  },
  ignorePatterns: [
    "public/",
    "node_modules/",
    "dist/",
    ".eslintrc.js",
    "env.d.ts",
    "playwright-report/",
  ],
  rules: {
    // PLEASE ALWAYS PROVIDE A REASON FOR DISABLING/OVERWRITING A RULE
    // IT'S HARD TO EVALUATE THIS SECTION AT A LATER POINT IN TIME

    // Imports can be very messy in JavaScript and we should automatically
    // sort them to make them more readable and consistent accross the project.
    // this plugin does that automatically and this rule enforces it.
    // https://marketplace.visualstudio.com/items?itemName=amatiasq.sort-imports
    "sort-imports": "error",

    // nested ternary operators are hard to read and should be avoided
    "no-nested-ternary": "error",

    // this rule will enforce the object shorthand syntax and prefers { something }
    // over { something: something }
    // this is more about consitencey than right or wrong
    "object-shorthand": "error",

    // prefer arrow functions over function declarations, this is more for consistency
    // than for right or wrong. Also VSCode can automatically convert them when using
    // comman + .
    "arrow-body-style": "error",

    // simpe rule to avoid unnecessary curly braces
    "react/jsx-curly-brace-presence": "error",

    // It's save to import React when using vite
    "react/react-in-jsx-scope": "off",

    // overwriting the default to use tailwind.config.cjs instead of tailwind.config.js
    "tailwindcss/no-custom-classname": [
      "error",
      {
        config: "tailwind.config.cjs",
      },
    ],

    // there seems to be a missmatch between the order
    // of the classes from the linting rule vs the prettier plugin
    // since prettier is part of the dev environment, we can disable
    // the rule
    "tailwindcss/classnames-order": "off",

    // console logs are fine in development, but eslint can help us
    // remember to remove them. console.error and console.warn are
    // allowed for now. They should only be placed in code that should
    // not be reached and can provide a helpful hint to the developer.
    "no-console": ["error", { allow: ["error", "warn"] }],

    // unused variables should not be allowed in general, but you can use the
    // underscore prefix to indicate that the variable is unused on purpose
    // like f.e. in a function signature to indicate that the function takes
    // a parameter but does not use it
    "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],

    // REMOVE WHEN 100 % TYPESCRIPT IS ACHIEVED

    // we will use TypeScript's types for component props instead)
    "react/prop-types": "off",
    // this is the default in typescript and we want to enforce it in JavaScript as well
    "prefer-const": "error",
  },
  overrides: [
    {
      // allow require statements for commonjs modules
      files: ["*.cjs", "*.cts"],
      rules: {
        "@typescript-eslint/no-var-requires": "off",
      },
    },
  ],
};
