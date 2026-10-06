load_problem("13_contract_lifecycle")

# Shared timestamps (namespaced to this problem to avoid clashing with
# other spec files' own timestamp constants).
# LC_T0 = base datetime, LC_T1 = +1 day, LC_T2 = +5 days, LC_T3 = +10 days,
# LC_T4 = +40 days (> 30 days for overdue tests), LC_T5 = +45 days
LC_T0 = "2025-01-01T09:00:00".freeze
LC_T1 = "2025-01-02T10:00:00".freeze
LC_T2 = "2025-01-06T11:00:00".freeze
LC_T3 = "2025-01-11T12:00:00".freeze
LC_T4 = "2025-02-10T09:00:00".freeze # 40 days after LC_T0
LC_T5 = "2025-02-15T09:00:00".freeze # 45 days after LC_T0

RSpec.describe ContractLifecycle do
  let(:cl) { described_class.new }

  # Pre-seeded lifecycle manager:
  #   c-seed-1  "Vendor MSA"    state=approved    (draft->in_review->approved)
  #   c-seed-2  "NDA Agreement" state=draft
  #   c-seed-3  "SaaS License"  state=terminated  (draft->in_review->approved->executed->active->terminated)
  let(:seeded_cl) do
    c = described_class.new
    # c-seed-1: draft -> in_review -> approved
    c.create_contract("c-seed-1", "Vendor MSA", LC_T0, "alice")
    c.transition("c-seed-1", "in_review", LC_T1, "alice")
    c.transition("c-seed-1", "approved", LC_T2, "bob")
    # c-seed-2: stays draft
    c.create_contract("c-seed-2", "NDA Agreement", LC_T0, "alice")
    # c-seed-3: full path to terminated
    c.create_contract("c-seed-3", "SaaS License", LC_T0, "carol")
    c.transition("c-seed-3", "in_review", LC_T1, "carol")
    c.transition("c-seed-3", "approved", LC_T2, "bob")
    c.transition("c-seed-3", "executed", LC_T3, "carol")
    c.transition("c-seed-3", "active", LC_T4, "carol")
    c.transition("c-seed-3", "terminated", LC_T5, "carol")
    c
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Contract creation, field management, transitions
  # ---------------------------------------------------------------------------
  describe "#create_contract" do
    it "returns a contract in draft" do
      c = cl.create_contract("c-cr-1", "Title", LC_T0, "alice")
      expect(c[:contract_id]).to eq("c-cr-1")
      expect(c[:state]).to eq("draft")
      expect(c[:title]).to eq("Title")
      expect(c[:fields]).to eq({})
    end

    it "stores created_at" do
      c = cl.create_contract("c-cr-ts", "Title", LC_T0, "alice")
      expect(c[:created_at]).to eq(LC_T0)
    end

    it "raises ArgumentError on a duplicate" do
      cl.create_contract("c-dup-lc", "A", LC_T0, "alice")
      expect { cl.create_contract("c-dup-lc", "B", LC_T1, "alice") }.to raise_error(ArgumentError)
    end

    it "records the initial audit entry" do
      cl.create_contract("c-audit-init", "T", LC_T0, "alice")
      trail = cl.get_audit_trail("c-audit-init")
      expect(trail.size).to eq(1)
      expect(trail.first[:from_state]).to be_nil
      expect(trail.first[:to_state]).to eq("draft")
      expect(trail.first[:actor]).to eq("alice")
    end
  end

  describe "#set_field" do
    it "sets a field" do
      seeded_cl.set_field("c-seed-1", "value", 100_000)
      expect(seeded_cl.get_contract("c-seed-1")[:fields]["value"]).to eq(100_000)
    end

    it "updates an existing field" do
      seeded_cl.set_field("c-seed-1", "value", 50_000)
      seeded_cl.set_field("c-seed-1", "value", 75_000)
      expect(seeded_cl.get_contract("c-seed-1")[:fields]["value"]).to eq(75_000)
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_cl.set_field("no-such", "key", "val") }.to raise_error(KeyError)
    end
  end

  describe "#get_contract" do
    it "returns the contract" do
      c = seeded_cl.get_contract("c-seed-1")
      expect(c[:contract_id]).to eq("c-seed-1")
      expect(c[:state]).to eq("approved")
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_cl.get_contract("nonexistent") }.to raise_error(KeyError)
    end
  end

  describe "#transition" do
    it "updates the state on a valid transition" do
      cl.create_contract("c-tr-1", "T", LC_T0, "alice")
      cl.transition("c-tr-1", "in_review", LC_T1, "alice")
      expect(cl.get_contract("c-tr-1")[:state]).to eq("in_review")
    end

    it "raises ArgumentError on an invalid transition" do
      cl.create_contract("c-tr-inv", "T", LC_T0, "alice")
      # draft -> approved is invalid
      expect { cl.transition("c-tr-inv", "approved", LC_T1, "alice") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError when the contract is in a terminal state" do
      # terminated is terminal
      expect { seeded_cl.transition("c-seed-3", "draft", LC_T5, "alice") }.to raise_error(ArgumentError)
    end

    it "allows draft -> in_review -> draft" do
      cl.create_contract("c-back", "T", LC_T0, "alice")
      cl.transition("c-back", "in_review", LC_T1, "alice")
      cl.transition("c-back", "draft", LC_T2, "alice")
      expect(cl.get_contract("c-back")[:state]).to eq("draft")
    end

    it "appends an audit entry" do
      cl.create_contract("c-tr-audit", "T", LC_T0, "alice")
      cl.transition("c-tr-audit", "in_review", LC_T1, "bob")
      trail = cl.get_audit_trail("c-tr-audit")
      expect(trail.size).to eq(2)
      last = trail.last
      expect(last[:from_state]).to eq("draft")
      expect(last[:to_state]).to eq("in_review")
      expect(last[:actor]).to eq("bob")
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_cl.transition("no-such", "in_review", LC_T1, "alice") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Audit trail, by-state query, bulk advance
  # ---------------------------------------------------------------------------
  describe "#get_audit_trail" do
    it "returns all entries in order" do
      trail = seeded_cl.get_audit_trail("c-seed-1")
      # create(draft) + in_review + approved = 3 entries
      expect(trail.size).to eq(3)
      expect(trail.map { |e| e[:to_state] }).to eq(%w[draft in_review approved])
    end

    it "returns the full trail for a terminal contract" do
      trail = seeded_cl.get_audit_trail("c-seed-3")
      expect(trail.map { |e| e[:to_state] }).to eq(%w[draft in_review approved executed active terminated])
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_cl.get_audit_trail("no-such") }.to raise_error(KeyError)
    end
  end

  describe "#get_contracts_by_state" do
    it "returns the contracts in that state" do
      approved = seeded_cl.get_contracts_by_state("approved")
      ids = approved.map { |c| c[:contract_id] }
      expect(ids).to include("c-seed-1")
      expect(ids).not_to include("c-seed-2")
    end

    it "is sorted by contract_id" do
      cl.create_contract("c-z", "Z", LC_T0, "a")
      cl.create_contract("c-a", "A", LC_T0, "a")
      cl.create_contract("c-m", "M", LC_T0, "a")
      ids = cl.get_contracts_by_state("draft").map { |c| c[:contract_id] }
      expect(ids).to eq(ids.sort)
    end

    it "is empty for an unused state" do
      expect(seeded_cl.get_contracts_by_state("expired")).to eq([])
    end
  end

  describe "#bulk_advance" do
    it "succeeds for every contract" do
      3.times { |i| cl.create_contract("c-bulk-#{i}", "Contract #{i}", LC_T0, "alice") }
      result = cl.bulk_advance(%w[c-bulk-0 c-bulk-1 c-bulk-2], "in_review", LC_T1, "alice")
      expect(result[:succeeded].size).to eq(3)
      expect(result[:failed].size).to eq(0)
    end

    it "continues past a partial failure" do
      # c-ok is draft; c-bad is already in_review (can't go back to draft via bulk)
      cl.create_contract("c-ok", "OK", LC_T0, "alice")
      cl.create_contract("c-bad", "Bad", LC_T0, "alice")
      cl.transition("c-bad", "in_review", LC_T1, "alice")
      result = cl.bulk_advance(%w[c-ok c-bad], "in_review", LC_T2, "alice")
      expect(result[:succeeded]).to include("c-ok")
      expect(result[:failed].any? { |f| f[:contract_id] == "c-bad" }).to be true
    end

    it "updates state for every success" do
      cl.create_contract("c-bs-1", "A", LC_T0, "alice")
      cl.create_contract("c-bs-2", "B", LC_T0, "alice")
      cl.bulk_advance(%w[c-bs-1 c-bs-2], "in_review", LC_T1, "alice")
      expect(cl.get_contract("c-bs-1")[:state]).to eq("in_review")
      expect(cl.get_contract("c-bs-2")[:state]).to eq("in_review")
    end

    it "includes a reason for each failure" do
      cl.create_contract("c-fail-r", "X", LC_T0, "alice")
      result = cl.bulk_advance(["c-fail-r"], "approved", LC_T1, "alice")
      expect(result[:failed].size).to eq(1)
      expect(result[:failed].first[:reason]).not_to eq("")
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Lifecycle metrics and overdue contracts
  # ---------------------------------------------------------------------------
  describe "#get_lifecycle_metrics" do
    it "reports the total count" do
      expect(seeded_cl.get_lifecycle_metrics[:total]).to eq(3)
    end

    it "reports per-state counts" do
      by_state = seeded_cl.get_lifecycle_metrics[:by_state]
      expect(by_state["approved"]).to eq(1)
      expect(by_state["draft"]).to eq(1)
      expect(by_state["terminated"]).to eq(1)
    end

    it "only includes states with a nonzero count" do
      seeded_cl.get_lifecycle_metrics[:by_state].each_value do |count|
        expect(count).to be > 0
      end
    end

    it "reports the terminal count" do
      # c-seed-3 is terminated
      expect(seeded_cl.get_lifecycle_metrics[:terminal_count]).to eq(1)
    end

    it "returns zeros for an empty manager" do
      metrics = cl.get_lifecycle_metrics
      expect(metrics[:total]).to eq(0)
      expect(metrics[:by_state]).to eq({})
      expect(metrics[:terminal_count]).to eq(0)
    end
  end

  describe "#get_overdue_contracts" do
    it "returns contracts stuck for more than 30 days" do
      # Create contract that transitions to in_review at LC_T0, then nothing
      cl.create_contract("c-overdue", "Old Contract", LC_T0, "alice")
      cl.transition("c-overdue", "in_review", LC_T0, "alice")
      # LC_T4 = LC_T0 + 40 days -> should be overdue
      ids = cl.get_overdue_contracts(LC_T4).map { |o| o[:contract_id] }
      expect(ids).to include("c-overdue")
    end

    it "does not flag a recent transition" do
      cl.create_contract("c-recent", "New Contract", LC_T0, "alice")
      cl.transition("c-recent", "in_review", LC_T3, "alice")
      # LC_T4 - LC_T3 = 30 days exactly -> not "more than 30 days"
      ids = cl.get_overdue_contracts(LC_T4).map { |o| o[:contract_id] }
      expect(ids).not_to include("c-recent")
    end

    it "excludes terminal contracts" do
      ids = seeded_cl.get_overdue_contracts(LC_T5).map { |o| o[:contract_id] }
      expect(ids).not_to include("c-seed-3") # terminated -> terminal
    end

    it "is sorted by days_stuck descending" do
      cl.create_contract("c-od-a", "A", LC_T0, "alice")
      cl.create_contract("c-od-b", "B", LC_T0, "alice")
      cl.transition("c-od-a", "in_review", LC_T0, "alice")
      cl.transition("c-od-b", "in_review", LC_T1, "alice")
      days = cl.get_overdue_contracts(LC_T4).map { |o| o[:days_stuck] }
      expect(days).to eq(days.sort.reverse)
    end

    it "includes all required fields" do
      cl.create_contract("c-od-f", "Fields Test", LC_T0, "alice")
      cl.transition("c-od-f", "in_review", LC_T0, "alice")
      entry = cl.get_overdue_contracts(LC_T4).find { |o| o[:contract_id] == "c-od-f" }
      expect(entry).to include(:title, :state, :stuck_since, :days_stuck)
      expect(entry[:days_stuck]).to be >= 31
    end
  end
end
