load_problem("05_medication_titration_tracker")

# ---------------------------------------------------------------------------
# Pre-existing patient event histories
# ---------------------------------------------------------------------------

# Maria: successfully de-escalated off metformin, still on low-dose glipizide
MARIA_EVENTS = [
  TitrationEvent.new("maria", "metformin", "start", 500.0, Date.new(2023, 6, 1)),
  TitrationEvent.new("maria", "metformin", "increase", 1000.0, Date.new(2023, 8, 1)),
  TitrationEvent.new("maria", "metformin", "decrease", 500.0, Date.new(2023, 11, 1)),
  TitrationEvent.new("maria", "metformin", "stop", 0.0, Date.new(2024, 2, 1)),
  TitrationEvent.new("maria", "glipizide", "start", 5.0, Date.new(2023, 6, 1)),
  TitrationEvent.new("maria", "glipizide", "decrease", 2.5, Date.new(2024, 1, 15)),
].freeze

# James: still on two active medications, one recently decreased
JAMES_EVENTS = [
  TitrationEvent.new("james", "metformin", "start", 500.0, Date.new(2023, 9, 1)),
  TitrationEvent.new("james", "metformin", "increase", 1000.0, Date.new(2023, 12, 1)),
  TitrationEvent.new("james", "insulin_glargine", "start", 10.0, Date.new(2023, 9, 1)),
  TitrationEvent.new("james", "insulin_glargine", "increase", 15.0, Date.new(2024, 1, 1)),
  TitrationEvent.new("james", "insulin_glargine", "decrease", 10.0, Date.new(2024, 3, 1)),
].freeze

# Susan: completely off all medications
SUSAN_EVENTS = [
  TitrationEvent.new("susan", "metformin", "start", 500.0, Date.new(2023, 3, 1)),
  TitrationEvent.new("susan", "metformin", "increase", 750.0, Date.new(2023, 5, 1)),
  TitrationEvent.new("susan", "metformin", "stop", 0.0, Date.new(2023, 10, 1)),
  TitrationEvent.new("susan", "glipizide", "start", 5.0, Date.new(2023, 3, 1)),
  TitrationEvent.new("susan", "glipizide", "stop", 0.0, Date.new(2023, 9, 1)),
].freeze

ALL_EVENTS = (MARIA_EVENTS + JAMES_EVENTS + SUSAN_EVENTS).freeze

RSpec.describe TitrationTracker do
  let(:tracker) { described_class.new(ALL_EVENTS.dup) }
  let(:fresh_tracker) { described_class.new([]) }

  # ---------------------------------------------------------------------------
  # PART 1 — Current medication snapshot
  # ---------------------------------------------------------------------------
  describe "#current_medications" do
    it "includes an active medication" do
      names = tracker.current_medications("maria").map(&:name)
      expect(names).to include("glipizide")
    end

    it "excludes a stopped medication" do
      names = tracker.current_medications("maria").map(&:name)
      expect(names).not_to include("metformin")
    end

    it "returns empty when everything has been stopped" do
      expect(tracker.current_medications("susan")).to eq([])
    end

    it "returns empty for an unknown patient" do
      expect(tracker.current_medications("nobody")).to eq([])
    end

    it "populates the Medication fields correctly" do
      glip = tracker.current_medications("maria").find { |m| m.name == "glipizide" }
      expect(glip.current_dose).to eq(2.5)
      expect(glip.last_changed).to eq(Date.new(2024, 1, 15))
      expect(glip.total_changes).to eq(2) # start + decrease
    end

    it "returns multiple active medications" do
      names = tracker.current_medications("james").map(&:name)
      expect(names).to include("metformin", "insulin_glargine")
    end

    it "sorts events internally regardless of insertion order" do
      events = [
        TitrationEvent.new("pt_x", "metformin", "increase", 1000.0, Date.new(2024, 3, 1)),
        TitrationEvent.new("pt_x", "metformin", "start", 500.0, Date.new(2024, 1, 1)),
        TitrationEvent.new("pt_x", "metformin", "decrease", 750.0, Date.new(2024, 2, 1)),
      ]
      t = described_class.new(events)
      meds = t.current_medications("pt_x")
      expect(meds.size).to eq(1)
      expect(meds.first.current_dose).to eq(1000.0)
    end
  end

  describe "#get_medication_history" do
    it "returns events sorted chronologically" do
      history = tracker.get_medication_history("maria", "metformin")
      dates = history.map(&:recorded_on)
      expect(dates).to eq(dates.sort)
    end

    it "returns the correct events in order" do
      history = tracker.get_medication_history("maria", "metformin")
      expect(history.map(&:direction)).to eq(%w[start increase decrease stop])
    end

    it "returns an empty array for an unknown combination" do
      expect(tracker.get_medication_history("nobody", "metformin")).to eq([])
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Titration counts
  # ---------------------------------------------------------------------------
  describe "#titration_count" do
    it "counts all events for a medication" do
      expect(tracker.titration_count("maria", "metformin")).to eq(4)
    end

    it "filters by direction: decrease" do
      expect(tracker.titration_count("maria", "metformin", direction: "decrease")).to eq(1)
    end

    it "filters by direction: stop" do
      expect(tracker.titration_count("susan", "metformin", direction: "stop")).to eq(1)
    end

    it "returns 0 when no events match the direction" do
      expect(tracker.titration_count("james", "metformin", direction: "stop")).to eq(0)
    end

    it "returns 0 for an unknown patient" do
      expect(tracker.titration_count("nobody", "metformin")).to eq(0)
    end

    it "returns 0 for an unknown medication" do
      expect(tracker.titration_count("maria", "insulin_glargine")).to eq(0)
    end
  end

  describe "#de_escalation_summary" do
    it "includes decrease and stop counts per medication" do
      summary = tracker.de_escalation_summary("maria")
      # metformin: 1 decrease + 1 stop = 2; glipizide: 1 decrease = 1
      expect(summary["metformin"]).to eq(2)
      expect(summary["glipizide"]).to eq(1)
    end

    it "returns empty when there are no de-escalations" do
      events = [
        TitrationEvent.new("pt_y", "metformin", "start", 500.0, Date.new(2024, 1, 1)),
        TitrationEvent.new("pt_y", "metformin", "increase", 1000.0, Date.new(2024, 2, 1)),
      ]
      t = described_class.new(events)
      expect(t.de_escalation_summary("pt_y")).to eq({})
    end

    it "excludes medications with no decrease or stop events" do
      summary = tracker.de_escalation_summary("james")
      expect(summary["insulin_glargine"]).to eq(1)
      expect(summary).not_to have_key("metformin") # only start+increase
    end

    it "returns empty for an unknown patient" do
      expect(tracker.de_escalation_summary("nobody")).to eq({})
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Population-level queries
  # ---------------------------------------------------------------------------
  describe "#patients_on_medication" do
    it "returns only patients with the medication currently active" do
      # maria stopped metformin, susan stopped metformin, james still active
      expect(tracker.patients_on_medication("metformin")).to eq(["james"])
    end

    it "returns patient_ids sorted alphabetically" do
      extra = [
        TitrationEvent.new("zoe", "glipizide", "start", 5.0, Date.new(2024, 1, 1)),
        TitrationEvent.new("anna", "glipizide", "start", 5.0, Date.new(2024, 1, 1)),
      ]
      t = described_class.new(ALL_EVENTS + extra)
      patients = t.patients_on_medication("glipizide")
      expect(patients).to eq(patients.sort)
      expect(patients).to include("maria", "anna", "zoe")
      expect(patients).not_to include("susan") # susan stopped glipizide
    end

    it "returns empty for a medication nobody is on" do
      expect(tracker.patients_on_medication("glipizide_xr")).to eq([])
    end
  end

  describe "#most_titrated_medications" do
    it "returns medications sorted descending by total event count" do
      top = tracker.most_titrated_medications(top_n: 3)
      names = top.map(&:first)
      counts = top.map(&:last)
      # metformin: maria(4) + james(2) + susan(3) = 9 events total
      expect(names).to include("metformin")
      expect(counts).to eq(counts.sort.reverse)
    end

    it "returns fewer than top_n when fewer medications exist" do
      events = [TitrationEvent.new("pt_z", "drug_a", "start", 10.0, Date.new(2024, 1, 1))]
      t = described_class.new(events)
      top = t.most_titrated_medications(top_n: 5)
      expect(top.size).to eq(1)
    end

    it "reports the correct total count" do
      top = tracker.most_titrated_medications(top_n: 10).to_h
      expect(top["metformin"]).to eq(9) # 4 + 2 + 3
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — Live ingestion
  # ---------------------------------------------------------------------------
  describe "#add_event" do
    it "updates current_medications with a new event" do
      event = TitrationEvent.new("james", "metformin", "stop", 0.0, Date.new(2024, 6, 1))
      tracker.add_event(event)
      names = tracker.current_medications("james").map(&:name)
      expect(names).not_to include("metformin")
    end

    it "updates medication history" do
      event = TitrationEvent.new("maria", "glipizide", "stop", 0.0, Date.new(2024, 6, 1))
      tracker.add_event(event)
      directions = tracker.get_medication_history("maria", "glipizide").map(&:direction)
      expect(directions).to include("stop")
    end

    it "overwrites an existing event on the same date" do
      event = TitrationEvent.new("james", "metformin", "stop", 0.0, Date.new(2023, 12, 1))
      tracker.add_event(event)
      history = tracker.get_medication_history("james", "metformin")
      dec_event = history.find { |e| e.recorded_on == Date.new(2023, 12, 1) }
      expect(dec_event.direction).to eq("stop")
    end

    it "creates a new patient from an event" do
      event = TitrationEvent.new("new_pt", "metformin", "start", 500.0, Date.new(2024, 5, 1))
      tracker.add_event(event)
      meds = tracker.current_medications("new_pt")
      expect(meds.size).to eq(1)
      expect(meds.first.name).to eq("metformin")
    end

    it "affects population-level queries" do
      event = TitrationEvent.new("new_pt2", "metformin", "start", 500.0, Date.new(2024, 5, 1))
      tracker.add_event(event)
      expect(tracker.patients_on_medication("metformin")).to include("new_pt2")
    end
  end
end
