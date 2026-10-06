# Loads either the problem stub or a candidate's answer file, selected via the
# PRACTICE_ANSWER env var, so spec files never need to change their require path.
#
#   PRACTICE_ANSWER=cw_answer_01_ticket_escalation_tracker rspec tests/problem_01_ticket_escalation_tracker_spec.rb
#
# Running rspec with no PRACTICE_ANSWER set loads the stub, which should make
# every test fail with NotImplementedError — the expected baseline.

ROOT_DIR = File.expand_path("..", __dir__)

# Arbitrary fixed "now" shared by every spec file, for deterministic timestamp math.
BASE_TIME = 1_700_000_000.0

# Shared ISO-8601 timestamps (naive, lexicographically sortable) used by any
# spec file that needs a handful of deterministic points in time.
# T0 = base, T1 = T0+60s, T2 = T0+120s, T3 = T0+300s, T4 = T0+600s
T0 = "2024-06-01T10:00:00".freeze
T1 = "2024-06-01T10:01:00".freeze
T2 = "2024-06-01T10:02:00".freeze
T3 = "2024-06-01T10:05:00".freeze
T4 = "2024-06-01T10:10:00".freeze

def load_problem(problem_id)
  answer = ENV["PRACTICE_ANSWER"]
  path = if answer
    File.join(ROOT_DIR, "practice_problem_answers", "#{answer}.rb")
  else
    File.join(ROOT_DIR, "practice_problems", "problem_#{problem_id}.rb")
  end

  raise "file not found: #{path}" unless File.file?(path)

  load path
end

RSpec.configure do |config|
  config.expect_with :rspec do |c|
    c.syntax = :expect
  end
end
