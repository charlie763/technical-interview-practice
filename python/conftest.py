import sys
import os
import re
import importlib.util
import pytest

# Allow test files to import from practice_problems/ and practice_problem_answers/
sys.path.insert(0, os.path.dirname(__file__))


def pytest_addoption(parser):
    parser.addoption(
        "--answer",
        metavar="FILE",
        default=None,
        help=(
            "Path to an answer file to test against instead of the problem stub. "
            "The filename must contain a segment matching NN_<name> "
            "(e.g. cw_answer_03_permission_manager.py). "
            "The corresponding stub module is replaced in sys.modules before "
            "test collection so all imports resolve to your implementation."
        ),
    )


def pytest_configure(config):
    try:
        answer_path = config.getoption("--answer")
    except (ValueError, AttributeError):
        return
    if not answer_path:
        return

    answer_path = os.path.abspath(answer_path)
    if not os.path.isfile(answer_path):
        raise FileNotFoundError(f"--answer: file not found: {answer_path}")

    # Derive the stub module name from the answer filename.
    # Convention: filename contains a segment "NN_<name>"
    # e.g. cw_answer_03_permission_manager.py → problem_03_permission_manager
    #      → injected under practice_problems.problem_03_permission_manager
    filename = os.path.basename(answer_path)
    match = re.search(r'(\d{2}_[a-z_]+)\.py$', filename)
    if not match:
        raise ValueError(
            f"--answer: '{filename}' does not match the expected naming pattern.\n"
            f"Filename must contain a segment like '03_permission_manager'.\n"
            f"Example: cw_answer_03_permission_manager.py"
        )

    stub_module = f"practice_problems.problem_{match.group(1)}"
    spec = importlib.util.spec_from_file_location(stub_module, answer_path)
    module = importlib.util.module_from_spec(spec)
    sys.modules[stub_module] = module
    spec.loader.exec_module(module)


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
