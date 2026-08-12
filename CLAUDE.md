# Technical Interview Practice — Agent Guidelines

For language-specific instructions, see:
- **Python:** [`python/CLAUDE.md`](python/CLAUDE.md)
- **React:** [`react/CLAUDE.md`](react/CLAUDE.md)

## Repo structure

```
python/
  practice_problems/          # problem stubs (read-only during practice)
  practice_problem_answers/   # cw_answer_XX_... files (filled in by Charlie)
  tests/                      # pytest suites
react/
  practice_problems/          # JSX/TSX starter files (read-only during practice)
  practice_problem_answers/   # cw_answer_NN_<name>/App.jsx files (filled in by Charlie)
  src/App.jsx                 # placeholder — never edit during practice
  src/main.jsx                # Vite entry point (do not modify)
  tests/                      # Playwright e2e specs
  package.json                # Vite + Playwright deps
  playwright.config.js
  vite.config.js
```

## Problem design rules (all languages)

### Don't mention companies by name
You may receive prompts to design problems related to a specific company. Do some research
on that company but design problems generically for the sector/industry. Don't use specific
company names in the code. Instead, reference a specific sector like health tech, IoT, etc.

### Parts must be self-contained
- Tests for Part N must only call methods/functions defined in Parts 1–N.
- Never verify a Part 1 result by calling a Part 2 helper.

### Target ~45 min completion for a senior engineer
Frame problems around a realistic product context but keep core logic generic enough to
apply across company types (SaaS, API platform, IoT, dev tools, etc.).

---

## Keeping the problem index up to date

`index.html` at the repo root is a self-contained searchable index of all practice
problems. **Every time you create a new problem, add an entry to the `PROBLEMS` array**
near the top of the `<script>` block in that file (look for the comment that says
`PROBLEM INDEX — add new problems here`).

### Entry format

```javascript
{
  path: "python/practice_problems/problem_NN_<name>.py",  // path to the stub file
  test: "python/tests/test_problem_NN_<name>.py",         // path to the test file (null for React problems without a separate test)
  title: "Short Human-Readable Title",                     // shown as the card heading
  description: "One or two sentences describing what the candidate builds.",
  language: "python",           // "python" | "react"
  industry: "health-tech",      // see valid values below
  tags: ["tag-one", "tag-two"], // 2–5 kebab-case strings
  parts: 3,                     // number of implementation parts
  level: "senior"               // "junior" | "mid-level" | "senior" | "staff"
}
```

### Valid `language` values
`python` | `react`

### Valid `level` values (in order)
`junior` | `mid-level` | `senior` | `staff`

### Suggested `industry` values (add new ones as needed, always kebab-case)
`general` | `health-tech` | `iot` | `dev-tools` | `fintech`

Use `"general"` when a problem has no clear real-world industry vertical — e.g. a
rate limiter, a permissions system, or an activity feed. These are generic software
engineering patterns that appear everywhere; tag them with the relevant product
category instead (e.g. `"saas"`, `"api-platform"`).

**Industry = real-world vertical** (healthcare, finance, logistics).
**Tag = product/platform category or algorithmic pattern** (saas, api-platform, rate-limiting).

If you introduce a **new** industry value, also add a matching CSS rule to the
`<style>` block in `index.html` so its badge renders with distinct colours:

```css
.badge-my-new-industry { background: #f0f0f0; color: #333; }
```

### Tag conventions
Kebab-case strings that describe the core algorithmic pattern or domain concept.
Examples: `sliding-window`, `rbac`, `event-driven`, `time-series`,
`consecutive-tracking`, `deadline-tracking`, `optimistic-updates`.
Aim for 2–5 tags per problem.
