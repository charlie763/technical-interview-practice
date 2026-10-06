load_problem("04_biomarker_alert_monitor")

# ---------------------------------------------------------------------------
# Pre-existing data — represents a snapshot of patient readings already in
# the system when the monitor is initialized. Examples below build on this.
# ---------------------------------------------------------------------------

ALICE_READINGS = [
  # All in range — no outreach needed
  BiomarkerReading.new("alice", "glucose", 105.0, Date.new(2024, 1, 1)),
  BiomarkerReading.new("alice", "glucose", 98.0, Date.new(2024, 1, 2)),
  BiomarkerReading.new("alice", "glucose", 112.0, Date.new(2024, 1, 3)),
  BiomarkerReading.new("alice", "glucose", 91.0, Date.new(2024, 1, 4)),
].freeze

BOB_READINGS = [
  # 4 consecutive high-glucose days -> needs outreach
  BiomarkerReading.new("bob", "glucose", 195.0, Date.new(2024, 1, 1)),
  BiomarkerReading.new("bob", "glucose", 210.0, Date.new(2024, 1, 2)),
  BiomarkerReading.new("bob", "glucose", 188.0, Date.new(2024, 1, 3)),
  BiomarkerReading.new("bob", "glucose", 202.0, Date.new(2024, 1, 4)),
].freeze

CAROL_READINGS = [
  # Streak of 1 (Jan 1), in-range on Jan 2, then streak of 2 (Jan 3-4) -> max=2
  BiomarkerReading.new("carol", "glucose", 190.0, Date.new(2024, 1, 1)),
  BiomarkerReading.new("carol", "glucose", 150.0, Date.new(2024, 1, 2)), # in range
  BiomarkerReading.new("carol", "glucose", 185.0, Date.new(2024, 1, 3)),
  BiomarkerReading.new("carol", "glucose", 191.0, Date.new(2024, 1, 4)),
].freeze

DAVE_READINGS = [
  # 2 consecutive high-glucose days — below default threshold of 3
  BiomarkerReading.new("dave", "glucose", 199.0, Date.new(2024, 1, 3)),
  BiomarkerReading.new("dave", "glucose", 205.0, Date.new(2024, 1, 4)),
].freeze

EVE_READINGS = [
  # One dangerous low glucose (below 70) on Jan 1 — streak of 1 for glucose
  BiomarkerReading.new("eve", "glucose", 62.0, Date.new(2024, 1, 1)),
  # 3 consecutive days with ketones below target (< 0.5) -> ketone outreach
  BiomarkerReading.new("eve", "ketone", 0.3, Date.new(2024, 1, 2)),
  BiomarkerReading.new("eve", "ketone", 0.2, Date.new(2024, 1, 3)),
  BiomarkerReading.new("eve", "ketone", 0.4, Date.new(2024, 1, 4)),
  # Weight readings — never out of range
  BiomarkerReading.new("eve", "weight", 165.0, Date.new(2024, 1, 1)),
  BiomarkerReading.new("eve", "weight", 164.5, Date.new(2024, 1, 2)),
].freeze

ALL_READINGS = (ALICE_READINGS + BOB_READINGS + CAROL_READINGS + DAVE_READINGS + EVE_READINGS).freeze

RSpec.describe BiomarkerMonitor do
  let(:monitor) { described_class.new(ALL_READINGS.dup) }
  let(:fresh_monitor) { described_class.new([]) }

  # ---------------------------------------------------------------------------
  # PART 1 — Single-reading classification
  # ---------------------------------------------------------------------------
  describe "#is_out_of_range" do
    it "flags glucose above range" do
      r = BiomarkerReading.new("p1", "glucose", 181.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be true
    end

    it "flags glucose below range" do
      r = BiomarkerReading.new("p1", "glucose", 69.9, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be true
    end

    it "does not flag glucose at the upper boundary" do
      r = BiomarkerReading.new("p1", "glucose", 180.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be false
    end

    it "does not flag glucose at the lower boundary" do
      r = BiomarkerReading.new("p1", "glucose", 70.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be false
    end

    it "does not flag glucose in range" do
      r = BiomarkerReading.new("p1", "glucose", 120.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be false
    end

    it "flags ketone above range" do
      r = BiomarkerReading.new("p1", "ketone", 3.1, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be true
    end

    it "flags ketone below range" do
      r = BiomarkerReading.new("p1", "ketone", 0.4, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be true
    end

    it "does not flag ketone in range" do
      r = BiomarkerReading.new("p1", "ketone", 1.5, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be false
    end

    it "does not flag ketone at either boundary" do
      lo = BiomarkerReading.new("p1", "ketone", 0.5, Date.new(2024, 1, 1))
      hi = BiomarkerReading.new("p1", "ketone", 3.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(lo)).to be false
      expect(fresh_monitor.is_out_of_range(hi)).to be false
    end

    it "never flags weight" do
      r = BiomarkerReading.new("p1", "weight", 9999.0, Date.new(2024, 1, 1))
      expect(fresh_monitor.is_out_of_range(r)).to be false
    end

    it "classifies any reading regardless of preloaded monitor state" do
      in_range = BiomarkerReading.new("alice", "glucose", 100.0, Date.new(2024, 1, 5))
      out_of_range = BiomarkerReading.new("bob", "glucose", 250.0, Date.new(2024, 1, 5))
      expect(monitor.is_out_of_range(in_range)).to be false
      expect(monitor.is_out_of_range(out_of_range)).to be true
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Streak detection
  # ---------------------------------------------------------------------------
  describe "#max_consecutive_out_of_range_days" do
    it "returns 0 for an unknown patient" do
      expect(monitor.max_consecutive_out_of_range_days("unknown_patient", "glucose")).to eq(0)
    end

    it "returns 0 when everything is in range" do
      expect(monitor.max_consecutive_out_of_range_days("alice", "glucose")).to eq(0)
    end

    it "returns the full length of an unbroken streak" do
      expect(monitor.max_consecutive_out_of_range_days("bob", "glucose")).to eq(4)
    end

    it "returns the longest of multiple streaks broken by an in-range day" do
      expect(monitor.max_consecutive_out_of_range_days("carol", "glucose")).to eq(2)
    end

    it "handles a short streak below the outreach threshold" do
      expect(monitor.max_consecutive_out_of_range_days("dave", "glucose")).to eq(2)
    end

    it "counts a single out-of-range day as a streak of one" do
      expect(monitor.max_consecutive_out_of_range_days("eve", "glucose")).to eq(1)
    end

    it "tracks a streak independently per reading type" do
      expect(monitor.max_consecutive_out_of_range_days("eve", "ketone")).to eq(3)
    end

    it "never streaks for weight" do
      expect(monitor.max_consecutive_out_of_range_days("eve", "weight")).to eq(0)
    end

    it "collapses multiple same-day readings into one day" do
      readings = [
        BiomarkerReading.new("frank", "glucose", 200.0, Date.new(2024, 2, 1)),
        BiomarkerReading.new("frank", "glucose", 210.0, Date.new(2024, 2, 1)), # same day
        BiomarkerReading.new("frank", "glucose", 195.0, Date.new(2024, 2, 2)),
      ]
      m = described_class.new(readings)
      expect(m.max_consecutive_out_of_range_days("frank", "glucose")).to eq(2)
    end

    it "treats a skipped calendar day as two separate streaks" do
      readings = [
        BiomarkerReading.new("grace", "glucose", 200.0, Date.new(2024, 3, 1)),
        BiomarkerReading.new("grace", "glucose", 200.0, Date.new(2024, 3, 3)), # skip Mar 2
      ]
      m = described_class.new(readings)
      expect(m.max_consecutive_out_of_range_days("grace", "glucose")).to eq(1)
    end

    it "ignores readings of a different type" do
      expect(monitor.max_consecutive_out_of_range_days("eve", "glucose")).to eq(1)
      expect(monitor.max_consecutive_out_of_range_days("eve", "ketone")).to eq(3)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Outreach list
  # ---------------------------------------------------------------------------
  describe "#get_outreach_list" do
    it "filters by the default threshold of 3" do
      result = monitor.get_outreach_list
      patient_types = result.map { |e| [e[:patient_id], e[:reading_type]] }
      expect(patient_types).to include(%w[bob glucose], %w[eve ketone])
      expect(patient_types).not_to include(%w[alice glucose], %w[carol glucose], %w[dave glucose])
    end

    it "never includes weight" do
      result = monitor.get_outreach_list(min_consecutive_days: 1)
      expect(result.map { |e| e[:reading_type] }).not_to include("weight")
    end

    it "is sorted by consecutive_days descending" do
      result = monitor.get_outreach_list(min_consecutive_days: 1)
      days = result.map { |e| e[:consecutive_days] }
      expect(days).to eq(days.sort.reverse)
    end

    it "returns entries with the expected fields" do
      result = monitor.get_outreach_list
      bob_entry = result.find { |e| e[:patient_id] == "bob" }
      expect(bob_entry.keys).to contain_exactly(:patient_id, :reading_type, :consecutive_days, :latest_value)
      expect(bob_entry[:consecutive_days]).to eq(4)
      expect(bob_entry[:reading_type]).to eq("glucose")
      expect(bob_entry[:latest_value]).to be_a(Float)
    end

    it "reports the most recent out-of-range value" do
      result = monitor.get_outreach_list
      bob_entry = result.find { |e| e[:patient_id] == "bob" }
      expect(bob_entry[:latest_value]).to eq(202.0)
    end

    it "lets a patient appear twice for different reading types" do
      readings = [
        BiomarkerReading.new("hank", "glucose", 200.0, Date.new(2024, 1, 1)),
        BiomarkerReading.new("hank", "glucose", 210.0, Date.new(2024, 1, 2)),
        BiomarkerReading.new("hank", "glucose", 195.0, Date.new(2024, 1, 3)),
        BiomarkerReading.new("hank", "ketone", 0.2, Date.new(2024, 1, 1)),
        BiomarkerReading.new("hank", "ketone", 0.3, Date.new(2024, 1, 2)),
        BiomarkerReading.new("hank", "ketone", 0.1, Date.new(2024, 1, 3)),
      ]
      m = described_class.new(readings)
      result = m.get_outreach_list(min_consecutive_days: 3)
      patient_types = result.map { |e| [e[:patient_id], e[:reading_type]] }
      expect(patient_types).to include(%w[hank glucose], %w[hank ketone])
    end

    it "honors a custom threshold" do
      result = monitor.get_outreach_list(min_consecutive_days: 2)
      patient_types = result.map { |e| [e[:patient_id], e[:reading_type]] }
      expect(patient_types).to include(%w[carol glucose], %w[dave glucose])
    end

    it "returns an empty list for an empty monitor" do
      expect(fresh_monitor.get_outreach_list).to eq([])
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — Deduplication on ingestion
  # ---------------------------------------------------------------------------
  describe "#add_reading" do
    it "returns true for a new reading" do
      r = BiomarkerReading.new("alice", "glucose", 100.0, Date.new(2024, 1, 10))
      expect(monitor.add_reading(r)).to be true
    end

    it "returns false for an exact duplicate" do
      r = BiomarkerReading.new("alice", "glucose", 105.0, Date.new(2024, 1, 1))
      expect(monitor.add_reading(r)).to be false
    end

    it "returns false for a duplicate within tolerance" do
      # alice Jan 1 = 105.0; 105.4 is within +/-0.5
      r = BiomarkerReading.new("alice", "glucose", 105.4, Date.new(2024, 1, 1))
      expect(monitor.add_reading(r)).to be false
    end

    it "returns true just outside the tolerance window" do
      # alice Jan 1 = 105.0; 105.6 is outside +/-0.5
      r = BiomarkerReading.new("alice", "glucose", 105.6, Date.new(2024, 1, 1))
      expect(monitor.add_reading(r)).to be true
    end

    it "is not a duplicate on a different date" do
      r = BiomarkerReading.new("alice", "glucose", 105.0, Date.new(2024, 1, 5))
      expect(monitor.add_reading(r)).to be true
    end

    it "is not a duplicate for a different reading type" do
      r = BiomarkerReading.new("alice", "ketone", 1.0, Date.new(2024, 1, 1))
      expect(monitor.add_reading(r)).to be true
    end

    it "is not a duplicate for a different patient" do
      r = BiomarkerReading.new("frank", "glucose", 105.0, Date.new(2024, 1, 1))
      expect(monitor.add_reading(r)).to be true
    end

    it "extends the streak once the new day is added" do
      # bob's current streak is Jan 1-4 (4 days). Add Jan 5.
      r = BiomarkerReading.new("bob", "glucose", 195.0, Date.new(2024, 1, 5))
      monitor.add_reading(r)
      expect(monitor.max_consecutive_out_of_range_days("bob", "glucose")).to eq(5)
    end

    it "does not let a rejected duplicate inflate the streak" do
      r = BiomarkerReading.new("bob", "glucose", 195.0, Date.new(2024, 1, 1))
      monitor.add_reading(r) # duplicate, rejected
      expect(monitor.max_consecutive_out_of_range_days("bob", "glucose")).to eq(4)
    end
  end
end
