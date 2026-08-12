# React — Agent Guidelines

## Workflow

### One-time setup

```bash
cd react
npm install
npx playwright install chromium
```

### Working on a problem

All answer files live in `react/practice_problem_answers/` and are never shared.
`src/App.jsx` is a permanent placeholder — never edit it during practice.

The `PRACTICE_ANSWER` environment variable points Vite (and Playwright) at your
answer directory. The Vite plugin in `vite.config.js` intercepts `main.jsx`'s
`./App` import and redirects it to the answer's `App.jsx` or `App.tsx`.

1. Create your answer directory and copy the stub into it:

   ```bash
   cd react
   mkdir -p practice_problem_answers/cw_answer_02_incident_dashboard
   cp practice_problems/problem_02_incident_dashboard.jsx \
      practice_problem_answers/cw_answer_02_incident_dashboard/App.jsx
   ```

   For TypeScript problems, use `.tsx`:

   ```bash
   mkdir -p practice_problem_answers/cw_answer_05_ts_example
   cp practice_problems/problem_05_ts_example.tsx \
      practice_problem_answers/cw_answer_05_ts_example/App.tsx
   ```

2. Start the Vite dev server pointing at your answer:

   ```bash
   PRACTICE_ANSWER=practice_problem_answers/cw_answer_02_incident_dashboard npm run dev
   ```

   This opens http://localhost:5173 and hot-reloads your answer file.
   You can create additional files (components, hooks, etc.) inside the answer
   directory and import them normally — Vite resolves them relative to your `App.jsx`/`App.tsx`.

3. Run the Playwright tests against your answer:

   ```bash
   PRACTICE_ANSWER=practice_problem_answers/cw_answer_02_incident_dashboard npm run test:02
   ```

   When `PRACTICE_ANSWER` is set, Playwright always spawns a fresh dev server
   (ignoring any server already running on port 5173) so it picks up the env var.
   Stop any existing `npm run dev` session before running tests to avoid port conflicts.

   Or open interactive Playwright UI mode:

   ```bash
   PRACTICE_ANSWER=practice_problem_answers/cw_answer_02_incident_dashboard npm run test:ui
   ```

No `git restore` needed — `src/App.jsx` is never touched.

### Test commands

| Command | What it runs |
|---|---|
| `npm run test:01` | Problem 01 — Activity Feed |
| `npm run test:02` | Problem 02 — Incident Dashboard |
| `npm run test:03` | Problem 03 — Alert Triage Console |
| `npm test` | All spec files |
| `npm run test:ui` | Playwright interactive UI |

Prefix any test command with `PRACTICE_ANSWER=practice_problem_answers/cw_answer_NN_<name>`.

---

## TypeScript problems

### What makes a problem a TypeScript problem

A problem is TypeScript when its stub file has a `.tsx` (or `.ts`) extension. The
candidate's answer file must also be `.tsx` / `.ts`. Using `.jsx` for a TypeScript
problem stub is not valid.

### How TypeScript is enforced

When `npm run test:NN` is run and the answer directory contains any `.ts` or `.tsx`
files, Playwright's global setup (`global-setup.js`) automatically runs `tsc --noEmit`
**before** opening a browser. A type error aborts the test run immediately.

You can also run the type-check manually at any time:

```bash
cd react
npm run typecheck
```

### TypeScript config (`tsconfig.json`)

The config is intentionally permissive — the goal is to verify basic TypeScript
literacy, not to require production-grade strictness:

| Setting | Value | Effect |
|---|---|---|
| `noImplicitAny` | `true` | Parameters / variables whose type can't be inferred must be annotated |
| `strictNullChecks` | `false` | `null` and `undefined` are assignable anywhere |
| `strict` | *(not set)* | All other strict checks are off |
| `allowJs` | `false` | Only `.ts` / `.tsx` files are type-checked |
| `skipLibCheck` | `true` | Declaration files in `node_modules` are not checked |

### Writing TypeScript problem stubs

- Name the stub file `.tsx` (component) or `.ts` (pure logic).
- The Vite redirect plugin prefers `App.tsx` over `App.jsx` when both exist.
- Define all public interfaces / types in the stub using `// TODO` placeholders
  so the candidate knows what's expected.
- Add a `// TYPE CONTRACT` block listing required types/interfaces, similar to
  the `// TEST CONTRACT` block used for testids in JSX problems.

---

## Linting (ESLint)

`eslint.config.js` covers both `.jsx` and `.tsx` files with React + hooks rules.
TypeScript files additionally get `@typescript-eslint` rules (errors downgraded to
warnings — `tsc` is the authoritative type gate).

Run manually:

```bash
cd react
npm run lint
```

### Lint on save in VS Code

The repo-root `.vscode/settings.json` configures ESLint auto-fix on save for
`.js`, `.jsx`, `.ts`, and `.tsx` files. This requires the
[ESLint VS Code extension](https://marketplace.visualstudio.com/items?itemName=dbaeumer.vscode-eslint).
No additional setup is needed — `eslint.workingDirectories` already points the
extension at `react/` where `eslint.config.js` lives.

---

## Test query strategy

Playwright specs follow RTL's query priority (excluding `getByRole`, which
requires accessibility knowledge not expected in a coding interview):

1. **Seed/mock data text** — `toContainText('Downtown')` etc. Used wherever
   the problem provides deterministic seed data whose text will appear on screen.
2. **Element type + visible text** — `locator('button').filter({ hasText: /ack/i })`.
   Used for buttons whose labels are specified in the problem (e.g. "Ack", "Contain",
   "Assign", "Ack All"). Candidate must use the specified label, but no `data-testid`
   is required on these elements.
3. **Input attributes** — `getByPlaceholder(/search/i)` for search inputs. The problem
   specifies the placeholder text.
4. **`data-testid`** — last resort, used only when structural identification is
   unavoidable (counting rows, distinguishing multiple selects on the same page,
   editable cells, spinners). The problem file lists every required testid explicitly.

**When writing new problems:** document required testids in a "TEST CONTRACT" or
"Required data-testid attributes" block in the problem docstring. Prefer button-text
or placeholder approaches over testids for interactive controls whenever the label
text is fixed by the problem spec.
