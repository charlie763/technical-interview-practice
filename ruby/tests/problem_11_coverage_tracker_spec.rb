load_problem("11_coverage_tracker")

RSpec.describe CoverageTracker do
  let(:ct) { described_class.new }

  # Pre-seeded tracker:
  #   sta-seed-1  North Tower  downtown  last hb: T0
  #   sta-seed-2  South Tower  downtown  last hb: T1
  #   sta-seed-3  East Hub     eastside  (never sent a heartbeat)
  let(:seeded_ct) do
    c = described_class.new
    c.register_station("sta-seed-1", "North Tower", "downtown")
    c.register_station("sta-seed-2", "South Tower", "downtown")
    c.register_station("sta-seed-3", "East Hub", "eastside")
    c.record_heartbeat("sta-seed-1", T0)
    c.record_heartbeat("sta-seed-2", T1)
    # sta-seed-3 intentionally has no heartbeat
    c
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Station registration and heartbeats
  # ---------------------------------------------------------------------------
  describe "#register_station" do
    it "stores and returns the station" do
      s = ct.register_station("sta_reg_test", "Tower", "north")
      expect(s[:station_id]).to eq("sta_reg_test")
      expect(s[:name]).to eq("Tower")
      expect(s[:region]).to eq("north")
    end

    it "raises ArgumentError on a duplicate" do
      expect { seeded_ct.register_station("sta-seed-1", "Dup", "downtown") }.to raise_error(ArgumentError)
    end
  end

  describe "#record_heartbeat" do
    it "updates the last heartbeat" do
      seeded_ct.record_heartbeat("sta-seed-1", T2)
      expect(seeded_ct.get_last_heartbeat("sta-seed-1")).to eq(T2)
    end

    it "raises KeyError for an unknown station" do
      expect { seeded_ct.record_heartbeat("ghost", T0) }.to raise_error(KeyError)
    end

    it "raises ArgumentError for a duplicate timestamp" do
      # sta-seed-1 last hb is T0; equal ts should be rejected
      expect { seeded_ct.record_heartbeat("sta-seed-1", T0) }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError for an earlier timestamp" do
      # sta-seed-2 last hb is T1; T0 < T1 should be rejected
      expect { seeded_ct.record_heartbeat("sta-seed-2", T0) }.to raise_error(ArgumentError)
    end

    it "accepts the first heartbeat" do
      # sta-seed-3 has never sent one
      seeded_ct.record_heartbeat("sta-seed-3", T0)
      expect(seeded_ct.get_last_heartbeat("sta-seed-3")).to eq(T0)
    end
  end

  describe "#get_last_heartbeat" do
    it "returns nil if never sent" do
      expect(seeded_ct.get_last_heartbeat("sta-seed-3")).to be_nil
    end

    it "returns the latest ts" do
      expect(seeded_ct.get_last_heartbeat("sta-seed-1")).to eq(T0)
    end

    it "raises KeyError for an unknown station" do
      expect { seeded_ct.get_last_heartbeat("ghost") }.to raise_error(KeyError)
    end
  end

  describe "#get_stations" do
    it "returns all stations when unfiltered" do
      expect(seeded_ct.get_stations.size).to eq(3)
    end

    it "filters by region" do
      downtown = seeded_ct.get_stations(region: "downtown")
      expect(downtown.size).to eq(2)
      expect(downtown).to all(include(region: "downtown"))
    end

    it "returns empty for an unknown region" do
      expect(seeded_ct.get_stations(region: "nowhere")).to eq([])
    end

    it "is sorted by station_id" do
      ids = seeded_ct.get_stations.map { |s| s[:station_id] }
      expect(ids).to eq(ids.sort)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Staleness detection and outage tracking
  # ---------------------------------------------------------------------------
  describe "#get_stale_stations" do
    it "treats a station with no heartbeat as stale" do
      stale_ids = seeded_ct.get_stale_stations(T2, 120).map { |s| s[:station_id] }
      expect(stale_ids).to include("sta-seed-3")
    end

    it "does not flag a recently-heard station as stale" do
      # sta-seed-2 hb=T1; T2-T1=60s <= 120s -> not stale
      stale_ids = seeded_ct.get_stale_stations(T2, 120).map { |s| s[:station_id] }
      expect(stale_ids).not_to include("sta-seed-2")
    end

    it "flags a station with an old heartbeat as stale" do
      # sta-seed-1 hb=T0; T3-T0=300s > 120s -> stale
      stale_ids = seeded_ct.get_stale_stations(T3, 120).map { |s| s[:station_id] }
      expect(stale_ids).to include("sta-seed-1")
    end

    it "flags everyone stale with a tiny threshold" do
      # 5s threshold: all stations stale at T4
      expect(seeded_ct.get_stale_stations(T4, 5).size).to eq(3)
    end

    it "is sorted by station_id" do
      ids = seeded_ct.get_stale_stations(T4, 5).map { |s| s[:station_id] }
      expect(ids).to eq(ids.sort)
    end
  end

  describe "#record_outage_start and #record_outage_end" do
    it "records an open outage" do
      ct.register_station("sta_out_test", "T", "north")
      ct.record_outage_start("sta_out_test", T0)
      outages = ct.get_outages("sta_out_test")
      expect(outages.size).to eq(1)
      expect(outages.first[:start_ts]).to eq(T0)
      expect(outages.first[:end_ts]).to be_nil
    end

    it "raises ArgumentError on a duplicate open outage" do
      ct.register_station("sta_dup_out", "T", "north")
      ct.record_outage_start("sta_dup_out", T0)
      expect { ct.record_outage_start("sta_dup_out", T1) }.to raise_error(ArgumentError)
    end

    it "closes the outage" do
      ct.register_station("sta_end_out", "T", "north")
      ct.record_outage_start("sta_end_out", T0)
      ct.record_outage_end("sta_end_out", T1)
      expect(ct.get_outages("sta_end_out").first[:end_ts]).to eq(T1)
    end

    it "raises ArgumentError when there's no open outage to end" do
      # sta-seed-1 has no open outage
      expect { seeded_ct.record_outage_end("sta-seed-1", T2) }.to raise_error(ArgumentError)
    end

    it "raises KeyError starting an outage on an unknown station" do
      expect { seeded_ct.record_outage_start("ghost", T0) }.to raise_error(KeyError)
    end

    it "raises KeyError ending an outage on an unknown station" do
      expect { seeded_ct.record_outage_end("ghost", T0) }.to raise_error(KeyError)
    end

    it "allows a second outage once the first is closed" do
      ct.register_station("sta_2nd_out", "T", "north")
      ct.record_outage_start("sta_2nd_out", T0)
      ct.record_outage_end("sta_2nd_out", T1)
      expect { ct.record_outage_start("sta_2nd_out", T2) }.not_to raise_error
      expect(ct.get_outages("sta_2nd_out").size).to eq(2)
    end

    it "returns outages sorted by start_ts" do
      ct.register_station("sta_sort_out", "T", "north")
      ct.record_outage_start("sta_sort_out", T0)
      ct.record_outage_end("sta_sort_out", T1)
      ct.record_outage_start("sta_sort_out", T2)
      ct.record_outage_end("sta_sort_out", T3)
      outages = ct.get_outages("sta_sort_out")
      expect(outages[0][:start_ts]).to eq(T0)
      expect(outages[1][:start_ts]).to eq(T2)
    end

    it "raises KeyError from get_outages on an unknown station" do
      expect { seeded_ct.get_outages("ghost") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Coverage analysis
  # ---------------------------------------------------------------------------
  describe "#get_region_coverage" do
    it "computes partial coverage" do
      # as_of=T2, threshold=90s
      # sta-seed-1: T2-T0=120s > 90s -> stale
      # sta-seed-2: T2-T1=60s  <= 90s -> healthy
      result = seeded_ct.get_region_coverage("downtown", T2, 90)
      expect(result[:region]).to eq("downtown")
      expect(result[:total]).to eq(2)
      expect(result[:healthy]).to eq(1)
      expect(result[:stale]).to eq(1)
      expect(result[:has_coverage]).to be true
    end

    it "computes full coverage" do
      # as_of=T2, threshold=300s -> both hbs are within window
      result = seeded_ct.get_region_coverage("downtown", T2, 300)
      expect(result[:healthy]).to eq(2)
      expect(result[:stale]).to eq(0)
      expect(result[:has_coverage]).to be true
    end

    it "reports no coverage when everyone is stale" do
      # as_of=T4, threshold=30s -> both hbs are way older than 30s
      result = seeded_ct.get_region_coverage("downtown", T4, 30)
      expect(result[:healthy]).to eq(0)
      expect(result[:has_coverage]).to be false
    end

    it "returns zeros for an empty region" do
      result = ct.get_region_coverage("ghost-region", T0, 60)
      expect(result[:total]).to eq(0)
      expect(result[:healthy]).to eq(0)
      expect(result[:has_coverage]).to be false
    end

    it "reports the correct total for a region" do
      result = seeded_ct.get_region_coverage("eastside", T0, 60)
      expect(result[:total]).to eq(1)
    end
  end

  describe "#get_outage_summary" do
    it "reports zeros when there are no outages" do
      summary = seeded_ct.get_outage_summary("sta-seed-1", T4)
      expect(summary[:station_id]).to eq("sta-seed-1")
      expect(summary[:total_outages]).to eq(0)
      expect(summary[:open_outage]).to be false
      expect(summary[:total_outage_secs]).to eq(0)
    end

    it "sums a closed outage's duration" do
      ct.register_station("sta_dur_test", "T", "north")
      ct.record_outage_start("sta_dur_test", T0)
      ct.record_outage_end("sta_dur_test", T2) # T2 - T0 = 120s
      summary = ct.get_outage_summary("sta_dur_test", T3)
      expect(summary[:total_outages]).to eq(1)
      expect(summary[:open_outage]).to be false
      expect(summary[:total_outage_secs]).to eq(120)
    end

    it "counts an open outage up to as_of_ts" do
      ct.register_station("sta_open_test", "T", "north")
      ct.record_outage_start("sta_open_test", T0)
      # T3 - T0 = 300s
      summary = ct.get_outage_summary("sta_open_test", T3)
      expect(summary[:open_outage]).to be true
      expect(summary[:total_outage_secs]).to eq(300)
    end

    it "accumulates duration across multiple closed outages" do
      ct.register_station("sta_cumul_test", "T", "north")
      # Outage 1: T0 -> T1 = 60s
      ct.record_outage_start("sta_cumul_test", T0)
      ct.record_outage_end("sta_cumul_test", T1)
      # Outage 2: T2 -> T3 = 180s
      ct.record_outage_start("sta_cumul_test", T2)
      ct.record_outage_end("sta_cumul_test", T3)
      summary = ct.get_outage_summary("sta_cumul_test", T4)
      expect(summary[:total_outages]).to eq(2)
      expect(summary[:open_outage]).to be false
      expect(summary[:total_outage_secs]).to eq(60 + 180) # 240s total
    end

    it "raises KeyError for an unknown station" do
      expect { seeded_ct.get_outage_summary("ghost", T0) }.to raise_error(KeyError)
    end
  end
end
