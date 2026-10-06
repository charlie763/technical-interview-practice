load_problem("12_contract_alert_scheduler")

# Shared dates
# D0 = base date, D30 = D0+30d, D60 = D0+60d, D90 = D0+90d
D0 = "2025-01-01".freeze
D30 = "2025-01-31".freeze
D60 = "2025-03-02".freeze # Jan has 31 days, Feb 2025 has 28 days -> Jan 1 + 60 = Mar 2
D90 = "2025-04-01".freeze

RSpec.describe ContractAlertScheduler do
  let(:sched) { described_class.new }

  # Pre-seeded scheduler:
  #   c-seed-1  "Vendor MSA"        owner=legal@acme.com  expires=2025-06-30
  #   c-seed-2  "SaaS Subscription" owner=ops@acme.com    expires=2025-09-15
  #   c-seed-3  "NDA Agreement"     owner=legal@acme.com  expires=2025-12-31
  #
  # Alert configs:
  #   cfg-30   days_before=30  label="30-day notice"
  #   cfg-7    days_before=7   label="final warning"
  let(:seeded_sched) do
    s = described_class.new
    s.add_contract("c-seed-1", "Vendor MSA", "legal@acme.com", "2025-06-30")
    s.add_contract("c-seed-2", "SaaS Subscription", "ops@acme.com", "2025-09-15")
    s.add_contract("c-seed-3", "NDA Agreement", "legal@acme.com", "2025-12-31")
    s.add_alert_config("cfg-30", 30, "30-day notice")
    s.add_alert_config("cfg-7", 7, "final warning")
    s
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Contract and alert-config management
  # ---------------------------------------------------------------------------
  describe "#add_contract" do
    it "returns the stored contract" do
      c = sched.add_contract("c-add-1", "Test Contract", "a@b.com", "2025-06-01")
      expect(c[:contract_id]).to eq("c-add-1")
      expect(c[:title]).to eq("Test Contract")
      expect(c[:owner_email]).to eq("a@b.com")
      expect(c[:expires_on]).to eq("2025-06-01")
    end

    it "raises ArgumentError on a duplicate" do
      sched.add_contract("c-dup-1", "A", "a@b.com", "2025-06-01")
      expect { sched.add_contract("c-dup-1", "B", "b@c.com", "2025-07-01") }.to raise_error(ArgumentError)
    end

    it "stores multiple contracts" do
      results = seeded_sched.get_contracts_expiring_between("2025-01-01", "2025-12-31")
      ids = results.map { |r| r[:contract_id] }
      expect(ids).to include("c-seed-1", "c-seed-2", "c-seed-3")
    end
  end

  describe "#add_alert_config" do
    it "returns the stored config" do
      cfg = sched.add_alert_config("cfg-add-1", 14, "two-week notice")
      expect(cfg[:config_id]).to eq("cfg-add-1")
      expect(cfg[:days_before]).to eq(14)
      expect(cfg[:label]).to eq("two-week notice")
    end

    it "raises ArgumentError on a duplicate" do
      sched.add_alert_config("cfg-dup-1", 30, "notice")
      expect { sched.add_alert_config("cfg-dup-1", 60, "other") }.to raise_error(ArgumentError)
    end
  end

  describe "#get_contracts_expiring_between" do
    it "matches an exact single-day range" do
      results = seeded_sched.get_contracts_expiring_between("2025-06-30", "2025-06-30")
      expect(results.size).to eq(1)
      expect(results.first[:contract_id]).to eq("c-seed-1")
    end

    it "spans a range covering multiple contracts" do
      results = seeded_sched.get_contracts_expiring_between("2025-06-01", "2025-09-30")
      ids = results.map { |r| r[:contract_id] }
      expect(ids).to include("c-seed-1", "c-seed-2")
      expect(ids).not_to include("c-seed-3")
    end

    it "is sorted ascending" do
      dates = seeded_sched.get_contracts_expiring_between("2025-01-01", "2025-12-31").map { |r| r[:expires_on] }
      expect(dates).to eq(dates.sort)
    end

    it "is empty when nothing is in range" do
      expect(seeded_sched.get_contracts_expiring_between("2024-01-01", "2024-12-31")).to eq([])
    end

    it "includes the start boundary" do
      # c-seed-1 expires exactly on 2025-06-30; start = 2025-06-30 should include it
      ids = seeded_sched.get_contracts_expiring_between("2025-06-30", "2025-12-31").map { |r| r[:contract_id] }
      expect(ids).to include("c-seed-1")
    end

    it "includes the end boundary" do
      ids = seeded_sched.get_contracts_expiring_between("2025-01-01", "2025-06-30").map { |r| r[:contract_id] }
      expect(ids).to include("c-seed-1")
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Alert schedule computation
  # ---------------------------------------------------------------------------
  describe "#compute_alert_schedule" do
    it "returns one entry per config" do
      # 2 configs registered -> 2 schedule entries
      expect(seeded_sched.compute_alert_schedule("c-seed-1").size).to eq(2)
    end

    it "computes the correct alert_on dates" do
      # c-seed-1 expires 2025-06-30
      # cfg-30: 2025-06-30 - 30d = 2025-05-31
      # cfg-7:  2025-06-30 - 7d  = 2025-06-23
      by_cfg = seeded_sched.compute_alert_schedule("c-seed-1").to_h { |e| [e[:config_id], e] }
      expect(by_cfg["cfg-30"][:alert_on]).to eq("2025-05-31")
      expect(by_cfg["cfg-7"][:alert_on]).to eq("2025-06-23")
    end

    it "is sorted by alert_on ascending" do
      dates = seeded_sched.compute_alert_schedule("c-seed-1").map { |e| e[:alert_on] }
      expect(dates).to eq(dates.sort)
    end

    it "includes the config's label" do
      labels = seeded_sched.compute_alert_schedule("c-seed-1").map { |e| e[:label] }
      expect(labels).to include("30-day notice", "final warning")
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_sched.compute_alert_schedule("no-such-contract") }.to raise_error(KeyError)
    end

    it "returns an empty array when no configs are registered" do
      sched.add_contract("c-no-cfg", "Bare Contract", "a@b.com", "2025-06-01")
      expect(sched.compute_alert_schedule("c-no-cfg")).to eq([])
    end
  end

  describe "#get_due_alerts" do
    it "returns alerts on or before the given date" do
      # cfg-30 for c-seed-1 fires on 2025-05-31
      due = seeded_sched.get_due_alerts("2025-05-31")
      entries = due.map { |e| [e[:contract_id], e[:config_id]] }
      expect(entries).to include(["c-seed-1", "cfg-30"])
    end

    it "excludes future alerts" do
      expect(seeded_sched.get_due_alerts("2025-01-01")).to eq([])
    end

    it "includes owner_email and expires_on" do
      due = seeded_sched.get_due_alerts("2025-05-31")
      entry = due.find { |e| e[:contract_id] == "c-seed-1" && e[:config_id] == "cfg-30" }
      expect(entry[:owner_email]).to eq("legal@acme.com")
      expect(entry[:expires_on]).to eq("2025-06-30")
    end

    it "is sorted by alert_on then contract_id" do
      # Ask for a date far enough in the future to capture many alerts
      dates = seeded_sched.get_due_alerts("2025-12-31").map { |e| e[:alert_on] }
      expect(dates).to eq(dates.sort)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Sent records and upcoming alerts
  # ---------------------------------------------------------------------------
  describe "#record_alert_sent" do
    it "returns the sent record" do
      rec = seeded_sched.record_alert_sent("c-seed-1", "cfg-30", "2025-05-31")
      expect(rec[:contract_id]).to eq("c-seed-1")
      expect(rec[:config_id]).to eq("cfg-30")
      expect(rec[:sent_on]).to eq("2025-05-31")
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_sched.record_alert_sent("no-contract", "cfg-30", "2025-05-31") }.to raise_error(KeyError)
    end

    it "raises KeyError for an unknown config" do
      expect { seeded_sched.record_alert_sent("c-seed-1", "no-cfg", "2025-05-31") }.to raise_error(KeyError)
    end

    it "stores multiple sends" do
      seeded_sched.record_alert_sent("c-seed-1", "cfg-30", "2025-05-31")
      seeded_sched.record_alert_sent("c-seed-1", "cfg-7", "2025-06-23")
      upcoming = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-05-01")
      sent_ids = upcoming.select { |e| e[:sent] }.map { |e| e[:config_id] }
      expect(sent_ids).to include("cfg-30", "cfg-7")
    end
  end

  describe "#get_upcoming_alerts" do
    it "excludes past alerts" do
      # as_of_date = 2025-06-01; cfg-30 alert_on=2025-05-31 is in the past
      config_ids = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-06-01").map { |e| e[:config_id] }
      expect(config_ids).not_to include("cfg-30")
    end

    it "includes future alerts" do
      # cfg-7 alert_on=2025-06-23 is still upcoming from 2025-06-01
      config_ids = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-06-01").map { |e| e[:config_id] }
      expect(config_ids).to include("cfg-7")
    end

    it "defaults sent to false" do
      upcoming = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-05-01")
      expect(upcoming).to all(include(sent: false))
    end

    it "reports sent as true after recording" do
      seeded_sched.record_alert_sent("c-seed-1", "cfg-30", "2025-05-31")
      upcoming = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-05-01")
      entry = upcoming.find { |e| e[:config_id] == "cfg-30" }
      expect(entry[:sent]).to be true
    end

    it "is sorted by alert_on ascending" do
      dates = seeded_sched.get_upcoming_alerts("c-seed-1", "2025-01-01").map { |e| e[:alert_on] }
      expect(dates).to eq(dates.sort)
    end

    it "raises KeyError for an unknown contract" do
      expect { seeded_sched.get_upcoming_alerts("no-such", "2025-01-01") }.to raise_error(KeyError)
    end
  end
end
