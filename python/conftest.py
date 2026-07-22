import sys
import os
import pytest

# Allow test files to import from practice_problems/ and practice_problem_answers/
sys.path.insert(0, os.path.dirname(__file__))


@pytest.fixture(autouse=True)
def _reset_class_level_state():
    """
    After each test, clear class-level mutable set state in any imported
    practice-problem module.

    A correct implementation stores all state in instance variables, making
    this fixture a no-op. If removing this fixture causes test bleed, there
    is a bug in the implementation — either:
      (a) a class-level set used as instance state:
              class Foo:
                  SEEN_IDS = set()   # shared across ALL instances
      (b) a mutable default argument in __init__:
              def __init__(self, roles=set())  # all instances share one set

    This is a safety net for practice sessions, not a substitute for fixing
    the underlying bug.
    """
    yield  # test runs here
    for name, module in list(sys.modules.items()):
        if module is None or "practice_problem" not in name:
            continue
        for attr_name in dir(module):
            cls = getattr(module, attr_name, None)
            if not isinstance(cls, type):
                continue
            # Clear class-level set attributes
            for val in vars(cls).values():
                if isinstance(val, set):
                    val.clear()
            # Reset mutable default set arguments in __init__
            init_fn = getattr(cls, "__init__", None)
            for default in getattr(init_fn, "__defaults__", None) or ():
                if isinstance(default, set):
                    default.clear()
