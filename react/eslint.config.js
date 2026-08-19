import js from '@eslint/js'
import reactPlugin from 'eslint-plugin-react'
import reactHooks from 'eslint-plugin-react-hooks'
import tsPlugin from '@typescript-eslint/eslint-plugin'
import tsParser from '@typescript-eslint/parser'
import globals from 'globals'

// Shared permissive overrides — the goal is helpful hints, not a strict gate.
// Unused variables are completely expected in problem stubs and in-progress
// answers, so they're silenced globally.
const lenient = {
  'no-unused-vars': 'off',
  '@typescript-eslint/no-unused-vars': 'off',
  'no-undef': 'error',   // keep — catches genuine missing globals
}

export default [
  // ── Ignored paths ────────────────────────────────────────────────────────
  {
    ignores: ['node_modules/**', 'dist/**', 'test-results/**', 'playwright-report/**'],
  },

  // ── Node.js tooling files ────────────────────────────────────────────────
  // vite.config.js, playwright.config.js, global-setup.js run in Node.
  {
    files: ['*.config.js', 'global-setup.js'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.node },
    },
    rules: {
      ...js.configs.recommended.rules,
      ...lenient,
    },
  },

  // ── Playwright test specs ─────────────────────────────────────────────────
  // Specs run in Node but interact with browser APIs via Playwright's API.
  {
    files: ['tests/**/*.spec.{js,ts}'],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.node, ...globals.browser },
    },
    rules: {
      ...js.configs.recommended.rules,
      ...lenient,
    },
  },

  // ── JavaScript / JSX (stubs + answer files + src) ────────────────────────
  {
    files: ['practice_problems/**/*.{js,jsx}', 'practice_problem_answers/**/*.{js,jsx}', 'src/**/*.{js,jsx}'],
    plugins: {
      react: reactPlugin,
      'react-hooks': reactHooks,
    },
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: 'module',
      globals: { ...globals.browser },
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      ...js.configs.recommended.rules,
      ...reactPlugin.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      ...lenient,
      'react/react-in-jsx-scope': 'off',  // not needed with React 17+ JSX transform
      'react/prop-types': 'off',
      // Downgrade to warnings — helpful hints, not blockers
      'react/jsx-key': 'warn',
      'react/no-unknown-property': 'warn',
    },
  },

  // ── TypeScript / TSX (stubs + answer files + src) ────────────────────────
  {
    files: ['practice_problems/**/*.{ts,tsx}', 'practice_problem_answers/**/*.{ts,tsx}', 'src/**/*.{ts,tsx}'],
    plugins: {
      react: reactPlugin,
      'react-hooks': reactHooks,
      '@typescript-eslint': tsPlugin,
    },
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaFeatures: { jsx: true },
      },
      globals: { ...globals.browser },
    },
    settings: {
      react: { version: 'detect' },
    },
    rules: {
      ...js.configs.recommended.rules,
      ...reactPlugin.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      ...tsPlugin.configs.recommended.rules,
      ...lenient,
      'react/react-in-jsx-scope': 'off',
      'react/prop-types': 'off',
      'react/jsx-key': 'warn',
      'react/no-unknown-property': 'warn',
      '@typescript-eslint/no-explicit-any': 'warn',
    },
  },
]
