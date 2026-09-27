load_problem("07_care_team_assignments")

RSpec.describe CareTeamManager do
  let(:mgr) { described_class.new }

  # Pre-seeded manager:
  #   - coach_alpha  (coach,     max=3): assigned patient_seed_1, patient_seed_2
  #   - coach_beta   (coach,     max=1): no patients
  #   - dr_omega     (physician, max=100): assigned patient_seed_1
  let(:seeded_mgr) do
    m = described_class.new
    m.add_member("coach_alpha", "coach", 3)
    m.add_member("coach_beta", "coach", 1)
    m.add_member("dr_omega", "physician", 100)
    m.assign("patient_seed_1", "coach_alpha", 1000.0)
    m.assign("patient_seed_1", "dr_omega", 1000.0)
    m.assign("patient_seed_2", "coach_alpha", 2000.0)
    m
  end

  # ── Part 1: Basic assignment and lookup ──────────────────────────────────
  describe "#add_member" do
    it "lets a registered member receive an assignment" do
      mgr.add_member("member_reg_test", "coach", 5)
      expect { mgr.assign("patient_reg_test", "member_reg_test", 0.0) }.not_to raise_error
    end

    it "raises KeyError when assigning to an unregistered member" do
      expect { mgr.assign("patient_unreg_test", "ghost_member", 0.0) }.to raise_error(KeyError)
    end
  end

  describe "#assign" do
    it "records a basic assignment" do
      mgr.add_member("coach_basic", "coach", 5)
      mgr.assign("patient_basic", "coach_basic", 100.0)
      expect(mgr.get_assignment("patient_basic", "coach")).to eq("coach_basic")
    end

    it "replaces the current member on reassignment" do
      mgr.add_member("coach_orig", "coach", 5)
      mgr.add_member("coach_new", "coach", 5)
      mgr.assign("patient_reassign", "coach_orig", 100.0)
      mgr.assign("patient_reassign", "coach_new", 200.0)
      expect(mgr.get_assignment("patient_reassign", "coach")).to eq("coach_new")
    end

    it "removes the patient from the old member on reassignment" do
      mgr.add_member("coach_from", "coach", 5)
      mgr.add_member("coach_to", "coach", 5)
      mgr.assign("patient_move", "coach_from", 100.0)
      mgr.assign("patient_move", "coach_to", 200.0)
      expect(mgr.get_patients("coach_from")).not_to include("patient_move")
    end

    it "keeps assignments across roles independent" do
      mgr.add_member("coach_ind", "coach", 5)
      mgr.add_member("dr_ind", "physician", 5)
      mgr.assign("patient_ind", "coach_ind", 100.0)
      mgr.assign("patient_ind", "dr_ind", 100.0)
      expect(mgr.get_assignment("patient_ind", "coach")).to eq("coach_ind")
      expect(mgr.get_assignment("patient_ind", "physician")).to eq("dr_ind")
    end
  end

  describe "#get_assignment" do
    it "returns the current member" do
      expect(seeded_mgr.get_assignment("patient_seed_1", "coach")).to eq("coach_alpha")
    end

    it "returns nil for a role the patient has no member for" do
      expect(seeded_mgr.get_assignment("patient_seed_1", "dietitian")).to be_nil
    end

    it "returns nil for an unknown patient" do
      expect(mgr.get_assignment("no_such_patient", "coach")).to be_nil
    end
  end

  describe "#get_patients" do
    it "returns all patients assigned to the member" do
      patients = seeded_mgr.get_patients("coach_alpha")
      expect(patients).to include("patient_seed_1", "patient_seed_2")
    end

    it "returns patients sorted" do
      patients = seeded_mgr.get_patients("coach_alpha")
      expect(patients).to eq(patients.sort)
    end

    it "returns an empty array for a member with no patients" do
      expect(seeded_mgr.get_patients("coach_beta")).to eq([])
    end

    it "raises KeyError for an unregistered member" do
      expect { mgr.get_patients("nobody") }.to raise_error(KeyError)
    end
  end

  # ── Part 2: Capacity enforcement ─────────────────────────────────────────
  describe "capacity enforcement" do
    it "raises CapacityError when the member is full" do
      mgr.add_member("coach_full", "coach", 1)
      mgr.assign("patient_cap_1", "coach_full", 100.0)
      expect { mgr.assign("patient_cap_2", "coach_full", 200.0) }.to raise_error(CapacityError)
    end

    it "does not raise when reassigning an existing patient to the same member" do
      mgr.add_member("coach_same", "coach", 1)
      mgr.assign("patient_same", "coach_same", 100.0)
      # Member is "full" but the patient is already theirs — must not raise
      expect { mgr.assign("patient_same", "coach_same", 200.0) }.not_to raise_error
    end

    it "frees capacity once a patient is reassigned away" do
      mgr.add_member("coach_donor", "coach", 1)
      mgr.add_member("coach_recv", "coach", 5)
      mgr.assign("patient_freed", "coach_donor", 100.0)
      # Move the patient away — coach_donor now has a free slot
      mgr.assign("patient_freed", "coach_recv", 200.0)
      # A new patient should now fit on coach_donor
      mgr.assign("patient_new", "coach_donor", 300.0)
      expect(mgr.get_assignment("patient_new", "coach")).to eq("coach_donor")
    end
  end

  describe "#available_members" do
    it "returns members with open capacity" do
      mgr.add_member("avail_coach_open", "coach", 2)
      mgr.add_member("avail_coach_full", "coach", 1)
      mgr.assign("avail_patient_1", "avail_coach_full", 100.0)
      available = mgr.available_members("coach")
      expect(available).to include("avail_coach_open")
      expect(available).not_to include("avail_coach_full")
    end

    it "excludes a full member" do
      mgr.add_member("only_coach", "coach", 1)
      mgr.assign("only_patient", "only_coach", 100.0)
      expect(mgr.available_members("coach")).not_to include("only_coach")
    end

    it "returns a sorted result" do
      mgr.add_member("sort_coach_z", "coach", 5)
      mgr.add_member("sort_coach_a", "coach", 5)
      result = mgr.available_members("coach")
      expect(result).to eq(result.sort)
    end

    it "returns empty when no members have that role" do
      mgr.add_member("solo_physician", "physician", 100)
      expect(mgr.available_members("coach")).to eq([])
    end
  end

  # ── Part 3: Assignment history ───────────────────────────────────────────
  describe "#get_history" do
    it "has an open end for a single assignment" do
      mgr.add_member("hist_coach_1", "coach", 5)
      mgr.assign("hist_patient_1", "hist_coach_1", 1000.0)
      expect(mgr.get_history("hist_patient_1", "coach")).to eq([["hist_coach_1", 1000.0, nil]])
    end

    it "closes the previous entry on reassignment" do
      mgr.add_member("hist_coach_a", "coach", 5)
      mgr.add_member("hist_coach_b", "coach", 5)
      mgr.assign("hist_patient_2", "hist_coach_a", 1000.0)
      mgr.assign("hist_patient_2", "hist_coach_b", 3000.0)
      expect(mgr.get_history("hist_patient_2", "coach")).to eq([
        ["hist_coach_a", 1000.0, 3000.0],
        ["hist_coach_b", 3000.0, nil],
      ])
    end

    it "sorts multiple reassignments chronologically" do
      %w[hist_cx hist_cy hist_cz].each { |name| mgr.add_member(name, "coach", 5) }
      mgr.assign("hist_patient_3", "hist_cx", 100.0)
      mgr.assign("hist_patient_3", "hist_cy", 200.0)
      mgr.assign("hist_patient_3", "hist_cz", 300.0)
      history = mgr.get_history("hist_patient_3", "coach")
      expect(history.map { |entry| entry[1] }).to eq([100.0, 200.0, 300.0])
      expect(history.last[2]).to be_nil
    end

    it "returns empty for a role the patient was never assigned" do
      mgr.add_member("hist_coach_only", "coach", 5)
      mgr.assign("hist_patient_4", "hist_coach_only", 100.0)
      expect(mgr.get_history("hist_patient_4", "physician")).to eq([])
    end

    it "returns empty for an unknown patient" do
      expect(mgr.get_history("hist_nobody", "coach")).to eq([])
    end
  end

  describe "#get_assignment_at" do
    it "returns the member during the active window" do
      mgr.add_member("at_coach_1", "coach", 5)
      mgr.assign("at_patient_1", "at_coach_1", 1000.0)
      expect(mgr.get_assignment_at("at_patient_1", "coach", 2000.0)).to eq("at_coach_1")
    end

    it "returns nil before the first assignment" do
      mgr.add_member("at_coach_2", "coach", 5)
      mgr.assign("at_patient_2", "at_coach_2", 1000.0)
      expect(mgr.get_assignment_at("at_patient_2", "coach", 500.0)).to be_nil
    end

    it "returns the original member before a reassignment" do
      mgr.add_member("at_coach_old", "coach", 5)
      mgr.add_member("at_coach_new", "coach", 5)
      mgr.assign("at_patient_3", "at_coach_old", 1000.0)
      mgr.assign("at_patient_3", "at_coach_new", 3000.0)
      expect(mgr.get_assignment_at("at_patient_3", "coach", 2000.0)).to eq("at_coach_old")
    end

    it "returns the new member after a reassignment" do
      mgr.add_member("at_coach_prev", "coach", 5)
      mgr.add_member("at_coach_curr", "coach", 5)
      mgr.assign("at_patient_4", "at_coach_prev", 1000.0)
      mgr.assign("at_patient_4", "at_coach_curr", 3000.0)
      expect(mgr.get_assignment_at("at_patient_4", "coach", 4000.0)).to eq("at_coach_curr")
    end

    it "returns nil for a role the patient was never assigned" do
      mgr.add_member("at_coach_role", "coach", 5)
      mgr.assign("at_patient_5", "at_coach_role", 1000.0)
      expect(mgr.get_assignment_at("at_patient_5", "physician", 2000.0)).to be_nil
    end
  end
end
