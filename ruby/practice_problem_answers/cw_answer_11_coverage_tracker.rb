require "time"
require "set"

class CoverageTracker
  def initialize
    @stations = {}
    @heartbeats = {}
    @outages = {}
  end

  def register_station(station_id, name, region)
    raise ArgumentError, "duplicate station_id: #{station_id}" if @stations.key?(station_id)

    station = { station_id: station_id, name: name, region: region }
    @stations[station_id] = station
    @outages[station_id] = []
    station
  end

  def record_heartbeat(station_id, ts)
    fetch_station(station_id)
    last = @heartbeats[station_id]
    raise ArgumentError, "out-of-order heartbeat for #{station_id}" if last && Time.parse(ts) <= Time.parse(last)

    @heartbeats[station_id] = ts
  end

  def get_last_heartbeat(station_id)
    fetch_station(station_id)
    @heartbeats[station_id]
  end

  def get_stations(region: nil)
    stations = @stations.values
    stations = stations.select { |s| s[:region] == region } if region
    stations.sort_by { |s| s[:station_id] }
  end

  def get_stale_stations(as_of_ts, stale_after_secs)
    get_stations.select { |s| stale?(s[:station_id], as_of_ts, stale_after_secs) }
  end

  def record_outage_start(station_id, ts)
    fetch_station(station_id)
    outages = @outages[station_id]
    raise ArgumentError, "station already has an open outage: #{station_id}" if outages.any? { |o| o[:end_ts].nil? }

    outages << { station_id: station_id, start_ts: ts, end_ts: nil }
  end

  def record_outage_end(station_id, ts)
    fetch_station(station_id)
    outage = @outages[station_id].reverse.find { |o| o[:end_ts].nil? }
    raise ArgumentError, "no open outage for #{station_id}" unless outage

    outage[:end_ts] = ts
  end

  def get_outages(station_id)
    fetch_station(station_id)
    @outages[station_id].sort_by { |o| Time.parse(o[:start_ts]) }
  end

  def get_region_coverage(region, as_of_ts, stale_after_secs)
    stations = get_stations(region: region)
    stale_ids = get_stale_stations(as_of_ts, stale_after_secs).map { |s| s[:station_id] }.to_set
    stale_count = stations.count { |s| stale_ids.include?(s[:station_id]) }
    healthy_count = stations.size - stale_count

    {
      region: region,
      total: stations.size,
      healthy: healthy_count,
      stale: stale_count,
      has_coverage: healthy_count >= 1,
    }
  end

  def get_outage_summary(station_id, as_of_ts)
    outages = get_outages(station_id)
    total_secs = outages.sum do |o|
      end_time = o[:end_ts] ? Time.parse(o[:end_ts]) : Time.parse(as_of_ts)
      (end_time - Time.parse(o[:start_ts])).to_i
    end

    {
      station_id: station_id,
      total_outages: outages.size,
      open_outage: outages.any? { |o| o[:end_ts].nil? },
      total_outage_secs: total_secs,
    }
  end

  private

  def stale?(station_id, as_of_ts, stale_after_secs)
    last = @heartbeats[station_id]
    return true if last.nil?

    Time.parse(as_of_ts) - Time.parse(last) > stale_after_secs
  end

  def fetch_station(station_id)
    @stations.fetch(station_id) { raise KeyError, "station not found: #{station_id}" }
  end
end
