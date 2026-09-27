load_problem("08_enrollment_pipeline")

RSpec.describe EnrollmentPipeline do
  let(:pipeline) { described_class.new }

  # Pre-seeded pipeline with patients in various states (all timestamps in seconds):
  #   seed_active:      referred(0) -> screened(100) -> enrolled(200) -> active(300)
  #   seed_graduated:   referred(0) -> screened(50)  -> enrolled(150) -> active(250) -> graduated(1000)
  #   seed_ineligible:  referred(0) -> screened(10)  -> ineligible(20)
  #   seed_referred:    referred(0)  [no further transitions]
  let(:seeded_pipeline) do
    p = described_class.new

    p.add_patient("seed_active", 0.0)
    p.transition("seed_active", "screened", 100.0)
    p.transition("seed_active", "enrolled", 200.0)
    p.transition("seed_active", "active", 300.0)

    p.add_patient("seed_graduated", 0.0)
    p.transition("seed_graduated", "screened", 50.0)
    p.transition("seed_graduated", "enrolled", 150.0)
    p.transition("seed_graduated", "active", 250.0)
    p.transition("seed_graduated", "graduated", 1000.0)

    p.add_patient("seed_ineligible", 0.0)
    p.transition("seed_ineligible", "screened", 10.0)
    p.transition("seed_ineligible", "ineligible", 20.0)

    p.add_patient("seed_referred", 0.0)
    p
  end

  # ── Part 1: State tracking ────────────────────────────────────────────────
  describe "#add_patient" do
    it "starts a new patient in referred" do
      pipeline.add_patient("add_p1", 0.0)
      expect(pipeline.get_state("add_p1")).to eq("referred")
    end

    it "raises ArgumentError on a duplicate patient_id" do
      pipeline.add_patient("dup_p", 0.0)
      expect { pipeline.add_patient("dup_p", 10.0) }.to raise_error(ArgumentError)
    end
  end

  describe "#transition" do
    it "changes state on a valid transition" do
      pipeline.add_patient("trans_p1", 0.0)
      pipeline.transition("trans_p1", "screened", 100.0)
      expect(pipeline.get_state("trans_p1")).to eq("screened")
    end

    it "raises ArgumentError on an invalid transition" do
      pipeline.add_patient("trans_p2", 0.0)
      # Cannot jump from referred straight to graduated
      expect { pipeline.transition("trans_p2", "graduated", 100.0) }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError when skipping a required state" do
      pipeline.add_patient("trans_p3", 0.0)
      pipeline.transition("trans_p3", "screened", 10.0)
      # Cannot skip enrolled -> jump from screened to active
      expect { pipeline.transition("trans_p3", "active", 20.0) }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown patient" do
      expect { pipeline.transition("ghost", "screened", 100.0) }.to raise_error(KeyError)
    end

    it "raises ArgumentError when transitioning from a terminal state" do
      pipeline.add_patient("trans_p4", 0.0)
      pipeline.transition("trans_p4", "screened", 10.0)
      pipeline.transition("trans_p4", "ineligible", 20.0)
      expect { pipeline.transition("trans_p4", "enrolled", 30.0) }.to raise_error(ArgumentError)
    end

    it "allows both branches out of screened" do
      pipeline.add_patient("branch_p1", 0.0)
      pipeline.transition("branch_p1", "screened", 10.0)
      pipeline.transition("branch_p1", "enrolled", 20.0)
      expect(pipeline.get_state("branch_p1")).to eq("enrolled")

      pipeline.add_patient("branch_p2", 0.0)
      pipeline.transition("branch_p2", "screened", 10.0)
      pipeline.transition("branch_p2", "ineligible", 20.0)
      expect(pipeline.get_state("branch_p2")).to eq("ineligible")
    end
  end

  describe "#get_state" do
    it "returns the current state for each patient" do
      expect(seeded_pipeline.get_state("seed_active")).to eq("active")
      expect(seeded_pipeline.get_state("seed_graduated")).to eq("graduated")
      expect(seeded_pipeline.get_state("seed_ineligible")).to eq("ineligible")
      expect(seeded_pipeline.get_state("seed_referred")).to eq("referred")
    end

    it "raises KeyError for an unknown patient" do
      expect { pipeline.get_state("nobody") }.to raise_error(KeyError)
    end
  end

  describe "#get_patients_in_state" do
    it "returns patients currently in the state" do
      active_patients = seeded_pipeline.get_patients_in_state("active")
      expect(active_patients).to include("seed_active")
      expect(active_patients).not_to include("seed_graduated")
    end

    it "returns a sorted result" do
      result = seeded_pipeline.get_patients_in_state("referred")
      expect(result).to eq(result.sort)
    end

    it "returns empty for an unpopulated state" do
      expect(seeded_pipeline.get_patients_in_state("withdrawn")).to eq([])
    end

    it "excludes a patient who has left the state" do
      # seed_ineligible passed through screened but is no longer there
      expect(seeded_pipeline.get_patients_in_state("screened")).not_to include("seed_ineligible")
    end
  end

  # ── Part 2: Duration and conversion metrics ───────────────────────────────
  describe "#time_in_state" do
    it "returns the exact duration for a completed state" do
      pipeline.add_patient("dur_p1", 0.0)
      pipeline.transition("dur_p1", "screened", 1000.0)
      pipeline.transition("dur_p1", "enrolled", 4000.0)
      # Spent exactly 3000 s in screened; as_of is ignored for completed states
      expect(pipeline.time_in_state("dur_p1", "screened", 99999.0)).to eq(3000.0)
    end

    it "counts up to as_of for the current state" do
      pipeline.add_patient("dur_p2", 0.0)
      pipeline.transition("dur_p2", "screened", 1000.0)
      expect(pipeline.time_in_state("dur_p2", "screened", 4000.0)).to eq(3000.0)
    end

    it "returns zero for a state never visited" do
      pipeline.add_patient("dur_p3", 0.0)
      expect(pipeline.time_in_state("dur_p3", "enrolled", 99999.0)).to eq(0.0)
    end

    it "times the initial referred state from the add_patient timestamp" do
      pipeline.add_patient("dur_p4", 500.0)
      pipeline.transition("dur_p4", "screened", 1500.0)
      # Spent 1000 s in referred (1500 - 500)
      expect(pipeline.time_in_state("dur_p4", "referred", 99999.0)).to eq(1000.0)
    end
  end

  describe "#conversion_rate" do
    it "computes a fifty percent conversion" do
      pipeline.add_patient("conv_p1", 0.0)
      pipeline.transition("conv_p1", "screened", 10.0)
      pipeline.transition("conv_p1", "enrolled", 20.0)

      pipeline.add_patient("conv_p2", 0.0)
      pipeline.transition("conv_p2", "screened", 10.0)
      pipeline.transition("conv_p2", "ineligible", 20.0)

      expect(pipeline.conversion_rate("screened", "enrolled")).to be_within(0.0001).of(0.5)
    end

    it "excludes patients still sitting in from_state" do
      pipeline.add_patient("conv_p3", 0.0)
      pipeline.transition("conv_p3", "screened", 10.0)
      # conv_p3 is still in screened — must not be counted

      pipeline.add_patient("conv_p4", 0.0)
      pipeline.transition("conv_p4", "screened", 10.0)
      pipeline.transition("conv_p4", "enrolled", 20.0)

      # Only conv_p4 has exited; they enrolled -> rate is 1.0
      expect(pipeline.conversion_rate("screened", "enrolled")).to be_within(0.0001).of(1.0)
    end

    it "returns zero when nobody has exited from_state" do
      pipeline.add_patient("conv_p5", 0.0)
      # conv_p5 is still in referred
      expect(pipeline.conversion_rate("referred", "screened")).to be_within(0.0001).of(0.0)
    end

    it "returns one hundred percent when everyone converted" do
      %w[conv_all_1 conv_all_2].each do |pid|
        pipeline.add_patient(pid, 0.0)
        pipeline.transition(pid, "screened", 10.0)
        pipeline.transition(pid, "enrolled", 20.0)
      end
      expect(pipeline.conversion_rate("screened", "enrolled")).to be_within(0.0001).of(1.0)
    end
  end

  # ── Part 3: SLA monitoring ────────────────────────────────────────────────
  describe "#patients_overdue" do
    it "returns patients exceeding the threshold" do
      pipeline.add_patient("over_p1", 0.0)
      pipeline.transition("over_p1", "screened", 0.0)
      pipeline.transition("over_p1", "enrolled", 0.0)
      pipeline.transition("over_p1", "active", 0.0) # 10000 s in active

      pipeline.add_patient("over_p2", 0.0)
      pipeline.transition("over_p2", "screened", 0.0)
      pipeline.transition("over_p2", "enrolled", 0.0)
      pipeline.transition("over_p2", "active", 5000.0) # 5000 s in active

      overdue = pipeline.patients_overdue("active", 6000.0, 10000.0)
      expect(overdue).to include("over_p1")
      expect(overdue).not_to include("over_p2")
    end

    it "sorts by duration descending" do
      pipeline.add_patient("sort_p1", 0.0)
      pipeline.transition("sort_p1", "screened", 0.0)
      pipeline.transition("sort_p1", "enrolled", 0.0)
      pipeline.transition("sort_p1", "active", 0.0) # 10000 s

      pipeline.add_patient("sort_p2", 0.0)
      pipeline.transition("sort_p2", "screened", 0.0)
      pipeline.transition("sort_p2", "enrolled", 0.0)
      pipeline.transition("sort_p2", "active", 3000.0) # 7000 s

      overdue = pipeline.patients_overdue("active", 5000.0, 10000.0)
      expect(overdue).to eq(%w[sort_p1 sort_p2])
    end

    it "excludes patients not currently in the state" do
      # seed_graduated has left active — must not appear in active overdue list
      overdue = seeded_pipeline.patients_overdue("active", 0.0, 2000.0)
      expect(overdue).not_to include("seed_graduated")
    end

    it "returns empty when nobody is overdue" do
      pipeline.add_patient("noover_p", 0.0)
      pipeline.transition("noover_p", "screened", 0.0)
      pipeline.transition("noover_p", "enrolled", 0.0)
      pipeline.transition("noover_p", "active", 9900.0) # only 100 s in active
      expect(pipeline.patients_overdue("active", 500.0, 10000.0)).to eq([])
    end
  end

  describe "#average_time_in_state" do
    it "averages duration across all exited patients" do
      # avg_pa: 1000 s in screened; avg_pb: 3000 s in screened -> average = 2000
      pipeline.add_patient("avg_pa", 0.0)
      pipeline.transition("avg_pa", "screened", 0.0)
      pipeline.transition("avg_pa", "enrolled", 1000.0)

      pipeline.add_patient("avg_pb", 0.0)
      pipeline.transition("avg_pb", "screened", 0.0)
      pipeline.transition("avg_pb", "enrolled", 3000.0)

      expect(pipeline.average_time_in_state("screened", 99999.0)).to be_within(0.0001).of(2000.0)
    end

    it "excludes patients still in the state" do
      pipeline.add_patient("avg_pc", 0.0)
      pipeline.transition("avg_pc", "screened", 0.0)
      pipeline.transition("avg_pc", "enrolled", 1000.0) # exited: 1000 s

      pipeline.add_patient("avg_pd", 0.0)
      pipeline.transition("avg_pd", "screened", 0.0)
      # avg_pd is still in screened at as_of=5000 — must be excluded

      expect(pipeline.average_time_in_state("screened", 5000.0)).to be_within(0.0001).of(1000.0)
    end

    it "returns zero when nobody has exited" do
      pipeline.add_patient("avg_pe", 0.0)
      pipeline.transition("avg_pe", "screened", 0.0)
      # avg_pe is still in screened
      expect(pipeline.average_time_in_state("screened", 5000.0)).to be_within(0.0001).of(0.0)
    end
  end
end
