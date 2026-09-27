load_problem("17_claims_pipeline")

# Shared timestamps (namespaced to this problem to avoid clashing with
# other spec files' own timestamp constants).
CLM_D0 = "2025-01-15".freeze # incident date
CLM_T0 = "2025-02-01T09:00:00".freeze # filed
CLM_T1 = "2025-02-03T10:00:00".freeze # investigating
CLM_T2 = "2025-02-10T14:00:00".freeze # evaluation
CLM_T3 = "2025-02-20T11:00:00".freeze # settled / denied
CLM_T4 = "2025-03-01T09:00:00".freeze # closed

RSpec.describe ClaimsPipeline do
  let(:pipeline) { described_class.new }

  # Pre-seeded pipeline:
  #   clm-001  pol-101  epl  claimed=75000  reserve=50000  -> settled (approved=60000)
  #   clm-002  pol-101  epl  claimed=30000  reserve=30000  -> investigating
  #   clm-003  pol-102  do   claimed=200000 reserve=150000 -> evaluation
  #   clm-004  pol-103  epl  claimed=10000  reserve=10000  -> denied
  let(:seeded_pipeline) do
    p = described_class.new

    # clm-001: filed -> investigating -> evaluation -> settled -> closed
    p.file_claim(
      "clm-001", policy_id: "pol-101", coverage_type: "epl",
      incident_date: CLM_D0, filed_at: CLM_T0,
      claimed_amount: 75_000, reserve_amount: 50_000, actor: "adj-1"
    )
    p.advance_status("clm-001", "investigating", at: CLM_T1, actor: "adj-1")
    p.advance_status("clm-001", "evaluation", at: CLM_T2, actor: "adj-1")
    p.settle_claim("clm-001", approved_amount: 60_000, settled_at: CLM_T3, actor: "adj-1")

    # clm-002: filed -> investigating (stays there)
    p.file_claim(
      "clm-002", policy_id: "pol-101", coverage_type: "epl",
      incident_date: CLM_D0, filed_at: CLM_T0,
      claimed_amount: 30_000, reserve_amount: 30_000, actor: "adj-2"
    )
    p.advance_status("clm-002", "investigating", at: CLM_T1, actor: "adj-2")

    # clm-003: filed -> investigating -> evaluation (stays there)
    p.file_claim(
      "clm-003", policy_id: "pol-102", coverage_type: "do",
      incident_date: CLM_D0, filed_at: CLM_T0,
      claimed_amount: 200_000, reserve_amount: 150_000, actor: "adj-1"
    )
    p.advance_status("clm-003", "investigating", at: CLM_T1, actor: "adj-1")
    p.advance_status("clm-003", "evaluation", at: CLM_T2, actor: "adj-1")

    # clm-004: filed -> investigating -> evaluation -> denied
    p.file_claim(
      "clm-004", policy_id: "pol-103", coverage_type: "epl",
      incident_date: CLM_D0, filed_at: CLM_T0,
      claimed_amount: 10_000, reserve_amount: 10_000, actor: "adj-3"
    )
    p.advance_status("clm-004", "investigating", at: CLM_T1, actor: "adj-3")
    p.advance_status("clm-004", "evaluation", at: CLM_T2, actor: "adj-3")
    p.deny_claim("clm-004", reason: "Coverage exclusion applies.", denied_at: CLM_T3, actor: "adj-3")

    p
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Claim filing, status transitions, reserve updates
  # ---------------------------------------------------------------------------
  describe "#file_claim" do
    it "returns a claim in the filed state" do
      c = pipeline.file_claim(
        "clm-new", policy_id: "pol-200", coverage_type: "do",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 50_000, reserve_amount: 40_000, actor: "adj-1"
      )
      expect(c[:claim_id]).to eq("clm-new")
      expect(c[:status]).to eq("filed")
      expect(c[:approved_amount]).to be_nil
      expect(c[:claimed_amount]).to eq(50_000)
      expect(c[:reserve_amount]).to eq(40_000)
    end

    it "records the initial event" do
      pipeline.file_claim(
        "clm-evt", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      c = pipeline.get_claim("clm-evt")
      expect(c[:events].size).to eq(1)
      expect(c[:events].first[:action]).to eq("filed")
      expect(c[:events].first[:actor]).to eq("adj-1")
    end

    it "raises ArgumentError on a duplicate" do
      pipeline.file_claim(
        "clm-dup", policy_id: "pol-200", coverage_type: "do",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 10_000, reserve_amount: 10_000, actor: "adj-1"
      )
      expect do
        pipeline.file_claim(
          "clm-dup", policy_id: "pol-201", coverage_type: "do",
          incident_date: CLM_D0, filed_at: CLM_T1,
          claimed_amount: 5_000, reserve_amount: 5_000, actor: "adj-1"
        )
      end.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a zero claimed_amount" do
      expect do
        pipeline.file_claim(
          "clm-zero", policy_id: "pol-200", coverage_type: "epl",
          incident_date: CLM_D0, filed_at: CLM_T0,
          claimed_amount: 0, reserve_amount: 10_000, actor: "adj-1"
        )
      end.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a zero reserve_amount" do
      expect do
        pipeline.file_claim(
          "clm-zero-r", policy_id: "pol-200", coverage_type: "epl",
          incident_date: CLM_D0, filed_at: CLM_T0,
          claimed_amount: 10_000, reserve_amount: 0, actor: "adj-1"
        )
      end.to raise_error(ArgumentError)
    end
  end

  describe "#get_claim" do
    it "returns the existing claim" do
      expect(seeded_pipeline.get_claim("clm-001")[:claim_id]).to eq("clm-001")
    end

    it "raises KeyError for an unknown claim" do
      expect { seeded_pipeline.get_claim("no-such") }.to raise_error(KeyError)
    end
  end

  describe "#advance_status" do
    it "updates the status on a valid transition" do
      pipeline.file_claim(
        "clm-adv", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      pipeline.advance_status("clm-adv", "investigating", at: CLM_T1, actor: "adj-1")
      expect(pipeline.get_claim("clm-adv")[:status]).to eq("investigating")
    end

    it "raises ArgumentError on an invalid transition" do
      pipeline.file_claim(
        "clm-inv", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      expect { pipeline.advance_status("clm-inv", "evaluation", at: CLM_T1, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError when settling via advance_status" do
      # "settled" must go through settle_claim, not advance_status
      expect { seeded_pipeline.advance_status("clm-003", "settled", at: CLM_T3, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError when denying via advance_status" do
      expect { seeded_pipeline.advance_status("clm-003", "denied", at: CLM_T3, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "appends a status_change event" do
      pipeline.file_claim(
        "clm-ev2", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      pipeline.advance_status("clm-ev2", "investigating", at: CLM_T1, actor: "adj-2")
      last = pipeline.get_claim("clm-ev2")[:events].last
      expect(last[:action]).to eq("status_change")
      expect(last[:payload][:from_status]).to eq("filed")
      expect(last[:payload][:to_status]).to eq("investigating")
    end

    it "raises KeyError for an unknown claim" do
      expect { seeded_pipeline.advance_status("no-such", "investigating", at: CLM_T1, actor: "adj-1") }.to raise_error(KeyError)
    end
  end

  describe "#update_reserve" do
    it "updates the reserve amount" do
      seeded_pipeline.update_reserve("clm-002", 35_000, at: CLM_T2, actor: "adj-2")
      expect(seeded_pipeline.get_claim("clm-002")[:reserve_amount]).to eq(35_000)
    end

    it "appends a reserve_update event" do
      seeded_pipeline.update_reserve("clm-002", 35_000, at: CLM_T2, actor: "adj-2")
      last = seeded_pipeline.get_claim("clm-002")[:events].last
      expect(last[:action]).to eq("reserve_update")
      expect(last[:payload][:old_reserve]).to eq(30_000)
      expect(last[:payload][:new_reserve]).to eq(35_000)
    end

    it "raises ArgumentError for a zero reserve" do
      expect { seeded_pipeline.update_reserve("clm-002", 0, at: CLM_T2, actor: "adj-2") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a closed claim" do
      pipeline.file_claim(
        "clm-cls", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      pipeline.advance_status("clm-cls", "investigating", at: CLM_T1, actor: "adj-1")
      pipeline.advance_status("clm-cls", "evaluation", at: CLM_T2, actor: "adj-1")
      pipeline.settle_claim("clm-cls", approved_amount: 10_000, settled_at: CLM_T3, actor: "adj-1")
      pipeline.advance_status("clm-cls", "closed", at: CLM_T4, actor: "adj-1")
      expect { pipeline.update_reserve("clm-cls", 5_000, at: CLM_T4, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown claim" do
      expect { seeded_pipeline.update_reserve("no-such", 10_000, at: CLM_T2, actor: "adj-1") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Settlement, denial, and query methods
  # ---------------------------------------------------------------------------
  describe "#settle_claim" do
    it "settles the claim and sets approved_amount" do
      # clm-003 is in evaluation
      seeded_pipeline.settle_claim("clm-003", approved_amount: 150_000, settled_at: CLM_T3, actor: "adj-1")
      c = seeded_pipeline.get_claim("clm-003")
      expect(c[:status]).to eq("settled")
      expect(c[:approved_amount]).to eq(150_000)
    end

    it "appends a settled event" do
      seeded_pipeline.settle_claim("clm-003", approved_amount: 150_000, settled_at: CLM_T3, actor: "adj-1")
      last = seeded_pipeline.get_claim("clm-003")[:events].last
      expect(last[:action]).to eq("settled")
      expect(last[:payload][:approved_amount]).to eq(150_000)
    end

    it "raises ArgumentError from the wrong state" do
      # clm-002 is in investigating, not evaluation
      expect { seeded_pipeline.settle_claim("clm-002", approved_amount: 20_000, settled_at: CLM_T3, actor: "adj-2") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError when approved exceeds claimed" do
      expect { seeded_pipeline.settle_claim("clm-003", approved_amount: 300_000, settled_at: CLM_T3, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for a zero approved amount" do
      expect { seeded_pipeline.settle_claim("clm-003", approved_amount: 0, settled_at: CLM_T3, actor: "adj-1") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown claim" do
      expect { seeded_pipeline.settle_claim("no-such", approved_amount: 10_000, settled_at: CLM_T3, actor: "adj-1") }.to raise_error(KeyError)
    end
  end

  describe "#deny_claim" do
    it "denies the claim" do
      # clm-003 is in evaluation
      seeded_pipeline.deny_claim("clm-003", reason: "Outside coverage period.", denied_at: CLM_T3, actor: "adj-1")
      expect(seeded_pipeline.get_claim("clm-003")[:status]).to eq("denied")
    end

    it "appends a denied event" do
      seeded_pipeline.deny_claim("clm-003", reason: "Exclusion.", denied_at: CLM_T3, actor: "adj-1")
      last = seeded_pipeline.get_claim("clm-003")[:events].last
      expect(last[:action]).to eq("denied")
      expect(last[:payload][:reason]).to eq("Exclusion.")
    end

    it "raises ArgumentError from the wrong state" do
      expect { seeded_pipeline.deny_claim("clm-002", reason: "Reason.", denied_at: CLM_T3, actor: "adj-2") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown claim" do
      expect { seeded_pipeline.deny_claim("no-such", reason: "Reason.", denied_at: CLM_T3, actor: "adj-1") }.to raise_error(KeyError)
    end
  end

  describe "#get_claims_by_policy" do
    it "returns claims for the policy" do
      ids = seeded_pipeline.get_claims_by_policy("pol-101").map { |c| c[:claim_id] }
      expect(ids).to include("clm-001", "clm-002")
      expect(ids).not_to include("clm-003")
    end

    it "is sorted by filed_at" do
      # Both filed at CLM_T0/CLM_T1; filing order should match
      pipeline.file_claim(
        "clm-b", policy_id: "pol-sort", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T1,
        claimed_amount: 10_000, reserve_amount: 8_000, actor: "adj-1"
      )
      pipeline.file_claim(
        "clm-a", policy_id: "pol-sort", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 10_000, reserve_amount: 8_000, actor: "adj-1"
      )
      filed_ats = pipeline.get_claims_by_policy("pol-sort").map { |c| c[:filed_at] }
      expect(filed_ats).to eq(filed_ats.sort)
    end

    it "returns empty for an unknown policy" do
      expect(seeded_pipeline.get_claims_by_policy("pol-999")).to eq([])
    end
  end

  describe "#get_open_claims" do
    it "excludes closed and denied claims" do
      ids = seeded_pipeline.get_open_claims.map { |c| c[:claim_id] }
      # clm-001 is settled (not yet closed, so it IS open); clm-004 is denied (excluded)
      expect(ids).to include("clm-001", "clm-002", "clm-003")
      expect(ids).not_to include("clm-004")
    end

    it "is sorted by filed_at" do
      filed_ats = seeded_pipeline.get_open_claims.map { |c| c[:filed_at] }
      expect(filed_ats).to eq(filed_ats.sort)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Reserve adequacy and metrics
  # ---------------------------------------------------------------------------
  describe "#get_reserve_adequacy" do
    it "sums reserves across all claims" do
      # clm-001: 50000  clm-002: 30000  clm-003: 150000  clm-004: 10000
      expect(seeded_pipeline.get_reserve_adequacy[:total_reserves]).to eq(240_000)
    end

    it "sums approved amounts for settled claims only" do
      # Only clm-001 is settled, approved=60000
      expect(seeded_pipeline.get_reserve_adequacy[:total_approved]).to eq(60_000)
    end

    it "counts and sums the under-reserved gap" do
      # clm-001: reserve=50000 vs approved=60000 -> under-reserved by 10000
      adequacy = seeded_pipeline.get_reserve_adequacy
      expect(adequacy[:under_reserved_count]).to eq(1)
      expect(adequacy[:under_reserved_gap]).to eq(10_000)
    end

    it "returns zeros for an empty pipeline" do
      adequacy = pipeline.get_reserve_adequacy
      expect(adequacy[:total_reserves]).to eq(0)
      expect(adequacy[:total_approved]).to eq(0)
      expect(adequacy[:under_reserved_count]).to eq(0)
      expect(adequacy[:under_reserved_gap]).to eq(0)
    end
  end

  describe "#get_claims_metrics" do
    it "reports the total count" do
      expect(seeded_pipeline.get_claims_metrics[:total]).to eq(4)
    end

    it "reports counts by status" do
      by_status = seeded_pipeline.get_claims_metrics[:by_status]
      expect(by_status["settled"]).to eq(1)
      expect(by_status["investigating"]).to eq(1)
      expect(by_status["evaluation"]).to eq(1)
      expect(by_status["denied"]).to eq(1)
    end

    it "only includes statuses with a nonzero count" do
      seeded_pipeline.get_claims_metrics[:by_status].each_value do |count|
        expect(count).to be > 0
      end
    end

    it "sums total_claimed" do
      expect(seeded_pipeline.get_claims_metrics[:total_claimed]).to eq(75_000 + 30_000 + 200_000 + 10_000)
    end

    it "sums total_paid from settled claims" do
      # only clm-001 settled, approved=60000
      expect(seeded_pipeline.get_claims_metrics[:total_paid]).to eq(60_000)
    end

    it "computes the average settlement ratio" do
      # 1 settled claim: approved=60000 / claimed=75000 = 0.8
      expect(seeded_pipeline.get_claims_metrics[:avg_settlement_ratio]).to eq((60_000 / 75_000.0).round(4))
    end

    it "returns zero average ratio with no settled claims" do
      pipeline.file_claim(
        "clm-ns", policy_id: "pol-200", coverage_type: "epl",
        incident_date: CLM_D0, filed_at: CLM_T0,
        claimed_amount: 20_000, reserve_amount: 15_000, actor: "adj-1"
      )
      expect(pipeline.get_claims_metrics[:avg_settlement_ratio]).to eq(0.0)
    end
  end

  describe "#get_policy_loss_history" do
    it "returns the correct stats" do
      history = seeded_pipeline.get_policy_loss_history("pol-101")
      expect(history[:policy_id]).to eq("pol-101")
      expect(history[:claim_count]).to eq(2)
      expect(history[:total_claimed]).to eq(105_000) # 75000 + 30000
      expect(history[:total_paid]).to eq(60_000) # only clm-001 settled
      expect(history[:loss_ratio]).to eq((60_000 / 105_000.0).round(4))
    end

    it "returns zeros for an unknown policy" do
      history = seeded_pipeline.get_policy_loss_history("pol-999")
      expect(history[:policy_id]).to eq("pol-999")
      expect(history[:claim_count]).to eq(0)
      expect(history[:total_claimed]).to eq(0)
      expect(history[:total_paid]).to eq(0)
      expect(history[:loss_ratio]).to eq(0.0)
    end
  end
end
