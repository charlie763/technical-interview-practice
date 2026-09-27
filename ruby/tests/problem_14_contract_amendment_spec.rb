load_problem("14_contract_amendment")

# Shared dates (namespaced to this problem to avoid clashing with other
# spec files' own date constants).
AMD_D_BASE = "2025-01-01".freeze # base / before any amendments
AMD_D1 = "2025-03-01".freeze # amendment 1 effective date
AMD_D2 = "2025-06-01".freeze # amendment 2 effective date
AMD_D3 = "2025-09-01".freeze # amendment 3 effective date
AMD_D_AFTER = "2025-12-31".freeze # well after all amendments

RSpec.describe ContractAmendmentManager do
  let(:mgr) { described_class.new }

  # Pre-seeded manager:
  #   c-seed-1  "Vendor MSA"
  #             fields = { value: 50000, payment_terms: "net-30", currency: "USD" }
  #             amd-s1 (2025-03-01): overrides = { payment_terms: "net-45" }
  #             amd-s2 (2025-06-01): overrides = { value: 75000 }
  #
  #   c-seed-2  "NDA Agreement"
  #             fields = { term_years: 2, auto_renew: true }
  #             (no amendments)
  let(:seeded_mgr) do
    m = described_class.new
    m.add_contract("c-seed-1", "Vendor MSA", { "value" => 50_000, "payment_terms" => "net-30", "currency" => "USD" })
    m.add_contract("c-seed-2", "NDA Agreement", { "term_years" => 2, "auto_renew" => true })
    m.add_amendment("amd-s1", "c-seed-1", AMD_D1, { "payment_terms" => "net-45" }, "extended payment terms")
    m.add_amendment("amd-s2", "c-seed-1", AMD_D2, { "value" => 75_000 }, "scope increase")
    m
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Base contract management
  # ---------------------------------------------------------------------------
  describe "#add_contract" do
    it "returns the contract hash" do
      c = mgr.add_contract("c-add-1", "Test", { "x" => 1 })
      expect(c[:contract_id]).to eq("c-add-1")
      expect(c[:title]).to eq("Test")
      expect(c[:fields]["x"]).to eq(1)
    end

    it "stores a copy of fields" do
      original = { "x" => 1 }
      mgr.add_contract("c-copy-1", "Test", original)
      original["x"] = 999
      expect(mgr.get_base_contract("c-copy-1")[:fields]["x"]).to eq(1)
    end

    it "raises ArgumentError on a duplicate" do
      mgr.add_contract("c-dup-am", "A", {})
      expect { mgr.add_contract("c-dup-am", "B", {}) }.to raise_error(ArgumentError)
    end
  end

  describe "#get_base_contract" do
    it "returns the original fields" do
      base = seeded_mgr.get_base_contract("c-seed-1")
      # Even though amendments exist, base fields are unchanged
      expect(base[:fields]["payment_terms"]).to eq("net-30")
      expect(base[:fields]["value"]).to eq(50_000)
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.get_base_contract("no-such") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Amendments and effective contract
  # ---------------------------------------------------------------------------
  describe "#add_amendment" do
    it "returns the amendment hash" do
      amd = seeded_mgr.add_amendment("amd-add-1", "c-seed-1", AMD_D3, { "currency" => "EUR" }, "switch currency")
      expect(amd[:amendment_id]).to eq("amd-add-1")
      expect(amd[:contract_id]).to eq("c-seed-1")
      expect(amd[:effective_on]).to eq(AMD_D3)
      expect(amd[:overrides]).to eq({ "currency" => "EUR" })
      expect(amd[:note]).to eq("switch currency")
    end

    it "stores a copy of overrides" do
      mgr.add_contract("c-amd-copy", "T", { "x" => 1 })
      overrides = { "x" => 2 }
      mgr.add_amendment("amd-copy-1", "c-amd-copy", AMD_D1, overrides, "")
      overrides["x"] = 999
      expect(mgr.get_amendments("c-amd-copy").first[:overrides]["x"]).to eq(2)
    end

    it "raises ArgumentError on a duplicate amendment_id" do
      expect { seeded_mgr.add_amendment("amd-s1", "c-seed-1", AMD_D3, {}, "") }.to raise_error(ArgumentError)
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.add_amendment("amd-new", "no-such", AMD_D1, {}, "") }.to raise_error(KeyError)
    end
  end

  describe "#get_amendments" do
    it "is sorted by effective_on" do
      dates = seeded_mgr.get_amendments("c-seed-1").map { |a| a[:effective_on] }
      expect(dates).to eq(dates.sort)
    end

    it "is empty when there are none" do
      expect(seeded_mgr.get_amendments("c-seed-2")).to eq([])
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.get_amendments("no-such") }.to raise_error(KeyError)
    end

    it "breaks same-date ties by amendment_id" do
      mgr.add_contract("c-same-dt", "T", { "x" => 0 })
      mgr.add_amendment("amd-z", "c-same-dt", AMD_D1, { "x" => 2 }, "")
      mgr.add_amendment("amd-a", "c-same-dt", AMD_D1, { "x" => 1 }, "")
      ids = mgr.get_amendments("c-same-dt").map { |a| a[:amendment_id] }
      expect(ids).to eq(ids.sort)
    end
  end

  describe "#get_effective_contract" do
    it "returns base fields before any amendments" do
      fields = seeded_mgr.get_effective_contract("c-seed-1", AMD_D_BASE)
      expect(fields["payment_terms"]).to eq("net-30")
      expect(fields["value"]).to eq(50_000)
    end

    it "applies only the first amendment when in its window" do
      fields = seeded_mgr.get_effective_contract("c-seed-1", "2025-04-15")
      expect(fields["payment_terms"]).to eq("net-45")
      expect(fields["value"]).to eq(50_000) # amd-s2 not yet effective
    end

    it "applies all amendments once past their effective dates" do
      fields = seeded_mgr.get_effective_contract("c-seed-1", AMD_D_AFTER)
      expect(fields["payment_terms"]).to eq("net-45")
      expect(fields["value"]).to eq(75_000)
    end

    it "preserves fields no amendment touched" do
      fields = seeded_mgr.get_effective_contract("c-seed-1", AMD_D_AFTER)
      expect(fields["currency"]).to eq("USD")
    end

    it "returns the base fields when there are no amendments" do
      fields = seeded_mgr.get_effective_contract("c-seed-2", AMD_D_AFTER)
      expect(fields["term_years"]).to eq(2)
      expect(fields["auto_renew"]).to be true
    end

    it "includes an amendment exactly on its effective date" do
      # amd-s1 effective_on = AMD_D1; querying exactly AMD_D1 should apply it
      fields = seeded_mgr.get_effective_contract("c-seed-1", AMD_D1)
      expect(fields["payment_terms"]).to eq("net-45")
    end

    it "does not mutate the base contract" do
      seeded_mgr.get_effective_contract("c-seed-1", AMD_D_AFTER)
      base = seeded_mgr.get_base_contract("c-seed-1")
      expect(base[:fields]["payment_terms"]).to eq("net-30")
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.get_effective_contract("no-such", AMD_D_AFTER) }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Value history and amendment summary
  # ---------------------------------------------------------------------------
  describe "#get_value_history" do
    it "includes the base value first" do
      history = seeded_mgr.get_value_history("c-seed-1", "value")
      expect(history.first[:source]).to eq("base")
      expect(history.first[:value]).to eq(50_000)
    end

    it "includes an amendment override" do
      sources = seeded_mgr.get_value_history("c-seed-1", "payment_terms").map { |e| e[:source] }
      expect(sources).to include("amd-s1")
    end

    it "excludes amendments that did not touch the field" do
      # amd-s2 changes "value", not "payment_terms"
      sources = seeded_mgr.get_value_history("c-seed-1", "payment_terms").map { |e| e[:source] }
      expect(sources).not_to include("amd-s2")
    end

    it "is sorted chronologically" do
      history = seeded_mgr.get_value_history("c-seed-1", "value")
      # base first, then amendment entries in date order
      expect(history.first[:source]).to eq("base")
      dates = history.reject { |e| e[:source] == "base" }.map { |e| e[:effective_on] }
      expect(dates).to eq(dates.sort)
    end

    it "raises KeyError when the field is never present" do
      expect { seeded_mgr.get_value_history("c-seed-1", "nonexistent_field") }.to raise_error(KeyError)
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.get_value_history("no-such", "value") }.to raise_error(KeyError)
    end
  end

  describe "#get_amendment_summary" do
    it "counts amendments" do
      expect(seeded_mgr.get_amendment_summary("c-seed-1")[:amendment_count]).to eq(2)
    end

    it "returns fields_amended sorted" do
      summary = seeded_mgr.get_amendment_summary("c-seed-1")
      expect(summary[:fields_amended]).to eq(%w[payment_terms value].sort)
    end

    it "reports the latest amendment date" do
      expect(seeded_mgr.get_amendment_summary("c-seed-1")[:latest_amendment]).to eq(AMD_D2)
    end

    it "reflects all amendments in current_fields" do
      summary = seeded_mgr.get_amendment_summary("c-seed-1")
      expect(summary[:current_fields]["payment_terms"]).to eq("net-45")
      expect(summary[:current_fields]["value"]).to eq(75_000)
    end

    it "reports nil latest_amendment when there are none" do
      summary = seeded_mgr.get_amendment_summary("c-seed-2")
      expect(summary[:latest_amendment]).to be_nil
      expect(summary[:amendment_count]).to eq(0)
      expect(summary[:fields_amended]).to eq([])
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_mgr.get_amendment_summary("no-such") }.to raise_error(KeyError)
    end
  end
end
