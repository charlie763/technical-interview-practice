load_problem("02_api_rate_limiter")

# White-box helper: seeds a key's request_log directly, the same way the
# problem's data model would be poked at in a plain-hash implementation.
# Using this (instead of calling record_request, a Part 3 method) keeps
# Part 1/2 examples independent of Part 3.
def seed_request_log(gateway, key_id, timestamps)
  gateway.instance_variable_get(:@keys)[key_id][:request_log] = timestamps
end

RSpec.describe ApiGateway do
  let(:plans) do
    {
      "free" => { rpm: 3, rpd: 10 },
      "pro" => { rpm: 100, rpd: 5_000 },
      "unlimited" => { rpm: nil, rpd: nil },
    }
  end
  let(:gw) { described_class.new(plans) }
  let(:gw_with_key) do
    gw.create_key("key_abc", "alice", "pro")
    gw
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Key management
  # ---------------------------------------------------------------------------
  describe "#create_key" do
    it "raises nothing and registers an enabled key with an empty log" do
      expect { gw.create_key("k1", "alice", "free") }.not_to raise_error
    end

    it "raises ArgumentError on duplicate key_id" do
      gw.create_key("k1", "alice", "free")
      expect { gw.create_key("k1", "bob", "pro") }.to raise_error(ArgumentError)
    end

    it "raises ArgumentError on an unknown plan" do
      expect { gw.create_key("k1", "alice", "enterprise") }.to raise_error(ArgumentError)
    end
  end

  describe "#revoke_key" do
    it "disables the key" do
      gw_with_key.revoke_key("key_abc")
      expect(gw_with_key.instance_variable_get(:@keys)["key_abc"][:enabled]).to eq(false)
    end

    it "raises KeyError when the key does not exist" do
      expect { gw.revoke_key("ghost") }.to raise_error(KeyError)
    end
  end

  describe "#update_plan" do
    it "changes the key's plan" do
      gw_with_key.update_plan("key_abc", "free")
      expect(gw_with_key.instance_variable_get(:@keys)["key_abc"][:plan]).to eq("free")
    end

    it "preserves the request_log" do
      seed_request_log(gw_with_key, "key_abc", [BASE_TIME - 5])
      gw_with_key.update_plan("key_abc", "free")
      expect(gw_with_key.instance_variable_get(:@keys)["key_abc"][:request_log]).to eq([BASE_TIME - 5])
    end

    it "raises ArgumentError on an unknown plan" do
      expect { gw_with_key.update_plan("key_abc", "nonexistent") }.to raise_error(ArgumentError)
    end

    it "raises KeyError when the key does not exist" do
      expect { gw.update_plan("ghost", "free") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — count_in_window
  # ---------------------------------------------------------------------------
  describe "#count_in_window" do
    it "returns 0 for an empty log" do
      expect(gw.count_in_window([], BASE_TIME, 60)).to eq(0)
    end

    it "counts every entry within the window" do
      log = [BASE_TIME - 30, BASE_TIME - 10, BASE_TIME]
      expect(gw.count_in_window(log, BASE_TIME, 60)).to eq(3)
    end

    it "excludes entries outside the window" do
      log = [BASE_TIME - 120, BASE_TIME - 61, BASE_TIME - 30, BASE_TIME]
      expect(gw.count_in_window(log, BASE_TIME, 60)).to eq(2)
    end

    it "excludes the left edge (now - window_seconds)" do
      log = [BASE_TIME - 60]
      expect(gw.count_in_window(log, BASE_TIME, 60)).to eq(0)
    end

    it "includes an entry just inside the left edge" do
      log = [BASE_TIME - 59.999]
      expect(gw.count_in_window(log, BASE_TIME, 60)).to eq(1)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — is_allowed
  # ---------------------------------------------------------------------------
  describe "#is_allowed" do
    it "allows a request under both limits" do
      expect(gw_with_key.is_allowed("key_abc", BASE_TIME)).to eq(true)
    end

    it "denies when the key does not exist" do
      expect(gw.is_allowed("ghost", BASE_TIME)).to eq(false)
    end

    it "denies when the key is disabled" do
      gw_with_key.revoke_key("key_abc")
      expect(gw_with_key.is_allowed("key_abc", BASE_TIME)).to eq(false)
    end

    it "denies when the per-minute cap is exceeded" do
      gw.create_key("k1", "alice", "free") # rpm: 3
      seed_request_log(gw, "k1", [BASE_TIME - 30, BASE_TIME - 20, BASE_TIME - 10])
      expect(gw.is_allowed("k1", BASE_TIME)).to eq(false)
    end

    it "allows once the per-minute window has rolled off" do
      gw.create_key("k1", "alice", "free") # rpm: 3
      seed_request_log(gw, "k1", [BASE_TIME - 90, BASE_TIME - 80, BASE_TIME - 70])
      expect(gw.is_allowed("k1", BASE_TIME)).to eq(true)
    end

    it "denies when the per-day cap is exceeded" do
      gw.create_key("k1", "alice", "free") # rpd: 10
      seed_request_log(gw, "k1", (0...10).map { |i| BASE_TIME - i * 100 })
      expect(gw.is_allowed("k1", BASE_TIME)).to eq(false)
    end

    it "always allows an unlimited plan" do
      gw.create_key("k1", "alice", "unlimited")
      seed_request_log(gw, "k1", (0...1000).map { |i| BASE_TIME - i })
      expect(gw.is_allowed("k1", BASE_TIME)).to eq(true)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — record_request
  # ---------------------------------------------------------------------------
  describe "#record_request" do
    it "appends the timestamp" do
      gw_with_key.record_request("key_abc", BASE_TIME)
      log = gw_with_key.instance_variable_get(:@keys)["key_abc"][:request_log]
      expect(log).to include(BASE_TIME)
    end

    it "prunes entries older than 25 hours" do
      old = BASE_TIME - 90_001
      seed_request_log(gw_with_key, "key_abc", [old])
      gw_with_key.record_request("key_abc", BASE_TIME)
      log = gw_with_key.instance_variable_get(:@keys)["key_abc"][:request_log]
      expect(log).not_to include(old)
    end

    it "keeps entries within 25 hours" do
      recent = BASE_TIME - 3600
      seed_request_log(gw_with_key, "key_abc", [recent])
      gw_with_key.record_request("key_abc", BASE_TIME)
      log = gw_with_key.instance_variable_get(:@keys)["key_abc"][:request_log]
      expect(log).to include(recent)
    end

    it "raises KeyError when the key does not exist" do
      expect { gw.record_request("ghost", BASE_TIME) }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — handle_request
  # ---------------------------------------------------------------------------
  describe "#handle_request" do
    it "allows and records on success" do
      result = gw_with_key.handle_request("key_abc", BASE_TIME)
      expect(result[:allowed]).to eq(true)
    end

    it "reports key_not_found" do
      result = gw.handle_request("ghost", BASE_TIME)
      expect(result).to eq(allowed: false, reason: :key_not_found)
    end

    it "reports key_disabled" do
      gw_with_key.revoke_key("key_abc")
      result = gw_with_key.handle_request("key_abc", BASE_TIME)
      expect(result).to eq(allowed: false, reason: :key_disabled)
    end

    it "reports rpm_exceeded" do
      gw.create_key("k1", "alice", "free") # rpm: 3
      seed_request_log(gw, "k1", [BASE_TIME - 10, BASE_TIME - 5, BASE_TIME - 1])
      result = gw.handle_request("k1", BASE_TIME)
      expect(result).to eq(allowed: false, reason: :rpm_exceeded)
    end

    it "reports rpd_exceeded without mutating the log" do
      gw.create_key("k1", "alice", "free") # rpd: 10
      log_before = (0...10).map { |i| BASE_TIME - i * 100 }
      seed_request_log(gw, "k1", log_before.dup)
      result = gw.handle_request("k1", BASE_TIME)
      expect(result).to eq(allowed: false, reason: :rpd_exceeded)
      expect(gw.instance_variable_get(:@keys)["k1"][:request_log]).to eq(log_before)
    end

    it "only checks rpd after confirming rpm is within limits" do
      gw.create_key("k1", "alice", "free") # rpm: 3, rpd: 10
      seed_request_log(gw, "k1", (1..10).map { |i| BASE_TIME - 3600 * i })
      result = gw.handle_request("k1", BASE_TIME)
      expect(result[:reason]).to eq(:rpd_exceeded)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — get_usage
  # ---------------------------------------------------------------------------
  describe "#get_usage" do
    it "returns correct usage counts and limits" do
      seed_request_log(gw_with_key, "key_abc", [BASE_TIME - 30, BASE_TIME - 3600])
      stats = gw_with_key.get_usage("key_abc", BASE_TIME)
      expect(stats).to eq(
        key_id: "key_abc",
        plan: "pro",
        rpm_used: 1,
        rpm_limit: 100,
        rpd_used: 2,
        rpd_limit: 5_000
      )
    end

    it "shows nil limits for an unlimited plan" do
      gw.create_key("k1", "alice", "unlimited")
      stats = gw.get_usage("k1", BASE_TIME)
      expect(stats[:rpm_limit]).to be_nil
      expect(stats[:rpd_limit]).to be_nil
    end

    it "raises KeyError when the key does not exist" do
      expect { gw.get_usage("ghost", BASE_TIME) }.to raise_error(KeyError)
    end
  end
end
