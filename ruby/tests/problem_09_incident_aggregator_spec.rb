load_problem("09_incident_aggregator")

RSpec.describe IncidentAggregator do
  let(:agg) { described_class.new }

  # Pre-seeded aggregator:
  #   Reports: r1 (shooting/downtown/T0), r2 (shooting/downtown/T1),
  #            r3 (car-crash/midtown/T2)
  #   Incidents: inc-001 (shooting/downtown) containing r1 and r2
  #   r3 is unassigned.
  let(:seeded_agg) do
    a = described_class.new
    a.ingest_report("r1", "radio-north", "shooting", "downtown", T0)
    a.ingest_report("r2", "radio-south", "shooting", "downtown", T1)
    a.ingest_report("r3", "social-feed", "car-crash", "midtown", T2)
    a.create_incident("inc-001", "shooting", "downtown")
    a.add_report_to_incident("inc-001", "r1")
    a.add_report_to_incident("inc-001", "r2")
    a
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Report ingestion
  # ---------------------------------------------------------------------------
  describe "#ingest_report" do
    it "stores and returns the report" do
      r = agg.ingest_report("r_store", "radio-north", "shooting", "downtown", T0)
      expect(r[:report_id]).to eq("r_store")
      expect(r[:source_id]).to eq("radio-north")
      expect(r[:event_type]).to eq("shooting")
      expect(r[:location_key]).to eq("downtown")
      expect(r[:ts]).to eq(T0)
      expect(r[:incident_id]).to be_nil
    end

    it "raises ArgumentError on a duplicate report_id" do
      agg.ingest_report("r_dup", "src-a", "fire", "east", T0)
      expect { agg.ingest_report("r_dup", "src-b", "fire", "east", T1) }.to raise_error(ArgumentError)
    end

    it "returns nil from get_report when missing" do
      expect(agg.get_report("nonexistent")).to be_nil
    end

    it "returns the stored report via get_report" do
      r = seeded_agg.get_report("r1")
      expect(r).not_to be_nil
      expect(r[:event_type]).to eq("shooting")
    end
  end

  describe "#get_reports" do
    it "returns all reports when unfiltered" do
      expect(seeded_agg.get_reports.size).to eq(3)
    end

    it "filters by location" do
      reports = seeded_agg.get_reports(location_key: "downtown")
      expect(reports.size).to eq(2)
      expect(reports).to all(include(location_key: "downtown"))
    end

    it "filters by event_type" do
      reports = seeded_agg.get_reports(event_type: "car-crash")
      expect(reports.size).to eq(1)
      expect(reports.first[:report_id]).to eq("r3")
    end

    it "applies both filters together" do
      reports = seeded_agg.get_reports(location_key: "downtown", event_type: "shooting")
      expect(reports.size).to eq(2)
    end

    it "returns empty when nothing matches" do
      expect(seeded_agg.get_reports(location_key: "mars")).to eq([])
    end

    it "is sorted by ts ascending" do
      agg.ingest_report("r_sort_b", "src", "fire", "zone-1", T1)
      agg.ingest_report("r_sort_a", "src", "fire", "zone-1", T0)
      tss = agg.get_reports.map { |r| r[:ts] }
      expect(tss).to eq(tss.sort)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Manual incident grouping
  # ---------------------------------------------------------------------------
  describe "#create_incident" do
    it "creates an empty incident" do
      inc = agg.create_incident("inc_create_test", "fire", "east-side")
      expect(inc[:incident_id]).to eq("inc_create_test")
      expect(inc[:event_type]).to eq("fire")
      expect(inc[:location_key]).to eq("east-side")
      expect(inc[:report_count]).to eq(0)
      expect(inc[:report_ids]).to eq([])
      expect(inc[:latest_ts]).to be_nil
    end

    it "raises ArgumentError on a duplicate incident_id" do
      expect { seeded_agg.create_incident("inc-001", "shooting", "downtown") }.to raise_error(ArgumentError)
    end
  end

  describe "#add_report_to_incident" do
    it "assigns the report" do
      agg.ingest_report("r_assign", "src", "fire", "east", T0)
      agg.create_incident("inc_assign_test", "fire", "east")
      agg.add_report_to_incident("inc_assign_test", "r_assign")
      inc = agg.get_incident("inc_assign_test")
      expect(inc[:report_ids]).to include("r_assign")
      expect(inc[:report_count]).to eq(1)
      expect(inc[:latest_ts]).to eq(T0)
    end

    it "updates the report's incident_id" do
      agg.ingest_report("r_update", "src", "fire", "east", T0)
      agg.create_incident("inc_update_test", "fire", "east")
      agg.add_report_to_incident("inc_update_test", "r_update")
      expect(agg.get_report("r_update")[:incident_id]).to eq("inc_update_test")
    end

    it "raises ArgumentError when the report is already assigned" do
      # r1 is already in inc-001
      expect { seeded_agg.add_report_to_incident("inc-001", "r1") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown incident" do
      agg.ingest_report("r_bad_inc", "src", "fire", "east", T0)
      expect { agg.add_report_to_incident("nonexistent_inc", "r_bad_inc") }.to raise_error(KeyError)
    end

    it "raises KeyError for an unknown report" do
      expect { seeded_agg.add_report_to_incident("inc-001", "nonexistent_report") }.to raise_error(KeyError)
    end

    it "keeps report_ids sorted by ts" do
      agg.ingest_report("r_ts_b", "src", "fire", "east", T1)
      agg.ingest_report("r_ts_a", "src", "fire", "east", T0)
      agg.create_incident("inc_ts_test", "fire", "east")
      agg.add_report_to_incident("inc_ts_test", "r_ts_b")
      agg.add_report_to_incident("inc_ts_test", "r_ts_a")
      inc = agg.get_incident("inc_ts_test")
      expect(inc[:report_ids]).to eq(%w[r_ts_a r_ts_b])
    end

    it "updates latest_ts to the newest report" do
      agg.ingest_report("r_lt_a", "src", "fire", "east", T0)
      agg.ingest_report("r_lt_b", "src", "fire", "east", T2)
      agg.create_incident("inc_lt_test", "fire", "east")
      agg.add_report_to_incident("inc_lt_test", "r_lt_a")
      agg.add_report_to_incident("inc_lt_test", "r_lt_b")
      expect(agg.get_incident("inc_lt_test")[:latest_ts]).to eq(T2)
    end
  end

  describe "#get_incident" do
    it "returns nil when not found" do
      expect(agg.get_incident("ghost")).to be_nil
    end

    it "returns the incident" do
      inc = seeded_agg.get_incident("inc-001")
      expect(inc).not_to be_nil
      expect(inc[:report_count]).to eq(2)
    end
  end

  describe "#get_unassigned_reports" do
    it "returns unassigned reports" do
      unassigned = seeded_agg.get_unassigned_reports
      expect(unassigned.size).to eq(1)
      expect(unassigned.first[:report_id]).to eq("r3")
    end

    it "is empty once everything is assigned" do
      agg.ingest_report("r_all", "src", "fire", "east", T0)
      agg.create_incident("inc_all_test", "fire", "east")
      agg.add_report_to_incident("inc_all_test", "r_all")
      expect(agg.get_unassigned_reports).to eq([])
    end

    it "is sorted by ts ascending" do
      agg.ingest_report("r_ua_b", "src", "fire", "east", T1)
      agg.ingest_report("r_ua_a", "src", "fire", "east", T0)
      tss = agg.get_unassigned_reports.map { |r| r[:ts] }
      expect(tss).to eq(tss.sort)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Automatic deduplication
  # ---------------------------------------------------------------------------
  describe "#auto_ingest_report" do
    it "creates a new incident when there is no match" do
      incident_id = agg.auto_ingest_report("r_auto_new", "src", "fire", "east-side", T0, 120)
      expect(incident_id).not_to be_nil
      inc = agg.get_incident(incident_id)
      expect(inc[:report_count]).to eq(1)
      expect(inc[:report_ids]).to include("r_auto_new")
    end

    it "merges into an existing incident within the window" do
      # inc-001 latest_ts = T1; new ts = T2; T2 - T1 = 60s <= 120s
      result = seeded_agg.auto_ingest_report("r_merge", "radio-east", "shooting", "downtown", T2, 120)
      expect(result).to eq("inc-001")
      expect(seeded_agg.get_incident("inc-001")[:report_count]).to eq(3)
    end

    it "does not merge across a different event_type" do
      result = seeded_agg.auto_ingest_report("r_diff_type", "src", "car-crash", "downtown", T2, 120)
      expect(result).not_to eq("inc-001")
    end

    it "does not merge across a different location" do
      result = seeded_agg.auto_ingest_report("r_diff_loc", "src", "shooting", "uptown", T2, 120)
      expect(result).not_to eq("inc-001")
    end

    it "does not merge outside the time window" do
      # inc-001 latest_ts = T1; new ts = T4; T4 - T1 = 540s > 120s
      result = seeded_agg.auto_ingest_report("r_expired", "src", "shooting", "downtown", T4, 120)
      expect(result).not_to eq("inc-001")
    end

    it "stores the new report" do
      incident_id = agg.auto_ingest_report("r_stored", "src", "fire", "east", T0, 60)
      expect(agg.get_report("r_stored")).not_to be_nil
    end

    it "picks the most recently active matching incident" do
      # Two incidents for the same type+location, both within the window
      agg.ingest_report("ra1", "src", "shooting", "downtown", T0)
      agg.ingest_report("ra2", "src", "shooting", "downtown", T1)
      agg.create_incident("inc_older", "shooting", "downtown")
      agg.add_report_to_incident("inc_older", "ra1") # latest_ts = T0
      agg.create_incident("inc_newer", "shooting", "downtown")
      agg.add_report_to_incident("inc_newer", "ra2") # latest_ts = T1

      # T2 - T0 = 120s <= 300s and T2 - T1 = 60s <= 300s -> both active
      result = agg.auto_ingest_report("r_pick", "src", "shooting", "downtown", T2, 300)
      # inc_newer is more recently active (T1 > T0)
      expect(result).to eq("inc_newer")
    end

    it "generates non-clashing auto ids" do
      id1 = agg.auto_ingest_report("r_id1", "src", "fire", "west", T0, 0)
      id2 = agg.auto_ingest_report("r_id2", "src", "fire", "east", T1, 0)
      expect(id1).not_to eq(id2)
    end
  end

  describe "#get_active_incidents" do
    it "returns active incidents" do
      # inc-001 latest_ts=T1; T2-T1=60s <= 120s -> active
      active = seeded_agg.get_active_incidents(T2, 120)
      expect(active.any? { |i| i[:incident_id] == "inc-001" }).to be true
    end

    it "excludes stale incidents" do
      # inc-001 latest_ts=T1; T4-T1=540s > 120s -> not active
      active = seeded_agg.get_active_incidents(T4, 120)
      expect(active.any? { |i| i[:incident_id] == "inc-001" }).to be false
    end

    it "excludes empty incidents" do
      agg.create_incident("inc_empty_active", "fire", "north")
      active = agg.get_active_incidents(T0, 300)
      expect(active.any? { |i| i[:incident_id] == "inc_empty_active" }).to be false
    end

    it "sorts by latest_ts descending" do
      agg.ingest_report("ri1", "src", "fire", "east", T0)
      agg.ingest_report("ri2", "src", "fire", "west", T2)
      agg.create_incident("inc_sort_a", "fire", "east")
      agg.add_report_to_incident("inc_sort_a", "ri1") # latest_ts=T0
      agg.create_incident("inc_sort_b", "fire", "west")
      agg.add_report_to_incident("inc_sort_b", "ri2") # latest_ts=T2

      active = agg.get_active_incidents(T3, 600)
      ids = active.map { |i| i[:incident_id] }
      expect(ids.index("inc_sort_b")).to be < ids.index("inc_sort_a")
    end
  end
end
