load_problem("01_geofence_alert_engine")

# White-box helpers: reach into instance state directly rather than through a
# later Part's public method, so earlier-Part examples stay independent.
def zones_of(tracker)
  tracker.instance_variable_get(:@zones)
end

def assets_of(tracker)
  tracker.instance_variable_get(:@assets)
end

def alert_log_of(tracker)
  tracker.instance_variable_get(:@alert_log)
end

RSpec.describe GeofenceTracker do
  # A tracker with two adjacent zones and two assets.
  let(:tracker) do
    t = described_class.new
    # Warehouse: lat [35.00, 35.10], lng [-106.70, -106.60]
    t.add_zone("warehouse", "Warehouse A", 35.00, 35.10, -106.70, -106.60)
    # Loading dock: lat [35.10, 35.20], lng [-106.70, -106.60]
    t.add_zone("loading_dock", "Loading Dock", 35.10, 35.20, -106.70, -106.60)
    t.add_asset("forklift_1", "Forklift #1")
    t.add_asset("drone_1", "Drone #1")
    t
  end

  # ---------------------------------------------------------------------------
  # PART 1 — in_zone?
  # ---------------------------------------------------------------------------
  describe "#in_zone?" do
    def zone(min_lat, max_lat, min_lng, max_lng)
      { id: "z1", name: "Z", bounds: { min_lat: min_lat, max_lat: max_lat, min_lng: min_lng, max_lng: max_lng } }
    end

    def asset(lat, lng)
      { id: "a1", name: "A", lat: lat, lng: lng, zone_id: nil }
    end

    it "returns true when inside the box" do
      expect(tracker.in_zone?(asset(35.05, -106.65), zone(35.0, 35.1, -106.7, -106.6))).to eq(true)
    end

    it "returns true on the min corner (inclusive)" do
      expect(tracker.in_zone?(asset(35.0, -106.7), zone(35.0, 35.1, -106.7, -106.6))).to eq(true)
    end

    it "returns true on the max corner (inclusive)" do
      expect(tracker.in_zone?(asset(35.1, -106.6), zone(35.0, 35.1, -106.7, -106.6))).to eq(true)
    end

    it "returns false when outside on lat" do
      expect(tracker.in_zone?(asset(35.15, -106.65), zone(35.0, 35.1, -106.7, -106.6))).to eq(false)
    end

    it "returns false when outside on lng" do
      expect(tracker.in_zone?(asset(35.05, -106.5), zone(35.0, 35.1, -106.7, -106.6))).to eq(false)
    end

    it "returns false when lat is nil" do
      expect(tracker.in_zone?(asset(nil, -106.65), zone(35.0, 35.1, -106.7, -106.6))).to eq(false)
    end

    it "returns false when lng is nil" do
      expect(tracker.in_zone?(asset(35.05, nil), zone(35.0, 35.1, -106.7, -106.6))).to eq(false)
    end

    it "returns false when both are nil" do
      expect(tracker.in_zone?(asset(nil, nil), zone(35.0, 35.1, -106.7, -106.6))).to eq(false)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — current_zone_id
  # ---------------------------------------------------------------------------
  describe "#current_zone_id" do
    it "finds an asset in the warehouse" do
      assets_of(tracker)["forklift_1"][:lat] = 35.05
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      expect(tracker.current_zone_id("forklift_1")).to eq("warehouse")
    end

    it "finds an asset in the loading dock" do
      assets_of(tracker)["forklift_1"][:lat] = 35.15
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      expect(tracker.current_zone_id("forklift_1")).to eq("loading_dock")
    end

    it "returns nil when outside all zones" do
      assets_of(tracker)["forklift_1"][:lat] = 36.0
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      expect(tracker.current_zone_id("forklift_1")).to be_nil
    end

    it "returns nil when the asset has no location" do
      expect(tracker.current_zone_id("forklift_1")).to be_nil
    end

    it "returns nil for an unknown asset id" do
      expect(tracker.current_zone_id("ghost")).to be_nil
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — process_location_update
  # ---------------------------------------------------------------------------
  describe "#process_location_update" do
    it "fires no alert on a first update outside any zone" do
      alerts = tracker.process_location_update("forklift_1", 36.0, -106.65, "t1")
      expect(alerts).to eq([])
      expect(assets_of(tracker)["forklift_1"][:lat]).to eq(36.0)
      expect(assets_of(tracker)["forklift_1"][:zone_id]).to be_nil
    end

    it "fires a matching rule on zone entry" do
      tracker.add_alert_rule("rule_entry", nil, "warehouse", nil)
      alerts = tracker.process_location_update("forklift_1", 35.05, -106.65, "t1")
      expect(alerts.length).to eq(1)
      expect(alerts[0]).to eq(
        rule_id: "rule_entry",
        asset_id: "forklift_1",
        from_zone_id: nil,
        to_zone_id: "warehouse",
        timestamp: "t1"
      )
    end

    it "fires a matching rule on zone exit" do
      assets_of(tracker)["forklift_1"][:lat] = 35.05
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      assets_of(tracker)["forklift_1"][:zone_id] = "warehouse"
      tracker.add_alert_rule("rule_exit", "warehouse", nil, nil)
      alerts = tracker.process_location_update("forklift_1", 36.0, -106.65, "t2")
      expect(alerts.length).to eq(1)
      expect(alerts[0][:from_zone_id]).to eq("warehouse")
      expect(alerts[0][:to_zone_id]).to be_nil
    end

    it "fires a matching rule on a zone-to-zone transition" do
      assets_of(tracker)["forklift_1"][:lat] = 35.05
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      assets_of(tracker)["forklift_1"][:zone_id] = "warehouse"
      tracker.add_alert_rule("rule_wh_to_dock", "warehouse", "loading_dock", nil)
      alerts = tracker.process_location_update("forklift_1", 35.15, -106.65, "t3")
      expect(alerts.length).to eq(1)
      expect(alerts[0][:from_zone_id]).to eq("warehouse")
      expect(alerts[0][:to_zone_id]).to eq("loading_dock")
    end

    it "fires nothing when the zone is unchanged" do
      assets_of(tracker)["forklift_1"][:lat] = 35.05
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      assets_of(tracker)["forklift_1"][:zone_id] = "warehouse"
      tracker.add_alert_rule("rule_any", nil, nil, nil)
      alerts = tracker.process_location_update("forklift_1", 35.06, -106.65, "t4")
      expect(alerts).to eq([])
    end

    it "ignores an asset-specific rule for other assets" do
      tracker.add_alert_rule("rule_drone_only", nil, "warehouse", "drone_1")
      alerts = tracker.process_location_update("forklift_1", 35.05, -106.65, "t5")
      expect(alerts).to eq([])
    end

    it "fires an asset-specific rule for the correct asset" do
      tracker.add_alert_rule("rule_forklift", nil, "warehouse", "forklift_1")
      alerts = tracker.process_location_update("forklift_1", 35.05, -106.65, "t6")
      expect(alerts.length).to eq(1)
    end

    it "fires every matching rule" do
      tracker.add_alert_rule("rule_a", nil, "warehouse", nil)
      tracker.add_alert_rule("rule_b", nil, nil, nil)
      alerts = tracker.process_location_update("forklift_1", 35.05, -106.65, "t7")
      expect(alerts.length).to eq(2)
    end

    it "appends triggered alerts to the alert log" do
      tracker.add_alert_rule("rule_1", nil, "warehouse", nil)
      tracker.process_location_update("forklift_1", 35.05, -106.65, "t8")
      expect(alert_log_of(tracker).length).to eq(1)
    end

    it "raises KeyError for an unknown asset" do
      expect { tracker.process_location_update("ghost_asset", 35.05, -106.65, "t9") }.to raise_error(KeyError)
    end

    it "updates lat/lng even when the zone doesn't change" do
      assets_of(tracker)["forklift_1"][:lat] = 35.05
      assets_of(tracker)["forklift_1"][:lng] = -106.65
      assets_of(tracker)["forklift_1"][:zone_id] = "warehouse"
      tracker.process_location_update("forklift_1", 35.06, -106.64, "t10")
      expect(assets_of(tracker)["forklift_1"][:lat]).to eq(35.06)
      expect(assets_of(tracker)["forklift_1"][:lng]).to eq(-106.64)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — CRUD helpers
  # ---------------------------------------------------------------------------
  describe "#add_zone" do
    it "adds a zone" do
      z = tracker.add_zone("yard", "Yard", 35.3, 35.4, -106.7, -106.6)
      expect(zones_of(tracker)["yard"]).to eq(z)
      expect(z[:name]).to eq("Yard")
    end

    it "raises ArgumentError on a duplicate zone_id" do
      expect { tracker.add_zone("warehouse", "Duplicate", 0, 1, 0, 1) }.to raise_error(ArgumentError)
    end
  end

  describe "#remove_zone" do
    it "removes the zone" do
      tracker.remove_zone("warehouse")
      expect(zones_of(tracker)).not_to have_key("warehouse")
    end

    it "clears zone_id on assets that were in it" do
      assets_of(tracker)["forklift_1"][:zone_id] = "warehouse"
      tracker.remove_zone("warehouse")
      expect(assets_of(tracker)["forklift_1"][:zone_id]).to be_nil
    end

    it "does not affect assets in other zones" do
      assets_of(tracker)["forklift_1"][:zone_id] = "loading_dock"
      tracker.remove_zone("warehouse")
      expect(assets_of(tracker)["forklift_1"][:zone_id]).to eq("loading_dock")
    end

    it "raises KeyError for a missing zone" do
      expect { tracker.remove_zone("nonexistent") }.to raise_error(KeyError)
    end
  end

  describe "#add_asset" do
    it "adds an asset with no initial location" do
      a = tracker.add_asset("scanner_1", "Scanner #1")
      expect(assets_of(tracker)["scanner_1"]).to eq(a)
      expect(a[:lat]).to be_nil
      expect(a[:lng]).to be_nil
      expect(a[:zone_id]).to be_nil
    end

    it "raises ArgumentError on a duplicate asset_id" do
      expect { tracker.add_asset("forklift_1", "Duplicate") }.to raise_error(ArgumentError)
    end
  end

  describe "#add_alert_rule" do
    it "adds a rule" do
      r = tracker.add_alert_rule("r1", "warehouse", "loading_dock", nil)
      expect(r[:id]).to eq("r1")
    end

    it "raises ArgumentError on a duplicate rule_id" do
      tracker.add_alert_rule("r1", nil, nil, nil)
      expect { tracker.add_alert_rule("r1", nil, nil, nil) }.to raise_error(ArgumentError)
    end
  end
end
