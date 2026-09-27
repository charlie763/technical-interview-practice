require "set"

ALLOWED_TRANSITIONS = {
  "referred" => Set["screened"],
  "screened" => Set["enrolled", "ineligible"],
  "enrolled" => Set["active", "withdrawn"],
  "active" => Set["graduated", "churned", "withdrawn"],
}.freeze

TERMINAL_STATES = Set["ineligible", "withdrawn", "graduated", "churned"].freeze

class EnrollmentPipeline
  def initialize
    @patients = {}
  end

  def add_patient(patient_id, timestamp = 0.0)
    raise ArgumentError, "duplicate patient_id: #{patient_id}" if @patients.key?(patient_id)

    @patients[patient_id] = {
      id: patient_id,
      last_state_change: timestamp,
      current_state: "referred",
      state_history: {},
    }
  end

  def transition(patient_id, new_state, timestamp)
    patient = fetch_patient(patient_id)
    current = patient[:current_state]
    if TERMINAL_STATES.include?(current) || !ALLOWED_TRANSITIONS[current].include?(new_state)
      raise ArgumentError, "cannot transition from #{current} to #{new_state}"
    end

    patient[:state_history][current] = timestamp - patient[:last_state_change]
    patient[:current_state] = new_state
    patient[:last_state_change] = timestamp
  end

  def get_state(patient_id)
    fetch_patient(patient_id)[:current_state]
  end

  def get_patients_in_state(state)
    @patients.values.select { |p| p[:current_state] == state }.map { |p| p[:id] }.sort
  end

  def time_in_state(patient_id, state, as_of)
    patient = fetch_patient(patient_id)
    if patient[:state_history].key?(state)
      patient[:state_history][state]
    elsif patient[:current_state] == state
      as_of - patient[:last_state_change]
    else
      0.0
    end
  end

  def conversion_rate(from_state, to_state)
    exited = @patients.values.select { |p| p[:state_history].key?(from_state) }
    return 0.0 if exited.empty?

    converted = exited.select { |p| p[:current_state] == to_state || p[:state_history].key?(to_state) }
    converted.size.to_f / exited.size
  end

  def patients_overdue(state, max_seconds, as_of)
    @patients.values
      .select { |p| p[:current_state] == state && time_in_state(p[:id], state, as_of) > max_seconds }
      .sort_by { |p| -time_in_state(p[:id], state, as_of) }
      .map { |p| p[:id] }
  end

  def average_time_in_state(state, as_of)
    exited = @patients.values.select { |p| p[:state_history].key?(state) }
    return 0.0 if exited.empty?

    exited.sum { |p| p[:state_history][state] } / exited.size.to_f
  end

  private

  def fetch_patient(patient_id)
    @patients.fetch(patient_id) { raise KeyError, "patient not found: #{patient_id}" }
  end
end
