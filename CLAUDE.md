# Technical Interview Practice — Agent Guidelines

## Repo structure

```
python/
  practice_problems/          # problem stubs (read-only during practice)
  practice_problem_answers/   # cw_answer_XX_... files (filled in by Charlie)
  tests/                      # pytest suites
react/
  practice_problems/          # JSX/TSX starter files
```

**Workflow per problem:**
1. Problem file has the prompt + empty stubs.
2. User copies the problem file to `practice_problem_answers/cw_answer_XX_<name>.py` and implements it there.
3. He updates the test file's import to point at his answer file.
4. He runs `pytest tests/test_problem_XX_<name>.py -v` from the `python/` directory.

Test files **always** import from `practice_problems.problem_NN_<name>` (the stub).
The `--answer` flag in `python/conftest.py` injects the answer module under that same
module path before test collection, so no manual import changes are ever needed.

---

## Problem design rules

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
moment. Signal this explicitly in the stub docstring:

```python
def get_outreach_list(self, min_consecutive_days: int = 3) -> list[dict]:
    """
    ...
    Hint: max_consecutive_out_of_range_days can be helpful here.
    """
```

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
