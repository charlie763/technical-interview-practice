load_problem("06_lab_cadence_monitor")

RSpec.describe "Lab Cadence Monitor" do
  let(:m) { make_monitor }

  let(:seeded) do
    monitor = make_monitor
    register_patient(monitor, "alice", ["hba1c", "bmp", "lipids"])
    register_patient(monitor, "bob", ["hba1c", "bmp"])
    register_patient(monitor, "carol", ["hba1c"])

    # Alice: hba1c quarterly deadlines; submitted first one, missed second
    set_lab_deadline(monitor, "alice", "hba1c", Date.new(2024, 3, 31))
    set_lab_deadline(monitor, "alice", "hba1c", Date.new(2024, 6, 30))
    record_submission(monitor, "alice", "hba1c", Date.new(2024, 3, 28))

    # Alice: bmp due but no submission
    set_lab_deadline(monitor, "alice", "bmp", Date.new(2024, 4, 15))

    # Bob: hba1c submitted on time; bmp overdue
    set_lab_deadline(monitor, "bob", "hba1c", Date.new(2024, 3, 31))
    record_submission(monitor, "bob", "hba1c", Date.new(2024, 3, 25))
    set_lab_deadline(monitor, "bob", "bmp", Date.new(2024, 3, 31))

    # Carol: all labs submitted on time
    set_lab_deadline(monitor, "carol", "hba1c", Date.new(2024, 3, 31))
    record_submission(monitor, "carol", "hba1c", Date.new(2024, 3, 15))

    monitor
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Patient & lab registration
  # ---------------------------------------------------------------------------
  describe "#register_patient" do
    it "registers a new patient" do
      register_patient(m, "dave", ["hba1c"])
      expect(get_required_labs(m, "dave")).to eq(Set["hba1c"])
    end

    it "raises ArgumentError when required_labs is empty" do
      expect { register_patient(m, "eve", []) }.to raise_error(ArgumentError)
    end

    it "is idempotent for a lab that's already required" do
      register_patient(m, "frank", ["hba1c"])
      register_patient(m, "frank", ["hba1c"])
      expect(get_required_labs(m, "frank")).to eq(Set["hba1c"])
    end

    it "adds new labs to an existing patient" do
      register_patient(m, "grace", ["hba1c"])
      register_patient(m, "grace", ["bmp"])
      expect(get_required_labs(m, "grace")).to eq(Set["hba1c", "bmp"])
    end

    it "does not remove existing labs" do
      register_patient(m, "hank", ["hba1c", "bmp"])
      register_patient(m, "hank", ["lipids"])
      labs = get_required_labs(m, "hank")
      expect(labs).to include("hba1c", "bmp")
    end
  end

  describe "#add_required_lab" do
    it "adds a lab to an existing patient" do
      register_patient(m, "iris", ["hba1c"])
      add_required_lab(m, "iris", "bmp")
      expect(get_required_labs(m, "iris")).to include("bmp")
    end

    it "is idempotent" do
      register_patient(m, "jack", ["hba1c"])
      add_required_lab(m, "jack", "hba1c")
      expect(get_required_labs(m, "jack")).to eq(Set["hba1c"])
    end

    it "raises KeyError for an unknown patient" do
      expect { add_required_lab(m, "nobody", "hba1c") }.to raise_error(KeyError)
    end
  end

  describe "#get_required_labs" do
    it "returns the correct set" do
      register_patient(m, "kate", ["hba1c", "lipids"])
      expect(get_required_labs(m, "kate")).to eq(Set["hba1c", "lipids"])
    end

    it "raises KeyError for an unknown patient" do
      expect { get_required_labs(m, "nobody") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Deadlines and submissions
  # ---------------------------------------------------------------------------
  describe "#set_lab_deadline" do
    it "sets a deadline without raising" do
      expect { set_lab_deadline(seeded, "alice", "lipids", Date.new(2024, 5, 1)) }.not_to raise_error
    end

    it "ignores a duplicate deadline" do
      set_lab_deadline(seeded, "alice", "lipids", Date.new(2024, 5, 1))
      set_lab_deadline(seeded, "alice", "lipids", Date.new(2024, 5, 1)) # duplicate
      expect(is_overdue(seeded, "alice", "lipids", Date.new(2024, 5, 2))).to be true
    end

    it "raises KeyError for an unknown patient" do
      expect { set_lab_deadline(m, "nobody", "hba1c", Date.new(2024, 3, 31)) }.to raise_error(KeyError)
    end

    it "raises ArgumentError for a lab that isn't required" do
      register_patient(m, "leo", ["hba1c"])
      expect { set_lab_deadline(m, "leo", "bmp", Date.new(2024, 3, 31)) }.to raise_error(ArgumentError)
    end
  end

  describe "#record_submission" do
    it "clears the matching deadline" do
      # carol submitted hba1c on Mar 15, deadline was Mar 31 -> not overdue
      expect(is_overdue(seeded, "carol", "hba1c", Date.new(2024, 4, 1))).to be false
    end

    it "clears only the earliest applicable deadline" do
      # alice: hba1c deadlines Mar 31 (cleared) and Jun 30 (not cleared)
      # as of Jul 1, the Jun 30 deadline should be overdue
      expect(is_overdue(seeded, "alice", "hba1c", Date.new(2024, 7, 1))).to be true
    end

    it "raises KeyError for an unknown patient" do
      expect { record_submission(m, "nobody", "hba1c", Date.new(2024, 3, 28)) }.to raise_error(KeyError)
    end

    it "raises ArgumentError for a lab that isn't required" do
      register_patient(m, "mia", ["hba1c"])
      expect { record_submission(m, "mia", "bmp", Date.new(2024, 3, 28)) }.to raise_error(ArgumentError)
    end

    it "does not let a late submission clear a past deadline" do
      register_patient(m, "noah", ["hba1c"])
      set_lab_deadline(m, "noah", "hba1c", Date.new(2024, 3, 31))
      record_submission(m, "noah", "hba1c", Date.new(2024, 4, 15)) # submitted late
      expect(is_overdue(m, "noah", "hba1c", Date.new(2024, 4, 1))).to be true
    end
  end

  describe "#is_overdue" do
    it "is true for a deadline with no submission" do
      # bob: bmp deadline Mar 31, no submission
      expect(is_overdue(seeded, "bob", "bmp", Date.new(2024, 4, 1))).to be true
    end

    it "is false once submitted on time" do
      expect(is_overdue(seeded, "bob", "hba1c", Date.new(2024, 4, 1))).to be false
    end

    it "is false before the deadline has passed" do
      # alice: bmp deadline Apr 15; as of Apr 14 not yet overdue
      expect(is_overdue(seeded, "alice", "bmp", Date.new(2024, 4, 14))).to be false
    end

    it "is true the day after the deadline" do
      expect(is_overdue(seeded, "alice", "bmp", Date.new(2024, 4, 16))).to be true
    end

    it "is false when no deadline has been set" do
      # alice: lipids has no deadline set
      expect(is_overdue(seeded, "alice", "lipids", Date.new(2024, 6, 1))).to be false
    end

    it "is false for an unknown patient" do
      expect(is_overdue(seeded, "nobody", "hba1c", Date.new(2024, 4, 1))).to be false
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Compliance reporting
  # ---------------------------------------------------------------------------
  describe "#overdue_labs" do
    it "returns a sorted array of overdue lab names" do
      # alice as of Jul 1: bmp (missed Apr 15) and hba1c (missed Jun 30)
      labs = overdue_labs(seeded, "alice", Date.new(2024, 7, 1))
      expect(labs).to eq(labs.sort)
      expect(labs).to include("bmp", "hba1c")
    end

    it "returns empty when nothing is overdue" do
      expect(overdue_labs(seeded, "carol", Date.new(2024, 4, 1))).to eq([])
    end

    it "returns empty for an unknown patient" do
      expect(overdue_labs(seeded, "nobody", Date.new(2024, 4, 1))).to eq([])
    end
  end

  describe "#compliance_report" do
    it "includes patients with overdue labs" do
      report = compliance_report(seeded, as_of: Date.new(2024, 4, 16))
      patient_ids = report.map { |e| e[:patient_id] }
      expect(patient_ids).to include("alice", "bob")
    end

    it "excludes fully compliant patients" do
      report = compliance_report(seeded, as_of: Date.new(2024, 4, 16))
      patient_ids = report.map { |e| e[:patient_id] }
      expect(patient_ids).not_to include("carol")
    end

    it "returns entries with the expected fields" do
      report = compliance_report(seeded, as_of: Date.new(2024, 4, 16))
      bob_entry = report.find { |e| e[:patient_id] == "bob" }
      expect(bob_entry.keys).to contain_exactly(:patient_id, :overdue_labs, :overdue_count)
      expect(bob_entry[:overdue_count]).to eq(bob_entry[:overdue_labs].size)
    end

    it "sorts by overdue_count descending" do
      report = compliance_report(seeded, as_of: Date.new(2024, 7, 1))
      counts = report.map { |e| e[:overdue_count] }
      expect(counts).to eq(counts.sort.reverse)
    end

    it "returns an empty array when nobody is overdue" do
      register_patient(m, "perfectly_compliant", ["hba1c"])
      set_lab_deadline(m, "perfectly_compliant", "hba1c", Date.new(2024, 3, 31))
      record_submission(m, "perfectly_compliant", "hba1c", Date.new(2024, 3, 20))
      expect(compliance_report(m, as_of: Date.new(2024, 4, 1))).to eq([])
    end
  end

  # ---------------------------------------------------------------------------
  # PART 4 — Submission history
  # ---------------------------------------------------------------------------
  describe "#submission_history" do
    it "returns dates sorted chronologically" do
      register_patient(m, "quinn", ["hba1c"])
      set_lab_deadline(m, "quinn", "hba1c", Date.new(2024, 3, 31))
      set_lab_deadline(m, "quinn", "hba1c", Date.new(2024, 6, 30))
      record_submission(m, "quinn", "hba1c", Date.new(2024, 3, 20))
      record_submission(m, "quinn", "hba1c", Date.new(2024, 6, 15))
      history = submission_history(m, "quinn", "hba1c")
      expect(history).to eq(history.sort)
      expect(history.size).to eq(2)
    end

    it "returns empty when there are no submissions" do
      expect(submission_history(seeded, "alice", "lipids")).to eq([])
    end

    it "returns empty for an unknown patient" do
      expect(submission_history(m, "nobody", "hba1c")).to eq([])
    end
  end

  describe "#days_since_last_submission" do
    it "returns the correct day count" do
      register_patient(m, "rita", ["hba1c"])
      set_lab_deadline(m, "rita", "hba1c", Date.new(2024, 3, 31))
      record_submission(m, "rita", "hba1c", Date.new(2024, 3, 20))
      expect(days_since_last_submission(m, "rita", "hba1c", Date.new(2024, 4, 20))).to eq(31)
    end

    it "returns nil when there's no submission" do
      expect(days_since_last_submission(seeded, "alice", "lipids", Date.new(2024, 5, 1))).to be_nil
    end

    it "returns nil for an unknown patient" do
      expect(days_since_last_submission(m, "nobody", "hba1c", Date.new(2024, 4, 1))).to be_nil
    end
  end
end
