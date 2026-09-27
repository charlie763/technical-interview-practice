# =============================================================================
# INTERVIEW PROBLEM 2: Tiered API Rate Limiter
# Difficulty: Senior Software Engineer | Estimated time: 40 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the rate-limiting layer for a developer-facing API platform
# (think: any product where external developers call your API and pay for
# usage tiers).
#
# Each API key belongs to a plan with per-minute and per-day request caps.
# Rate limiting uses a sliding window: a request at time T is within the
# per-minute cap if fewer than `rpm` requests occurred in the window
# (T-60, T], and within the per-day cap if fewer than `rpd` requests occurred
# in (T-86400, T].
#
# All timestamps are Unix time (numeric seconds).
#
# You choose the internal data structures — the public interface below is
# what matters. Store all state in instance variables initialized in
# `initialize`. Class-level variables will bleed between tests and between
# instances — avoid them.
#
# DATA MODEL
# ----------
# Plan limits look like:
#   {
#     "free"       => { rpm: 60,    rpd: 1_000 },
#     "starter"    => { rpm: 300,   rpd: 25_000 },
#     "pro"        => { rpm: 1_000, rpd: 200_000 },
#     "enterprise" => { rpm: nil,   rpd: nil },   # nil = unlimited
#   }
#
# Each API key is conceptually:
#   {
#     id:          String,
#     owner:       String,
#     plan:        String,           # must be a key in the plans hash
#     enabled:     true | false,
#     request_log: [Numeric, ...],   # timestamps of recent requests;
#                                    # entries older than 25h may be pruned
#                                    # at any time
#   }
#
# Example
# -------
#   gw = ApiGateway.new
#   gw.create_key("key_abc", "alice", "pro")
#   gw.is_allowed("key_abc", 1_700_000_000)          # -> true
#   gw.handle_request("key_abc", 1_700_000_000)       # -> { allowed: true, key_id: "key_abc" }
# =============================================================================

class ApiGateway
  DEFAULT_PLANS = {
    "free" => { rpm: 60, rpd: 1_000 },
    "starter" => { rpm: 300, rpd: 25_000 },
    "pro" => { rpm: 1_000, rpd: 200_000 },
    "enterprise" => { rpm: nil, rpd: nil },
  }.freeze

  # plans: optional hash of plan_name => {rpm:, rpd:}. Uses DEFAULT_PLANS if omitted.
  def initialize(plans = nil)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Key management (~10 min)
  # ---------------------------------------------------------------------------

  # Register a new API key with enabled: true and an empty request_log.
  # Raise ArgumentError if key_id already exists or plan is not a known plan.
  def create_key(key_id, owner, plan)
    raise NotImplementedError
  end

  # Set enabled: false on the key. Raise KeyError if key_id not found.
  # (Keys are disabled rather than deleted so historical logs are preserved.)
  def revoke_key(key_id)
    raise NotImplementedError
  end

  # Change a key's plan. Raise KeyError if key_id not found.
  # Raise ArgumentError if new_plan is not a known plan.
  # The request_log is preserved (no reset on plan change).
  def update_plan(key_id, new_plan)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Sliding-window rate check (~15 min)
  # ---------------------------------------------------------------------------

  # Return the number of entries in request_log that fall within the
  # half-open window (now - window_seconds, now].
  #
  # Stateless helper — feel free to use this in Parts 2 and 3.
  def count_in_window(request_log, now, window_seconds)
    raise NotImplementedError
  end

  # Return true if the key is allowed to make a request at time `now`.
  #
  # A request is NOT allowed if any of the following:
  #   - key_id doesn't exist
  #   - the key is disabled
  #   - the per-minute sliding window is at or above the plan's rpm cap
  #   - the per-day sliding window is at or above the plan's rpd cap
  #
  # A nil limit means that dimension is unlimited.
  def is_allowed(key_id, now)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Recording requests + pruning (~5 min)
  # ---------------------------------------------------------------------------

  # Append `now` to the key's request_log and prune any entries older than
  # 25 hours (90_000 seconds) to bound memory usage.
  #
  # Raise KeyError if key_id not found.
  # Note: call this only AFTER confirming the request is allowed.
  def record_request(key_id, now)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 4 — Combined handler + usage stats (~10 min)
  # ---------------------------------------------------------------------------

  # Attempt to process a request. Return a result hash:
  #
  # On success (request allowed):
  #   { allowed: true, key_id: key_id }
  #   Side effect: records the request.
  #
  # On failure:
  #   { allowed: false, reason: <one of the symbols below> }
  #   No side effects.
  #
  # Reason symbols:
  #   :key_not_found — key_id not registered
  #   :key_disabled  — key exists but enabled is false
  #   :rpm_exceeded  — per-minute cap hit
  #   :rpd_exceeded  — per-day cap hit (only checked if rpm is OK)
  def handle_request(key_id, now)
    raise NotImplementedError
  end

  # Return current usage stats for a key:
  #   {
  #     key_id:    String,
  #     plan:      String,
  #     rpm_used:  Integer,        # requests in the last 60 seconds
  #     rpm_limit: Integer | nil,  # nil if unlimited
  #     rpd_used:  Integer,        # requests in the last 24 hours
  #     rpd_limit: Integer | nil,
  #   }
  # Raise KeyError if key_id not found.
  def get_usage(key_id, now)
    raise NotImplementedError
  end
end
