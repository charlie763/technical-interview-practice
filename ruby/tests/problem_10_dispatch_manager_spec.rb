load_problem("10_dispatch_manager")

RSpec.describe DispatchManager do
  let(:dm) { described_class.new }

  # Pre-seeded manager:
  #   Responders:
  #     unit-12  (shooting + robbery, capacity=3): has inc-seed-1 open
  #     unit-14  (car-crash + fire,   capacity=2): no assignments
  #   Incidents:
  #     inc-seed-1  shooting  sev=5  T0  -> assigned to unit-12
  #     inc-seed-2  shooting  sev=3  T1  -> unassigned
  #     inc-seed-3  car-crash sev=4  T2  -> unassigned
  let(:seeded_dm) do
    d = described_class.new
    d.register_responder("unit-12", "Alpha Team", %w[shooting robbery], 3)
    d.register_responder("unit-14", "Beta Team", %w[car-crash fire], 2)
    d.add_incident("inc-seed-1", "shooting", 5, T0)
    d.add_incident("inc-seed-2", "shooting", 3, T1)
    d.add_incident("inc-seed-3", "car-crash", 4, T2)
    d.assign_incident("inc-seed-1", "unit-12")
    d
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Registration and basic queries
  # ---------------------------------------------------------------------------
  describe "#register_responder" do
    it "stores and returns the responder" do
      r = dm.register_responder("unit_reg_test", "Gamma", ["fire"], 2)
      expect(r[:responder_id]).to eq("unit_reg_test")
      expect(r[:name]).to eq("Gamma")
      expect(r[:subscribed_types]).to eq(["fire"])
      expect(r[:capacity]).to eq(2)
    end

    it "raises ArgumentError on a duplicate" do
      expect { seeded_dm.register_responder("unit-12", "Duplicate", ["fire"], 1) }.to raise_error(ArgumentError)
    end
  end

  describe "#add_incident" do
    it "stores and returns the incident" do
      inc = dm.add_incident("inc_add_test", "fire", 2, T0)
      expect(inc[:incident_id]).to eq("inc_add_test")
      expect(inc[:incident_type]).to eq("fire")
      expect(inc[:severity]).to eq(2)
      expect(inc[:responder_id]).to be_nil
      expect(inc[:resolved]).to be false
    end

    it "raises ArgumentError on a duplicate" do
      expect { seeded_dm.add_incident("inc-seed-1", "fire", 1, T0) }.to raise_error(ArgumentError)
    end
  end

  describe "#get_incidents_for_responder" do
    it "returns only subscribed types" do
      incidents = seeded_dm.get_incidents_for_responder("unit-12")
      expect(incidents).to all(satisfy { |i| %w[shooting robbery].include?(i[:incident_type]) })
    end

    it "sorts by severity descending then ts ascending" do
      # inc-seed-1 sev=5, inc-seed-2 sev=3 — both are shooting
      ids = seeded_dm.get_incidents_for_responder("unit-12").map { |i| i[:incident_id] }
      expect(ids[0]).to eq("inc-seed-1") # higher severity first
      expect(ids[1]).to eq("inc-seed-2")
    end

    it "sorts by ts when severities are equal" do
      dm.register_responder("u_ts_test", "T", ["fire"], 5)
      dm.add_incident("inc_ts_early", "fire", 3, T0)
      dm.add_incident("inc_ts_late", "fire", 3, T1)
      ids = dm.get_incidents_for_responder("u_ts_test").map { |i| i[:incident_id] }
      expect(ids).to eq(%w[inc_ts_early inc_ts_late])
    end

    it "raises KeyError for an unknown responder" do
      expect { seeded_dm.get_incidents_for_responder("ghost") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Assignment and resolution
  # ---------------------------------------------------------------------------
  describe "#assign_incident" do
    it "sets the responder_id" do
      seeded_dm.assign_incident("inc-seed-2", "unit-12")
      open_ids = seeded_dm.get_open_assignments("unit-12").map { |i| i[:incident_id] }
      expect(open_ids).to include("inc-seed-2")
    end

    it "raises KeyError for an unknown incident" do
      expect { seeded_dm.assign_incident("ghost-inc", "unit-12") }.to raise_error(KeyError)
    end

    it "raises KeyError for an unknown responder" do
      expect { seeded_dm.assign_incident("inc-seed-2", "ghost-unit") }.to raise_error(KeyError)
    end

    it "raises ArgumentError when already assigned" do
      # inc-seed-1 is already assigned to unit-12
      expect { seeded_dm.assign_incident("inc-seed-1", "unit-12") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError at capacity" do
      dm.register_responder("cap_unit", "Cap", ["fire"], 1)
      dm.add_incident("cap_inc_1", "fire", 1, T0)
      dm.add_incident("cap_inc_2", "fire", 1, T1)
      dm.assign_incident("cap_inc_1", "cap_unit")
      expect { dm.assign_incident("cap_inc_2", "cap_unit") }.to raise_error(ArgumentError)
    end
  end

  describe "#resolve_incident" do
    it "marks resolved and removes it from open assignments" do
      seeded_dm.resolve_incident("inc-seed-1")
      open_ids = seeded_dm.get_open_assignments("unit-12").map { |i| i[:incident_id] }
      expect(open_ids).not_to include("inc-seed-1")
    end

    it "raises ArgumentError when already resolved" do
      seeded_dm.resolve_incident("inc-seed-1")
      expect { seeded_dm.resolve_incident("inc-seed-1") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown incident" do
      expect { seeded_dm.resolve_incident("ghost") }.to raise_error(KeyError)
    end

    it "frees capacity for the next assignment" do
      dm.register_responder("cap2_unit", "Cap2", ["fire"], 1)
      dm.add_incident("cap2_inc_1", "fire", 1, T0)
      dm.add_incident("cap2_inc_2", "fire", 1, T1)
      dm.assign_incident("cap2_inc_1", "cap2_unit")
      dm.resolve_incident("cap2_inc_1")
      expect { dm.assign_incident("cap2_inc_2", "cap2_unit") }.not_to raise_error
    end
  end

  describe "#get_open_assignments" do
    it "returns open assignments" do
      open_list = seeded_dm.get_open_assignments("unit-12")
      expect(open_list.size).to eq(1)
      expect(open_list.first[:incident_id]).to eq("inc-seed-1")
    end

    it "excludes resolved incidents" do
      seeded_dm.resolve_incident("inc-seed-1")
      expect(seeded_dm.get_open_assignments("unit-12")).to eq([])
    end

    it "raises KeyError for an unknown responder" do
      expect { seeded_dm.get_open_assignments("ghost") }.to raise_error(KeyError)
    end

    it "sorts by severity descending then ts ascending" do
      dm.register_responder("u_sort", "S", ["fire"], 5)
      dm.add_incident("inc_sort_low", "fire", 2, T0)
      dm.add_incident("inc_sort_high", "fire", 5, T1)
      dm.assign_incident("inc_sort_low", "u_sort")
      dm.assign_incident("inc_sort_high", "u_sort")
      expect(dm.get_open_assignments("u_sort").first[:incident_id]).to eq("inc_sort_high")
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Auto-assignment
  # ---------------------------------------------------------------------------
  describe "#auto_assign" do
    it "assigns to the eligible responder" do
      # inc-seed-3 is car-crash -> only unit-14 is subscribed
      result = seeded_dm.auto_assign("inc-seed-3")
      expect(result).to eq("unit-14")
      open_ids = seeded_dm.get_open_assignments("unit-14").map { |i| i[:incident_id] }
      expect(open_ids).to include("inc-seed-3")
    end

    it "raises KeyError for an unknown incident" do
      expect { seeded_dm.auto_assign("ghost") }.to raise_error(KeyError)
    end

    it "raises ArgumentError when already assigned" do
      expect { seeded_dm.auto_assign("inc-seed-1") }.to raise_error(ArgumentError) # already assigned to unit-12
    end

    it "raises ArgumentError when no responder is eligible" do
      dm.register_responder("only_unit", "Only", ["shooting"], 1)
      dm.add_incident("inc_no_sub", "fire", 1, T0)
      expect { dm.auto_assign("inc_no_sub") }.to raise_error(ArgumentError)
    end

    it "excludes a responder at full capacity" do
      dm.register_responder("full_unit", "Full", ["fire"], 1)
      dm.register_responder("open_unit", "Open", ["fire"], 2)
      dm.add_incident("inc_cap_fill", "fire", 1, T0)
      dm.add_incident("inc_cap_new", "fire", 1, T1)
      dm.assign_incident("inc_cap_fill", "full_unit")
      result = dm.auto_assign("inc_cap_new")
      expect(result).to eq("open_unit")
    end

    it "picks the least loaded responder" do
      dm.register_responder("u_loaded", "Loaded", ["fire"], 3)
      dm.register_responder("u_free", "Free", ["fire"], 3)
      dm.add_incident("inc_load_seed", "fire", 1, T0)
      dm.add_incident("inc_load_new", "fire", 1, T1)
      # Give u_loaded one open incident
      dm.assign_incident("inc_load_seed", "u_loaded")
      result = dm.auto_assign("inc_load_new")
      expect(result).to eq("u_free") # fewer open assignments
    end

    it "tie-breaks by highest capacity" do
      # Equal open assignments (0 each); tiebreak -> higher capacity wins
      dm.register_responder("u_low_cap", "Low", ["fire"], 1)
      dm.register_responder("u_high_cap", "High", ["fire"], 5)
      dm.add_incident("inc_cap_tb", "fire", 1, T0)
      result = dm.auto_assign("inc_cap_tb")
      expect(result).to eq("u_high_cap")
    end

    it "delegates to assign_incident rather than duplicating logic" do
      dm.register_responder("u_delegate", "D", ["fire"], 1)
      dm.add_incident("inc_delegate_1", "fire", 1, T0)
      dm.add_incident("inc_delegate_2", "fire", 1, T1)
      dm.auto_assign("inc_delegate_1")
      # Capacity should now be full; assigning again must raise via assign_incident
      expect { dm.auto_assign("inc_delegate_2") }.to raise_error(ArgumentError)
    end
  end

  describe "#get_dispatch_summary" do
    it "returns all responders" do
      ids = seeded_dm.get_dispatch_summary.map { |s| s[:responder_id] }
      expect(ids).to include("unit-12", "unit-14")
    end

    it "is sorted by responder_id" do
      ids = seeded_dm.get_dispatch_summary.map { |s| s[:responder_id] }
      expect(ids).to eq(ids.sort)
    end

    it "reports open_count and available_capacity" do
      # unit-12 has 1 open assignment; capacity=3
      u12 = seeded_dm.get_dispatch_summary.find { |s| s[:responder_id] == "unit-12" }
      expect(u12[:open_count]).to eq(1)
      expect(u12[:available_capacity]).to eq(2)
    end

    it "reports zero load correctly" do
      # unit-14 has no assignments
      u14 = seeded_dm.get_dispatch_summary.find { |s| s[:responder_id] == "unit-14" }
      expect(u14[:open_count]).to eq(0)
      expect(u14[:available_capacity]).to eq(2)
    end
  end
end
