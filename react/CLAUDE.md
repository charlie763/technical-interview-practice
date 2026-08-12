# React — Agent Guidelines

## Workflow

### One-time setup

```bash
cd react
npm install
```

Playwright browsers only need installing once (Chromium is used):

```bash
npx playwright install chromium
```

### Working on a problem

All answer files live in `react/practice_problem_answers/` and are never shared.
`src/App.jsx` is a permanent placeholder — never edit it during practice.

The `PRACTICE_ANSWER` environment variable points Vite (and Playwright) at your
answer directory. The Vite plugin in `vite.config.js` intercepts `main.jsx`'s
`./App.jsx` import and redirects it to `<PRACTICE_ANSWER>/App.jsx`.

1. Create your answer directory and copy the stub into it:

   ```bash
   cd react
   mkdir -p practice_problem_answers/cw_answer_02_incident_dashboard
   cp practice_problems/problem_02_incident_dashboard.jsx \
      practice_problem_answers/cw_answer_02_incident_dashboard/App.jsx
   ```

2. Start the Vite dev server pointing at your answer:

   ```bash
   PRACTICE_ANSWER=practice_problem_answers/cw_answer_02_incident_dashboard npm run dev
   ```

   This opens http://localhost:5173 and hot-reloads your answer file.
   You can create additional files (components, hooks, etc.) inside the answer
   directory and import them normally — Vite resolves them relative to your `App.jsx`.

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
