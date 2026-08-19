# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

A collection of Software Engineer technical interview practice problems, each
with a stub file and an isolated test suite. Problems live in `python/` and
`react/`. The repo is designed to be extended by an AI coding agent: prompt
Claude Code to generate a new problem, and it follows the conventions below.
`index.html` is a self-contained, searchable browser for all problems (open
it directly, no build step).

## Commands

### Setup (one-time)

```bash
# Python
python3.11 -m venv .venv && .venv/bin/pip install pytest

# React
cd react && npm install && npx playwright install chromium
```

### Running tests

The unified runner is `run_tests.sh`. It takes an answer file with `-f` and a
test command with `-c`, and wires the answer file into the test run for you.

```bash
# Python: run one problem's suite against an answer file
./run_tests.sh \
  -f python/practice_problem_answers/cw_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py -v

# Python: a single test class or method
./run_tests.sh -f <answer.py> -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole
./run_tests.sh -f <answer.py> -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole::test_empty_permissions_by_default

# React: run one problem's Playwright suite against an answer file
./run_tests.sh -f react/practice_problems/problem_02_incident_dashboard.jsx -c npm run test:02
```

For Python, `run_tests.sh` appends `--answer <abs-path-to-answer-file>`, which
`python/conftest.py` uses to inject the answer module in place of the problem
stub before test collection. This means test files never need their imports
edited, no matter which answer file you point at.

For React, `run_tests.sh` copies the answer `.jsx` file to `react/src/App.jsx`
(the file Vite serves), runs the given command from `react/`, then restores
the previous `App.jsx` on exit.

You can also skip `run_tests.sh` and run tests directly, as long as the
answer file is already in place (for Python, activate `.venv` first and pass
`--answer` yourself; for React, `App.jsx` must already hold the
implementation):

```bash
source .venv/bin/activate
cd python && pytest tests/test_problem_03_permission_manager.py -v --answer /abs/path/to/answer.py

cd react && npm run test:02      # or npm test for all specs, npm run test:ui for interactive mode
```

There is no lint command configured in this repo.

## Architecture

### Python problem/test wiring

Test files always import from the stub module path,
`practice_problems.problem_NN_<name>`, never from an answer file directly.
`python/conftest.py` adds a `--answer` pytest flag: when set, it loads the
given answer file and registers it in `sys.modules` under that same stub
module path before collection runs, so the test's import resolves to the
answer implementation instead. This is why answer files can live anywhere
under `practice_problem_answers/` with any prefix (`cw_answer_`, `en_answer_`,
`my_answer_`, ...) as long as the filename contains a `NN_<name>` segment
matching the stub it targets.

`conftest.py` also defines an autouse fixture, `_reset_class_level_state`,
that clears class-level `set` attributes and mutable default `set` arguments
on any class from an imported practice-problem module after each test. It is
a safety net for practice sessions, not a fix for a real bug: if tests only
pass because of it, the implementation has a state-bleed bug (see "Problem
design rules" below).

### React problem wiring

`react/src/App.jsx` is the single file Vite serves and the single file
Playwright tests exercise. There is one `App.jsx` slot, not one per problem,
so only one problem's implementation can be "active" at a time. Each
`npm run test:NN` script points at a fixed spec file
(`tests/test_problem_NN_<name>.spec.js`) that expects specific `data-testid`
attributes, listed in that problem's docstring. Problem 01 predates this
convention and uses role/text selectors instead.

### The problem index (`index.html`)

`index.html` is a static, dependency-free HTML/JS file with a `PROBLEMS`
array describing every problem (path, test path, title, description,
language, industry, tags, part count, level). It is not generated from the
`practice_problems/` directories, so every new problem must be added to this
array by hand. See "Keeping the problem index up to date" below for the
schema and conventions.

---

## Workflow for a candidate practicing a problem

1. The problem stub in `practice_problems/` has the prompt plus empty
   stubs that raise `NotImplementedError`.
2. Copy the stub to an answer file (Python:
   `practice_problem_answers/<prefix>_answer_NN_<name>.py`; React:
   `react/src/App.jsx` or any `.jsx` file to pass to `run_tests.sh`) and
   implement it there. Keep signatures identical to the stub.
3. Run the test suite for that problem against the answer file using
   `run_tests.sh` (see "Commands" above).

Never edit files in `practice_problems/` or the test files while practicing.

---

## Problem design rules

### Don't mention companies by name
You may receive prompts to design problems related to a specific company. In such a case, do
some research on said company, but try to design the problems to be something you'd generically
see in the sector/industry that company is a part of. In particular, don't use specific company
names in the code because we don't want to piss those companies off, or make them feel like 
we're stealing (which we aren't). Instead, reference a specific sector, like health tech, iot, etc..

### Parts must be self-contained
- Tests for Part N must only call methods/functions defined in Parts 1–N.
- Never verify a Part 1 result by calling a Part 2 helper.
- Any introspection method needed to check state in tests (e.g. `get_role_permissions`,
  `get_usage`, `get_all_permissions`) must live in the **earliest Part that requires it**.

### Methods must compose — no parallel implementations
Design the call chain so that Part N methods **call** Part N-1 methods rather than
re-implementing the same logic with a different return type.

**Anti-pattern to avoid:** A Part 1 method returns `bool` (e.g. `is_allowed`), then a
Part 3 method needs to do the same check but also know *why* it failed (e.g. rate-limited
per-minute vs. per-day). Because the Part 1 method only returns `bool`, Part 3 can't use
it — it has to duplicate all the logic. The candidate ends up writing the same thing twice,
which doesn't reflect good real-world design.

**How to avoid it:** Before finalising the return type of a Part N-1 method, ask: "Could
a higher-level method in a later Part use this return value directly, or would it need to
re-run the same check?" If the answer is "re-run", either:
- Make the lower-level method return richer data (an enum, a tuple, a typed dict), OR
- Introduce a private `_helper` in the problem notes that both methods can call (and
  tell the candidate to implement it first).

**Positive pattern:** Each Part should have a natural "use the method from the Part before"
moment. Design the method signatures so this composition is obvious from context — the docstring
wording and the naming of the methods should make the intended call chain clear without needing
explicit hints.

### Include a concrete usage example
Add a short `# Example` block in the problem docstring showing the data in use and
expected return values for 2–3 key operations. Interviewers always have an example;
the problem file should too.

```python
# Example
# pm = PermissionManager()
# pm.create_role("admin", ["users:write", "billing:read"])
# pm.assign_role("alice", "admin")
# pm.has_permission("alice", "billing:read")  # -> True
# pm.has_permission("alice", "posts:read")    # -> False
```

### Use string IDs, not auto-incrementing integers
String slugs like `"admin"`, `"viewer"`, `"pk.abc123"` are more readable in tests and
don't require the caller to track state. Auto-increment belongs in DB-backed systems,
not in-memory interview problems. If the problem intentionally models a DB entity,
call it out explicitly in the docstring.

### All mutable state must be instance-level
Include this note in every class-based problem docstring:

  "Store all state in instance variables initialized in `__init__`.
   Class-level variables will bleed between tests and between PermissionManager
   instances — avoid them."

---

## Test writing rules

### Use fixtures — never inline `ClassName()` inside test methods
All tests must receive instances via pytest fixtures, not inline construction:

```python
# BAD — bleeds if implementation uses class-level state
def test_creates_role(self):
    p = PermissionManager()
    p.create_role("mod", ...)

# GOOD
def test_creates_role(self, fresh_pm):
    fresh_pm.create_role("mod", ...)
```

Provide both:
- `pm` — a fixture pre-seeded with a realistic set of roles/data
- `fresh_pm` — a bare, empty instance for tests that need a clean slate

### Use unique identifiers per test method
When multiple test methods create objects with string IDs, give each a distinct ID
(e.g. `"role_creates_test"`, `"role_dup_test"`) rather than a shared generic name
like `"mod"`. This reduces false failures when an implementation accidentally stores
state at the class level.

### The conftest safety net
`python/conftest.py` contains an `autouse` fixture (`_reset_class_level_state`) that
runs after every test and clears class-level `set` attributes and mutable default
`set` arguments on all classes found in imported practice-problem modules. This guards
against two common implementation bugs in class-based problems:

- **Class-level sets used as instance state**: `class Foo: SEEN = set()` — shared
  across all instances, persists between tests.
- **Mutable default arguments**: `def __init__(self, roles=set())` — all instances
  share the same `set` object.

This fixture is a diagnostic aid, not a substitute for fixing the implementation.
If tests only pass because of it, there is a bug to fix. Do not remove this fixture.

### Always copy module-level collections in fixtures

When a test file defines module-level data (e.g. `ALL_READINGS`, `ALL_EVENTS`) and a
fixture passes it to a constructor, **always pass a defensive copy**:

```python
import copy

# BAD — if the implementation appends to the passed list, ALL_READINGS is corrupted
@pytest.fixture
def monitor():
    return BiomarkerMonitor(ALL_READINGS)

# GOOD — for lists of immutable objects (dataclasses, strings, ints)
@pytest.fixture
def monitor():
    return BiomarkerMonitor(list(ALL_READINGS))

# GOOD — for nested/mutable structures (dicts of lists, etc.)
@pytest.fixture
def monitor():
    return BiomarkerMonitor(copy.deepcopy(ALL_DATA))
```

Without the copy, a method like `add_reading` that appends to `self._readings` (where
`self._readings` *is* `ALL_READINGS`) will permanently grow the module-level list,
causing later tests to see extra data they didn't add.

**Which copy to use:**
- `list(DATA)` — sufficient when the top-level collection is a flat list of immutable
  objects (dataclass instances, strings, ints). The list is new but the objects inside
  are shared — fine as long as the implementation doesn't mutate the objects themselves.
- `copy.deepcopy(DATA)` — required when the module-level data is a dict, a list of
  dicts, or any nested mutable structure. A shallow copy of a `dict[str, list]` still
  shares the inner lists, so `append` on an inner list still bleeds.

The conftest safety net does not catch this pattern — it only clears class-level sets
and mutable default args, not module-level collections passed in from the test file.

### Test ordering = implementation ordering
Order test classes to match the Part order. A developer who finishes Part 1 and runs
the full suite should see only Part 1 tests passing, with Part 2/3 failing cleanly
due to `NotImplementedError` — not due to test coupling.

### No cross-part dependencies in assertions
A Part 1 test must not fail simply because Part 2 hasn't been implemented. If
checking a Part 1 result requires a function slated for Part 2, move that function
to Part 1.

---

## Problem style

- Target ~45 min completion for a senior engineer.
- Frame problems around a realistic product context but keep core logic generic
  enough to apply across company types (SaaS, API platform, IoT, dev tools, etc.).
- Don't make problems company-specific unless doing targeted prep for a named role.
- Class-based problems (non-existing data structure): the candidate chooses internal data structures. 
  Say so explicitly: "You choose the internal data structures — the public interface is what
  matters."
- Class-based problems: (existing data structure): the problem should have a predefined data
  structure and the problem is just about creating application logic that modifies/utilizes that data 
- Dict-based problems: provide a `make_<thing>()` factory and a clear schema comment.

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

---

## Contribution note

Do not add problems copied verbatim from an actual technical interview. See
"Don't mention companies by name" above; scrub company names and identifying
details before a problem is committed.

Do not modify an existing problem or test file in place once candidates may
already be working from it (this breaks in-progress answer files). Instead,
copy it to a new `_v2`-style file and improve the copy.
