"""
=============================================================================
INTERVIEW PROBLEM 2: Tiered API Rate Limiter
Difficulty: Senior Software Engineer | Estimated time: 40 min
=============================================================================

CONTEXT
-------
You're building the rate-limiting layer for a developer-facing API platform
(think: Stripe, Vercel, Twilio — any product where external developers call
your API and pay for usage tiers).

Each API key belongs to a plan with per-minute and per-day request caps.
Rate limiting uses a sliding window: a request at time T is within the
per-minute cap if fewer than `rpm` requests occurred in the window [T-60, T],
and within the per-day cap if fewer than `rpd` requests occurred in [T-86400, T].

All timestamps are Unix time (float seconds).

DATA MODEL
----------

Plan limits are stored in state["plans"]:
  {
    "free":       {"rpm": 60,    "rpd": 1_000},
    "starter":    {"rpm": 300,   "rpd": 25_000},
    "pro":        {"rpm": 1_000, "rpd": 200_000},
    "enterprise": {"rpm": None,  "rpd": None},   # None = unlimited
  }

Each API key entry in state["keys"] looks like:
  {
    "id":          str,
    "owner":       str,
    "plan":        str,          # must be a key in state["plans"]
    "enabled":     bool,
    "request_log": [float, ...], # sorted list of timestamps of recent requests
                                 # entries older than 24h may be pruned at any time
  }

TrackerState:
  {
    "plans": { plan_name: {"rpm": int | None, "rpd": int | None} },
    "keys":  { key_id: ApiKey },
  }
=============================================================================
"""

from typing import Optional


DEFAULT_PLANS = {
    "free":       {"rpm": 60,    "rpd": 1_000},
    "starter":    {"rpm": 300,   "rpd": 25_000},
    "pro":        {"rpm": 1_000, "rpd": 200_000},
    "enterprise": {"rpm": None,  "rpd": None},
}


def make_gateway(plans: Optional[dict] = None) -> dict:
    """Return a fresh gateway state. Uses DEFAULT_PLANS if none provided."""
    return {
        "plans": plans if plans is not None else DEFAULT_PLANS,
        "keys": {},
    }


# ---------------------------------------------------------------------------
# PART 1 — Key management  (~10 min)
# ---------------------------------------------------------------------------

def create_key(state: dict, key_id: str, owner: str, plan: str) -> dict:
    """
    Register a new API key. Return the key dict.
    Raise ValueError if key_id already exists or plan is not in state["plans"].
    """
    try:
      if plan not in state["plans"]:
        raise ValueError
    except ValueError as error:
      print(f'Plan does not exist: {error}')
      raise
    try:
      if state["keys"].get(key_id):
        raise ValueError
    except ValueError as error:
      print(f'Key already exists: {error}')
      raise

    api_key = {
      "id": key_id,
      "owner": owner,
      "plan": plan,
      "enabled": True,
      "request_log": []
    }

    state["keys"][key_id] = api_key
    return api_key

def revoke_key(state: dict, key_id: str) -> None:
    """
    Set key["enabled"] = False. Raise KeyError if key_id not found.
    (Keys are disabled rather than deleted so historical logs are preserved.)
    """
    try:
      if not state["keys"].get(key_id):
        raise KeyError
    except KeyError as error:
      raise

    state["keys"][key_id]["enabled"] = False
    return None


def update_plan(state: dict, key_id: str, new_plan: str) -> None:
    """
    Change a key's plan. Raise KeyError if key_id not found.
    Raise ValueError if new_plan is not in state["plans"].
    The request_log is preserved (no reset on plan change).
    """
    try:
      if new_plan not in state["plans"]:
        raise ValueError
    except ValueError as error:
      print(f'Plan does not exist: {error}')
      raise
    try:
      if not state["keys"].get(key_id):
        raise KeyError
    except KeyError as error:
      print(f'Key does not exist: {error}')
      raise

    state["keys"][key_id]["plan"] = new_plan


# ---------------------------------------------------------------------------
# PART 2 — Sliding-window rate check  (~15 min)
# ---------------------------------------------------------------------------

def _count_in_window(request_log: list, now: float, window_seconds: int) -> int:
    """
    Return the number of entries in request_log that fall within
    the half-open window (now - window_seconds, now].

    Helper — feel free to use this in Parts 2 and 3, or inline it.
    """
    window_end = now - window_seconds
    return sum(1 for request in request_log if
        (window_end < request and
        request <= now)
    )

def is_allowed(state: dict, key_id: str, now: float) -> bool:
    """
    Return True if the key is allowed to make a request at time `now`.

    A request is NOT allowed if any of the following:
      - key_id doesn't exist in state["keys"]
      - key["enabled"] is False
      - the per-minute sliding window is at or above the plan's rpm cap
      - the per-day sliding window is at or above the plan's rpd cap

    A None limit means that dimension is unlimited.
    """
    if not state["keys"].get(key_id):
      return False
    if state["keys"][key_id]["enabled"] == False:
      return False

    request_log = state["keys"][key_id]["request_log"]
    plan_name = state["keys"][key_id]["plan"]
    plan_rpm = state["plans"][plan_name]["rpm"]
    plan_rpd = state["plans"][plan_name]["rpd"]

    if plan_rpm is None or plan_rpd is None:
      return True
    if _count_in_window(request_log, now, 60) >= plan_rpm:
      return False
    if _count_in_window(request_log, now, 86_400) >= plan_rpd:
      return False
    return True


# ---------------------------------------------------------------------------
# PART 3 — Recording requests + pruning  (~5 min)
# ---------------------------------------------------------------------------

def record_request(state: dict, key_id: str, now: float) -> None:
    """
    Append `now` to key["request_log"] and prune any entries older than
    25 hours (90_000 seconds) to bound memory usage.

    Raise KeyError if key_id not found.
    Note: call this only AFTER confirming the request is allowed.
    """
    try:
      if not state["keys"].get(key_id):
        raise KeyError
    except KeyError as error:
      print(f'Key does not exist: {error}')
      raise

    out_of_date_time = now - 90_000
    request_log = state["keys"][key_id]["request_log"]
    state["keys"][key_id]["request_log"] = (
      [request for request in request_log if request >= out_of_date_time]
    )
    state["keys"][key_id]["request_log"].append(now)
    return None

# ---------------------------------------------------------------------------
# PART 4 — Combined handler + usage stats  (~10 min)
# ---------------------------------------------------------------------------

def handle_request(state: dict, key_id: str, now: float) -> dict:
    """
    Attempt to process a request. Return a result dict:

    On success (request allowed):
      {"allowed": True, "key_id": key_id}
      Side effect: records the request.

    On failure:
      {"allowed": False, "reason": <one of the strings below>}
      No side effects.

    Reason strings:
      "key_not_found"   — key_id not in state["keys"]
      "key_disabled"    — key exists but enabled=False
      "rpm_exceeded"    — per-minute cap hit
      "rpd_exceeded"    — per-day cap hit (only if rpm is OK)
    """
    raise NotImplementedError


def get_usage(state: dict, key_id: str, now: float) -> dict:
    """
    Return current usage stats for a key:
    {
      "key_id":    str,
      "plan":      str,
      "rpm_used":  int,          # requests in the last 60 seconds
      "rpm_limit": int | None,   # None if unlimited
      "rpd_used":  int,          # requests in the last 24 hours
      "rpd_limit": int | None,
    }
    Raise KeyError if key_id not found.
    """
    raise NotImplementedError
