// @ts-check
import js from '@eslint/js';
import globals from 'globals';
import react from 'eslint-plugin-react';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import tseslint from 'typescript-eslint';
import jsxA11y from 'eslint-plugin-jsx-a11y';
import prettier from 'eslint-plugin-prettier/recommended';
import { fixupPluginRules } from '@eslint/compat';
import { defineConfig } from 'eslint/config';

export default defineConfig(
  {
    ignores: ['dist', 'node_modules', 'build', 'coverage'],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,

  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 'latest',
      sourceType: 'module',
      globals: {
        ...globals.browser,
        ...globals.es2020,
      },
      parserOptions: {
        projectService: true,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    plugins: {
      react: react,
      'react-hooks': fixupPluginRules(/** @type {any} */ (reactHooks)),
      'react-refresh': reactRefresh,
      'jsx-a11y': jsxA11y,
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      // React Hooks safety
      ...reactHooks.configs.recommended.rules,

      // Accessibility safety (prevents broken UI for users)
      ...jsxA11y.configs.recommended.rules,

      // Custom rules for your Registry project
      'react-refresh/only-export-components': ['warn', { allowConstantExport: true }],
      'react/react-in-jsx-scope': 'off', // Not needed for React 17+
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],

      // Security: Prevent dangerouslySetInnerHTML without review
      'react/no-danger': 'error',
      'react-hooks/rules-of-hooks': 'error',
      'react-hooks/exhaustive-deps': 'error',
    },
  },

  // 3. Prettier MUST be last to ensure it wins formatting conflicts
  prettier
);
