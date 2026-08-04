# Technical Interview Practice

A collection of Software Engineer practice problems with test suites, designed to be generated and extended by an AI coding agent (Claude Code or similar).

---

## Design Principles
- It should be easy to get AI to add new practice problems that can (1) be tested under the existing framework and (2) be searcheable by opening index.html
- It should be easy to find relevant practice problems to work on
- SWEs practice should be able to run tests against their practice problem answer files without having to update any code besides the code in their answer file
- Tests should isolated and not bleed data
- Practice problem logic should be cumalitive; later stage logic should utilize earlier stage logic where possible
- Practice problems should be strive to realistic represent practice problems an SWE might encounter in a real technical interview
- This repo is meant to be used locally

## Repo structure
More directories may be added over time as problems in different languages are added, but the overall structure should stay the same.

```
index.html                    # Searchable problem browser — open in any browser
python/
  practice_problems/          # Problem stubs — read these, don't edit them
  practice_problem_answers/   # Your implementations go here
  tests/                      # pytest suites (one per problem)
  conftest.py                 # pytest config & --answer flag
react/
  practice_problems/          # React/JSX starter files
  src/App.jsx                 # Active problem (overwrite to work; restore with git)
  src/main.jsx                # Vite entry point (do not modify)
  tests/                      # Playwright e2e specs (one per problem)
  package.json                # Vite + Playwright deps
  playwright.config.js        # Playwright config (auto-starts Vite)
  vite.config.js
run_tests.sh                  # Unified test runner for Python and React (see below)
CLAUDE.md                     # Guidelines for the AI agent
```

---

## Prerequisites

### Python (one-time setup)

```bash
# From the repo root
python3.11 -m venv .venv
.venv/bin/pip install pytest
```

`run_tests.sh` activates `.venv` automatically, so you don't need to do it manually before running tests.

### React (one-time setup)

```bash
cd react
npm install
npx playwright install chromium   # download the test browser (~95 MB)
```

---

## Browsing problems

Open `index.html` in any browser (double-click it in Finder, or run `open index.html`).

The index lets you filter by language and industry, search by keyword, and click tags to narrow results. Each card has an **Open in VS Code** link that deep-links directly into your local clone — no path configuration needed.

---

## Practicing a problem

### 1. Find a problem

Browse `index.html` to pick something, then click **Open in VS Code** to open the stub.

---

### Python problems

#### 2. Copy the stub to your answers directory

```bash
cp python/practice_problems/problem_03_permission_manager.py \
   python/practice_problem_answers/my_answer_03_permission_manager.py
```

The prefix (`my_answer_`, `cw_answer_`, etc.) can be anything — the filename **must** keep the `NN_<name>` segment (e.g. `03_permission_manager`) so the runner can map it to the right test suite.

#### 3. Implement it

Fill in the `raise NotImplementedError` stubs. Keep the function/class signatures identical to the stub.

#### 4. Run tests against your answer

```bash
# Full test suite
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py -v

# Single test class
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole

# Single test method
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole::test_empty_permissions_by_default

# Stop on first failure
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py -x
```

`--answer` is injected automatically — you never need to edit the test files.

---

### React problems

#### 2. Copy the stub to `react/src/App.jsx`

```bash
cp react/practice_problems/problem_02_incident_dashboard.jsx \
   react/src/App.jsx
```

#### 3. Implement it in a browser

```bash
cd react && npm run dev    # http://localhost:5173
```

Edit `react/src/App.jsx` — Vite hot-reloads on every save.

#### 4. Run Playwright tests against your answer

Use `run_tests.sh` the same way as Python — just pass the `.jsx` file.
The script copies it to `react/src/App.jsx`, runs the tests, and restores the placeholder when done:

```bash
# Run the Playwright suite for problem 02
./run_tests.sh \
  -f react/practice_problems/problem_02_incident_dashboard.jsx \
  -c npm run test:02

# If you saved your work to a separate file instead:
./run_tests.sh \
  -f react/my_answer_02_incident_dashboard.jsx \
  -c npm run test:02

# Open Playwright's interactive UI (step through each test visually)
./run_tests.sh \
  -f react/practice_problems/problem_02_incident_dashboard.jsx \
  -c npm run test:ui
```

You can also run tests directly from the `react/` directory if `src/App.jsx`
already contains your implementation:

```bash
cd react
npm run test:02     # problem 02 only
npm run test:03     # problem 03 only
npm run test:ui     # interactive Playwright UI for all problems
npm test            # all 3 spec files
```

#### 5. Reset the placeholder when done

```bash
git restore react/src/App.jsx
```

## Adding new problems with an AI agent
⚠️ WARNING - PLEASE READ: If contributing, please do not add any problems verbatim from actual technical interviews. We don't want to get each other in trouble or cause issues for people actively interviewing. The `CLAUDE.md` file has instructions to scrub actual company names from problems, but please double check the code for that before submitting a PR.

This repo is designed to be extended by prompting an AI coding agent. Open Claude Code (or your preferred agent) in this repo and describe what you want. The agent will read `CLAUDE.md` for conventions automatically.

A good prompt specifies: the language, the industry/domain, the core data structure or algorithm pattern, the number of parts, and any constraints on problem style. Example:

```
Generate a new Senior SWE interview problem for the python/ directory.

- Domain: dev-tools / CI-CD infrastructure
- Core concept: a class-based job scheduler that manages a queue of build jobs
  with priorities, dependencies between jobs, and cancellation
- 3 progressive parts: Part 1 basic enqueue/dequeue, Part 2 job dependencies
  (a job can't start until its dependencies complete), Part 3 priority ordering
  within the ready queue
- Style: class-based, candidate chooses internal data structures
- Follow all conventions in CLAUDE.md (fixtures, no class-level state,
  self-contained parts, composing methods, concrete usage example in docstring)
- [optional] - please checkout this JD/company website, do some research on the company
   to see if the company asks a specific style of interview questions and base the practice 
   problem(s) off of that.

After creating the problem and test files, add an entry to the PROBLEMS array
in index.html following the format in CLAUDE.md.
```

The agent will create:
- `python/practice_problems/problem_NN_<name>.py` — the stub
- `python/tests/test_problem_NN_<name>.py` — the test suite
- An entry in `index.html`

DO NOT ask it to improve existing problems or test suites, unless you don't plan to merge your branch into `main`
as that will likely break other peoples' answers. Instead, if you would like to improve upon an existing problem/test
just copy and paste it into a new problem/test file with something like `v.x` appended to the file name, and then try
to improve the problem and/or test suite.


