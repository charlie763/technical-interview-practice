class DispatchManager
  def initialize
    @responders = {}
    @incidents = {}
  end

  def register_responder(responder_id, name, subscribed_types, capacity)
    raise ArgumentError, "duplicate responder_id: #{responder_id}" if @responders.key?(responder_id)

    responder = {
      responder_id: responder_id,
      name: name,
      subscribed_types: subscribed_types,
      capacity: capacity,
    }
    @responders[responder_id] = responder
    responder
  end

  def add_incident(incident_id, incident_type, severity, ts)
    raise ArgumentError, "duplicate incident_id: #{incident_id}" if @incidents.key?(incident_id)

    incident = {
      incident_id: incident_id,
      incident_type: incident_type,
      severity: severity,
      ts: ts,
      responder_id: nil,
      resolved: false,
    }
    @incidents[incident_id] = incident
    incident
  end

  def get_incidents_for_responder(responder_id)
    responder = fetch_responder(responder_id)
    @incidents.values
      .select { |i| responder[:subscribed_types].include?(i[:incident_type]) }
      .sort_by { |i| [-i[:severity], i[:ts]] }
  end

  def assign_incident(incident_id, responder_id)
    incident = fetch_incident(incident_id)
    responder = fetch_responder(responder_id)
    raise ArgumentError, "incident already assigned: #{incident_id}" if incident[:responder_id]
    raise ArgumentError, "responder at capacity: #{responder_id}" if open_assignment_count(responder_id) >= responder[:capacity]

    incident[:responder_id] = responder_id
  end

  def resolve_incident(incident_id)
    incident = fetch_incident(incident_id)
    raise ArgumentError, "already resolved: #{incident_id}" if incident[:resolved]

    incident[:resolved] = true
  end

  def get_open_assignments(responder_id)
    fetch_responder(responder_id)
    @incidents.values
      .select { |i| i[:responder_id] == responder_id && !i[:resolved] }
      .sort_by { |i| [-i[:severity], i[:ts]] }
  end

  def auto_assign(incident_id)
    incident = fetch_incident(incident_id)
    raise ArgumentError, "incident already assigned: #{incident_id}" if incident[:responder_id]

    eligible = @responders.values.select do |r|
      r[:subscribed_types].include?(incident[:incident_type]) &&
        open_assignment_count(r[:responder_id]) < r[:capacity]
    end
    raise ArgumentError, "no eligible responder for #{incident_id}" if eligible.empty?

    chosen = eligible.min_by { |r| [open_assignment_count(r[:responder_id]), -r[:capacity], r[:responder_id]] }
    assign_incident(incident_id, chosen[:responder_id])
    chosen[:responder_id]
  end

  def get_dispatch_summary
    @responders.values.sort_by { |r| r[:responder_id] }.map do |r|
      open_count = open_assignment_count(r[:responder_id])
      {
        responder_id: r[:responder_id],
        name: r[:name],
        capacity: r[:capacity],
        open_count: open_count,
        available_capacity: r[:capacity] - open_count,
      }
    end
  end

  private

  def open_assignment_count(responder_id)
    @incidents.values.count { |i| i[:responder_id] == responder_id && !i[:resolved] }
  end

  def fetch_responder(responder_id)
    @responders.fetch(responder_id) { raise KeyError, "responder not found: #{responder_id}" }
  end

  def fetch_incident(incident_id)
    @incidents.fetch(incident_id) { raise KeyError, "incident not found: #{incident_id}" }
  end
end
