require "time"
require "securerandom"

class IncidentAggregator
  def initialize
    @reports = {}
    @incidents = {}
  end

  def ingest_report(report_id, source_id, event_type, location_key, ts)
    raise ArgumentError, "duplicate report_id: #{report_id}" if @reports.key?(report_id)

    report = {
      report_id: report_id,
      source_id: source_id,
      event_type: event_type,
      location_key: location_key,
      ts: ts,
      incident_id: nil,
    }
    @reports[report_id] = report
    report
  end

  def get_report(report_id)
    @reports[report_id]
  end

  def get_reports(location_key: nil, event_type: nil)
    reports = @reports.values
    reports = reports.select { |r| r[:location_key] == location_key } if location_key
    reports = reports.select { |r| r[:event_type] == event_type } if event_type
    reports.sort_by { |r| Time.parse(r[:ts]) }
  end

  def create_incident(incident_id, event_type, location_key)
    raise ArgumentError, "duplicate incident_id: #{incident_id}" if @incidents.key?(incident_id)

    incident = {
      incident_id: incident_id,
      event_type: event_type,
      location_key: location_key,
      report_ids: [],
      report_count: 0,
      latest_ts: nil,
    }
    @incidents[incident_id] = incident
    incident
  end

  def add_report_to_incident(incident_id, report_id)
    incident = fetch_incident(incident_id)
    report = fetch_report(report_id)
    raise ArgumentError, "report already assigned: #{report_id}" if report[:incident_id]

    report[:incident_id] = incident_id
    incident[:report_ids] = (incident[:report_ids] + [report_id])
      .sort_by { |id| Time.parse(@reports[id][:ts]) }
    incident[:report_count] += 1
    incident[:latest_ts] = report[:ts]
  end

  def get_incident(incident_id)
    @incidents[incident_id]
  end

  def get_unassigned_reports
    @reports.values.select { |r| r[:incident_id].nil? }.sort_by { |r| Time.parse(r[:ts]) }
  end

  def auto_ingest_report(report_id, source_id, event_type, location_key, ts, time_window_secs)
    ingest_report(report_id, source_id, event_type, location_key, ts)

    cutoff = Time.parse(ts) - time_window_secs
    candidates = @incidents.values.select do |incident|
      incident[:event_type] == event_type &&
        incident[:location_key] == location_key &&
        incident[:latest_ts] &&
        Time.parse(incident[:latest_ts]) >= cutoff
    end

    matching_incident =
      if candidates.any?
        candidates.min_by { |incident| [-Time.parse(incident[:latest_ts]).to_f, incident[:incident_id]] }
      else
        create_incident(SecureRandom.uuid, event_type, location_key)
      end

    add_report_to_incident(matching_incident[:incident_id], report_id)
    matching_incident[:incident_id]
  end

  def get_active_incidents(as_of_ts, time_window_secs)
    cutoff = Time.parse(as_of_ts) - time_window_secs
    @incidents.values
      .select { |incident| incident[:latest_ts] && Time.parse(incident[:latest_ts]) >= cutoff }
      .sort_by { |incident| -Time.parse(incident[:latest_ts]).to_f }
  end

  private

  def fetch_incident(incident_id)
    @incidents.fetch(incident_id) { raise KeyError, "incident not found: #{incident_id}" }
  end

  def fetch_report(report_id)
    @reports.fetch(report_id) { raise KeyError, "report not found: #{report_id}" }
  end
end
