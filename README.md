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
More directories may be added over time as problems in d ifferent languages are added, but the overall structure should say the same

```
index.html                    # Searchable problem browser — open in any browser
python/
  practice_problems/          # Problem stubs — read these, don't edit them
  practice_problem_answers/   # Your implementations go here
  tests/                      # pytest suites (one per problem)
  conftest.py                 # pytest config & --answer flag
react/
  practice_problems/          # React/JSX starter files
run_tests.sh                  # Test runner (see below)
CLAUDE.md                     # Guidelines for the AI agent
```

---

## Prerequisites

```bash
# Create and activate the virtual environment (one-time setup)
python3 -m venv .venv
source .venv/bin/activate
pip install pytest

# React problems (coming soon)
# npm install
```

The `run_tests.sh` script automatically activates `.venv` if it exists in the repo root, so you don't need to activate it manually before running tests.

---

## Browsing problems

Open `index.html` in any browser (double-click it in Finder, or run `open index.html`).

The index lets you filter by language and industry, search by keyword, and click tags to narrow results. Each card has an **Open in VS Code** link that deep-links directly into your local clone — no path configuration needed.

---

## Practicing a problem

### 1. Find a problem

Browse `index.html` to pick something, then click **Open in VS Code** to open the stub.

### 2. Copy it to your answers directory

```bash
cp python/practice_problems/problem_03_permission_manager.py \
   python/practice_problem_answers/my_answer_03_permission_manager.py
```

The prefix (`my_answer_`, `cw_answer_`, etc.) can be anything. The filename **must** keep the `NN_<name>` segment (e.g. `03_permission_manager`) so the test runner can find the right test suite.

### 3. Implement it

Fill in the `raise NotImplementedError` stubs in your answer file. Keep the function/class signatures identical to the stub.

### 4. Run tests against your answer

Use `run_tests.sh` with `-f` (your answer file) and `-c` (the pytest command to run):

```bash
# Run the full test suite for your answer
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py -v

# Run a single test class
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole

# Run a single test method
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py::TestCreateRole::test_empty_permissions_by_default

# Stop on first failure
./run_tests.sh \
  -f python/practice_problem_answers/my_answer_03_permission_manager.py \
  -c pytest python/tests/test_problem_03_permission_manager.py -x
```

The `-c` flag accepts any valid pytest command — flags, node IDs, and extra options all pass through unchanged. `--answer` is injected automatically, so you never need to edit the test files.

## Adding new problems with an AI agent

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

You can also ask it to review and improve existing problems or test suites.

---

## Adding problems manually

Follow the conventions in `CLAUDE.md`. Key rules:

- Problem stubs live in `practice_problems/` and import from nowhere (pure Python).
- Test files always import from `practice_problems.problem_NN_<name>` — never from the answer directory directly. The `--answer` flag handles the swap at runtime.
- Tests are ordered Part 1 → Part 2 → Part 3. Each Part's tests only call functions defined in that Part or earlier.
- Every test uses a pytest fixture — no inline `ClassName()` construction inside test methods.
