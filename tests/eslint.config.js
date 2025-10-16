import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import globals from 'globals';
export default [
  js.configs.recommended,
  prettier,
  {
    files: ['**/*.{js,ts}'],
    languageOptions: {
      ecmaVersion: 2023,
      sourceType: 'module',
      globals: {
        ...globals.node,
        ...globals.nodeBuiltin,
      },
    },
  },
];
