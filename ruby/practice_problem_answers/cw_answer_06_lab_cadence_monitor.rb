require "date"
require "set"

def make_monitor
  { patients: {} }
end

def register_patient(monitor, patient_id, required_labs)
  raise ArgumentError, "required_labs must not be empty" if required_labs.empty?

  patient = monitor[:patients][patient_id]
  if patient
    patient[:required_labs].merge(required_labs)
  else
    monitor[:patients][patient_id] = {
      required_labs: Set.new(required_labs),
      deadlines: {},
      submissions: {},
    }
  end
end

def add_required_lab(monitor, patient_id, lab_type)
  fetch_patient(monitor, patient_id)[:required_labs].add(lab_type)
end

def get_required_labs(monitor, patient_id)
  fetch_patient(monitor, patient_id)[:required_labs]
end

def set_lab_deadline(monitor, patient_id, lab_type, due_date)
  patient = fetch_patient_with_required_lab(monitor, patient_id, lab_type)
  deadlines = (patient[:deadlines][lab_type] ||= [])
  deadlines << due_date unless deadlines.include?(due_date)
end

def record_submission(monitor, patient_id, lab_type, submitted_on)
  patient = fetch_patient_with_required_lab(monitor, patient_id, lab_type)
  (patient[:submissions][lab_type] ||= []) << submitted_on

  deadlines = patient[:deadlines][lab_type]
  return unless deadlines

  sorted = deadlines.sort
  index = sorted.find_index { |d| d >= submitted_on }
  sorted.delete_at(index) if index
  patient[:deadlines][lab_type] = sorted
end

def is_overdue(monitor, patient_id, lab_type, as_of)
  patient = monitor[:patients][patient_id]
  return false unless patient
  return false unless patient[:required_labs].include?(lab_type)

  deadlines = patient[:deadlines][lab_type]
  return false unless deadlines

  deadlines.any? { |due_date| due_date < as_of }
end

def overdue_labs(monitor, patient_id, as_of)
  patient = monitor[:patients][patient_id]
  return [] unless patient

  patient[:required_labs]
    .select { |lab_type| is_overdue(monitor, patient_id, lab_type, as_of) }
    .sort
end

def compliance_report(monitor, as_of:)
  report = monitor[:patients].keys.filter_map do |patient_id|
    labs = overdue_labs(monitor, patient_id, as_of)
    next if labs.empty?

    { patient_id: patient_id, overdue_labs: labs, overdue_count: labs.size }
  end

  report.sort_by { |entry| [-entry[:overdue_count], entry[:patient_id]] }
end

def submission_history(monitor, patient_id, lab_type)
  patient = monitor[:patients][patient_id]
  return [] unless patient

  (patient[:submissions][lab_type] || []).sort
end

def days_since_last_submission(monitor, patient_id, lab_type, as_of)
  history = submission_history(monitor, patient_id, lab_type)
  return nil if history.empty?

  (as_of - history.last).to_i
end

private

def fetch_patient(monitor, patient_id)
  monitor[:patients].fetch(patient_id) { raise KeyError, "patient not found: #{patient_id}" }
end

def fetch_patient_with_required_lab(monitor, patient_id, lab_type)
  patient = fetch_patient(monitor, patient_id)
  raise ArgumentError, "lab not required: #{lab_type}" unless patient[:required_labs].include?(lab_type)

  patient
end
