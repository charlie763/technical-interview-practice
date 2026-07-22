# Technical Interview Practice

A collection of Senior Software Engineer–level practice problems with test suites, designed to be generated and extended by an AI coding agent (Claude Code or similar).

---

## Repo structure

```
python/
  practice_problems/          # Problem stubs — read these, don't edit them
  practice_problem_answers/   # Your implementations go here
  tests/                      # pytest suites (one per problem)
  conftest.py                 # pytest config & --answer flag
react/
  practice_problems/          # React/JSX starter files
run_tests.sh                  # Universal test runner (see below)
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

## Practicing a problem

### 1. Read the problem stub

Open the problem file in `python/practice_problems/` and read the docstring — it explains the data model, the parts, and what each function should do.

### 2. Copy it to your answers directory

```bash
cp python/practice_problems/problem_03_permission_manager.py \
   python/practice_problem_answers/my_answer_03_permission_manager.py
```

The prefix (`my_answer_`, `cw_answer_`, etc.) can be anything. The filename **must** keep the `NN_<name>` segment (e.g. `03_permission_manager`) so the test runner can find the right test suite.

### 3. Implement it

Fill in the `raise NotImplementedError` stubs in your answer file. Keep the function/class signatures identical to the stub.

### 4. Run tests against your answer

```bash
./run_tests.sh python/practice_problem_answers/my_answer_03_permission_manager.py
```

Pass extra pytest flags after the filename if you want:

```bash
# Stop on first failure
./run_tests.sh python/practice_problem_answers/my_answer_03_permission_manager.py -x

# Run only Part 1 tests
./run_tests.sh python/practice_problem_answers/my_answer_03_permission_manager.py -k "TestCreate or TestGrant or TestAssign or TestHas"
```

### 5. Run tests against the stub (sanity check)

Running the test suite without `--answer` tests the stub, where everything should raise `NotImplementedError`:

```bash
cd python && pytest tests/test_problem_03_permission_manager.py -v
```

---

## How `run_tests.sh` works

The script detects the language from the file extension, locates the matching test file by problem number, then runs the appropriate test framework with your answer injected:

- **Python (`.py`)** — runs `pytest --answer <file>`. The `conftest.py` loads your answer file and injects it into Python's module registry *before* test collection, so the test suite's imports resolve to your code instead of the stub. No manual import changes needed.
- **JavaScript/TypeScript** — coming soon (jest/vitest).

---

## Adding new problems with an AI agent

This repo is designed to be extended by prompting an AI coding agent. Open Claude Code (or your preferred agent) in this repo and describe what you want:

```
Generate a new Senior SWE interview problem for the python/ directory.
Follow the conventions in CLAUDE.md — class-based data structure,
4 progressive parts, full pytest suite.
Theme: something broadly applicable to SaaS backends (not IoT-specific).
```

The agent will read `CLAUDE.md` for conventions and create:
- `python/practice_problems/problem_NN_<name>.py` — the stub
- `python/tests/test_problem_NN_<name>.py` — the test suite

You can also ask it to review and improve existing problems or test suites.

---

## Adding problems manually

Follow the conventions in `CLAUDE.md`. Key rules:

- Problem stubs live in `practice_problems/` and import from nowhere (pure Python).
- Test files always import from `practice_problems.problem_NN_<name>` — never from the answer directory directly. The `--answer` flag handles the swap at runtime.
- Tests are ordered Part 1 → Part 2 → Part 3. Each Part's tests only call functions defined in that Part or earlier.
- Every test uses a pytest fixture — no inline `ClassName()` construction inside test methods.
