require "date"

GLUCOSE_RANGE = (70.0..180.0)
KETONE_RANGE = (0.5..3.0)

BiomarkerReading = Struct.new(:patient_id, :reading_type, :value, :recorded_on)

class BiomarkerMonitor
  def initialize(readings)
    @readings = readings
  end

  def is_out_of_range(reading)
    case reading.reading_type
    when "glucose" then !GLUCOSE_RANGE.cover?(reading.value)
    when "ketone" then !KETONE_RANGE.cover?(reading.value)
    else false
    end
  end

  def max_consecutive_out_of_range_days(patient_id, reading_type)
    by_day = readings_by_day(patient_id, reading_type)
    return 0 if by_day.empty?

    longest = 0
    current = 0
    previous_day = nil

    by_day.keys.sort.each do |day|
      if by_day[day].any? { |r| is_out_of_range(r) }
        current = previous_day && (day - previous_day == 1) ? current + 1 : 1
        longest = current if current > longest
      else
        current = 0
      end
      previous_day = day
    end

    longest
  end

  def get_outreach_list(min_consecutive_days: 3)
    combos = @readings.map { |r| [r.patient_id, r.reading_type] }.uniq

    entries = combos.filter_map do |patient_id, reading_type|
      streak = max_consecutive_out_of_range_days(patient_id, reading_type)
      next if streak < min_consecutive_days

      latest = readings_by_day(patient_id, reading_type)
        .values.flatten
        .select { |r| is_out_of_range(r) }
        .max_by(&:recorded_on)

      {
        patient_id: patient_id,
        reading_type: reading_type,
        consecutive_days: streak,
        latest_value: latest.value,
      }
    end

    entries.sort_by { |e| -e[:consecutive_days] }
  end

  def add_reading(reading)
    same_day = @readings.select do |r|
      r.patient_id == reading.patient_id &&
        r.reading_type == reading.reading_type &&
        r.recorded_on == reading.recorded_on
    end
    is_duplicate = same_day.any? { |r| (r.value - reading.value).abs <= 0.5 }
    return false if is_duplicate

    @readings << reading
    true
  end

  private

  def readings_by_day(patient_id, reading_type)
    @readings
      .select { |r| r.patient_id == patient_id && r.reading_type == reading_type }
      .group_by(&:recorded_on)
  end
end
