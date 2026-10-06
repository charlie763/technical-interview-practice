# Ruby — Agent Guidelines

## Prerequisites

- Ruby 3.2+ installed (`ruby --version`)
- `rspec` gem available (`gem install rspec` if `rspec --version` fails)

## Workflow per problem

1. The problem stub is in `practice_problems/problem_NN_<name>.rb` — **do not edit it**.
2. Copy the stub to `practice_problem_answers/` and implement it there:

   ```bash
   cp practice_problems/problem_01_geofence_alert_engine.rb \
      practice_problem_answers/cw_answer_01_geofence_alert_engine.rb
   ```

3. Run tests against your answer via `run_tests.sh` from the repo root:

   ```bash
   ./run_tests.sh \
     -f ruby/practice_problem_answers/cw_answer_01_geofence_alert_engine.rb \
     -c rspec tests/problem_01_geofence_alert_engine_spec.rb
   ```

   Or directly from `ruby/` with the env var:

   ```bash
   cd ruby
   PRACTICE_ANSWER=cw_answer_01_geofence_alert_engine \
     rspec tests/problem_01_geofence_alert_engine_spec.rb
   ```

When `PRACTICE_ANSWER` is set, `tests/spec_helper.rb`'s `load_problem` helper loads
your answer file instead of the problem stub — no changes to spec files are ever
needed.

Running `rspec` from `ruby/` (without `PRACTICE_ANSWER`) runs all specs against the
stubs, which should all fail with `NotImplementedError`. This is the expected baseline.

## Test commands

| Command | Problem |
|---|---|
| `rspec tests/problem_01_geofence_alert_engine_spec.rb` | Problem 01 — Geofence Alert Engine |
| `rspec tests/problem_02_api_rate_limiter_spec.rb` | Problem 02 — Tiered API Rate Limiter |
| `rspec tests/problem_03_permission_manager_spec.rb` | Problem 03 — Permission Manager (RBAC) |
| `rspec tests/problem_04_biomarker_alert_monitor_spec.rb` | Problem 04 — Biomarker Alert Monitor |
| `rspec tests/problem_05_medication_titration_tracker_spec.rb` | Problem 05 — Medication Titration Tracker |
| `rspec tests/problem_06_lab_cadence_monitor_spec.rb` | Problem 06 — Lab Cadence Compliance Monitor |
| `rspec tests/problem_07_care_team_assignments_spec.rb` | Problem 07 — Care Team Assignment Manager |
| `rspec tests/problem_08_enrollment_pipeline_spec.rb` | Problem 08 — Patient Enrollment Pipeline |
| `rspec tests/problem_09_incident_aggregator_spec.rb` | Problem 09 — Multi-Source Incident Aggregator |
| `rspec tests/problem_10_dispatch_manager_spec.rb` | Problem 10 — Responder Dispatch Manager |
| `rspec tests/problem_11_coverage_tracker_spec.rb` | Problem 11 — Sensor Coverage Tracker |
| `rspec tests/problem_12_contract_alert_scheduler_spec.rb` | Problem 12 — Contract Expiration Alert Scheduler |
| `rspec tests/problem_13_contract_lifecycle_spec.rb` | Problem 13 — Contract Lifecycle State Machine |
| `rspec tests/problem_14_contract_amendment_spec.rb` | Problem 14 — Contract Amendment Manager |
| `rspec tests/problem_15_tic_tac_toe_engine_spec.rb` | Problem 15 — Tic-Tac-Toe Engine |
| `rspec tests/problem_16_premium_rating_engine_spec.rb` | Problem 16 — Policy Premium Rating Engine |
| `rspec tests/problem_17_claims_pipeline_spec.rb` | Problem 17 — Claims Processing Pipeline |

Prefix with `PRACTICE_ANSWER=cw_answer_NN_<name>` to test your implementation.

---

## Problem design rules

### Parts must be self-contained
- Tests for Part N must only call methods defined in Parts 1–N.
- Never verify a Part 1 result using a Part 2 helper.

### Methods must compose
Design later-part methods to **call** earlier-part methods rather than duplicating
logic. If a Part 1 method is later needed by a Part 3 method with richer requirements,
extract a shared private helper both can call rather than re-implementing the check.

### Class-based problems
All problems use a single class with instance variables initialised in `initialize`.
The candidate chooses internal data structures; the public interface is what matters.
State must never live in class-level (`@@class_variable` or class-level constants that
get mutated) variables — they bleed across test instances.

### Stubs raise `NotImplementedError`
Every stub method body is `raise NotImplementedError`. Running specs against the stub
→ all examples fail. Running specs against a correct implementation → all examples pass.

### Use idiomatic Ruby errors
- `ArgumentError` — invalid input (bad enum value, duplicate id, invalid state transition)
- `KeyError` — lookup of a missing id (raise via `Hash#fetch` with a block, or
  `raise KeyError, "..."` directly)

---

## Test writing rules

### Use `let`, never construct instances inline in the example body
```ruby
# BAD
it "does something" do
  tracker = GeofenceTracker.new
  # ...
end

# GOOD
let(:tracker) { described_class.new }
it "does something" do
  tracker.open_ticket(...)
end
```

### Pre-seeded `let` overrides for Parts 2+
Redefine `let(:tracker)` inside a nested `describe` block for Parts 2 and 3 to build a
realistic pre-seeded instance, so candidates don't need to re-set-up data in every example.

### Cross-part independence
A Part 1 example must never fail because Part 2 is not implemented. If a Part 1
assertion requires a method from Part 2, move that method to Part 1.

### Use unique identifiers per example
When multiple examples create tickets/records with string ids, give each a distinct id
(e.g. `"part1_open"`, `"part1_dup"`) rather than a shared generic name like `"t1"`. This
avoids false failures if an implementation accidentally leaks state across instances.

### Seeding or asserting on private state: use `instance_variable_get`, not a public method from a later Part
When a problem's data model needs pre-seeded state that only becomes reachable through
a later Part's method (e.g. seeding a request log that Part 3's `record_request` would
normally populate), reach into the instance directly instead of calling that later-Part
method:

```ruby
gateway.instance_variable_get(:@keys)["key_abc"][:request_log] = [...]
```

This mirrors the white-box access Python/Go tests get for free (plain dicts / same-package
structs) and avoids Part 2 examples depending on Part 3 being implemented. Put a shared
`seed_*` helper at the top of the spec file rather than repeating the `instance_variable_get`
call inline everywhere.

### Shared fixtures live in spec_helper.rb
`BASE_TIME` (an arbitrary fixed Unix timestamp) is defined once in `tests/spec_helper.rb`
and used by every spec file. Don't redefine it per-file — Ruby warns on reassigning an
already-initialized top-level constant when the full suite runs.

### Test ordering = implementation ordering
Order `describe`/`it` blocks to match Part order. A candidate who finishes Part 1 and
runs the full spec file should see only Part 1 examples passing, with Part 2/3 failing
cleanly due to `NotImplementedError` — not due to test coupling.
