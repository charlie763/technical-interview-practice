require "date"

TitrationEvent = Struct.new(:patient_id, :medication, :direction, :dose_mg, :recorded_on)
Medication = Struct.new(:name, :current_dose, :last_changed, :total_changes)

class TitrationTracker
  def initialize(events)
    @events = events.dup
  end

  def current_medications(patient_id)
    patient_events(patient_id)
      .group_by(&:medication)
      .filter_map do |name, events|
        sorted = events.sort_by(&:recorded_on)
        latest = sorted.last
        next if latest.direction == "stop"

        Medication.new(name, latest.dose_mg, latest.recorded_on, sorted.size)
      end
  end

  def get_medication_history(patient_id, medication)
    patient_events(patient_id)
      .select { |e| e.medication == medication }
      .sort_by(&:recorded_on)
  end

  def titration_count(patient_id, medication, direction: nil)
    history = get_medication_history(patient_id, medication)
    direction.nil? ? history.size : history.count { |e| e.direction == direction }
  end

  def de_escalation_summary(patient_id)
    patient_events(patient_id)
      .select { |e| %w[decrease stop].include?(e.direction) }
      .group_by(&:medication)
      .transform_values(&:size)
  end

  def patients_on_medication(medication)
    @events.map(&:patient_id).uniq.select do |patient_id|
      history = get_medication_history(patient_id, medication)
      !history.empty? && history.last.direction != "stop"
    end.sort
  end

  def most_titrated_medications(top_n: 3)
    @events
      .group_by(&:medication)
      .transform_values(&:size)
      .sort_by { |_name, count| -count }
      .first(top_n)
  end

  def add_event(event)
    @events.reject! do |e|
      e.patient_id == event.patient_id &&
        e.medication == event.medication &&
        e.recorded_on == event.recorded_on
    end
    @events << event
  end

  private

  def patient_events(patient_id)
    @events.select { |e| e.patient_id == patient_id }
  end
end
