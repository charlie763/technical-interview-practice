# =============================================================================
# INTERVIEW PROBLEM 11: Sensor Coverage Tracker
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building the health-monitoring subsystem for a platform that
# deploys radio-receiver sensor stations across geographic regions. Each
# station periodically sends a heartbeat. When a station falls silent,
# operators need to know, and the platform's incident-detection coverage
# for that region may be affected.
#
# For this problem you are building a CoverageTracker class. Store all
# state in instance variables initialized in `initialize`. Class-level
# variables will bleed between tests and between CoverageTracker instances
# — avoid them. You choose the internal data structures; the public
# interface is what matters.
#
# DATA MODEL
# ----------
# Station:
#   {
#     station_id: String,
#     name:       String,
#     region:     String,  # logical grouping, e.g. "downtown", "sector-7"
#   }
#
# Outage:
#   {
#     station_id: String,
#     start_ts:   String,       # ISO-8601 when the outage began
#     end_ts:     String | nil, # ISO-8601 when the outage ended; nil = ongoing
#   }
#
# Timestamps are ISO-8601 strings without timezone offset (e.g.
# "2024-01-01T10:00:00"). Use Time.parse (from the "time" library) for
# arithmetic.
#
# Example
# -------
#   ct = CoverageTracker.new
#   ct.register_station("sta-001", "North Tower", "downtown")
#   ct.register_station("sta-002", "South Tower", "downtown")
#   ct.record_heartbeat("sta-001", "2024-01-01T10:00:00")
#   ct.record_heartbeat("sta-002", "2024-01-01T10:00:05")
#   ct.get_last_heartbeat("sta-001")                   # -> "2024-01-01T10:00:00"
#   ct.get_stale_stations("2024-01-01T10:05:00", 120)  # -> []
#   ct.record_outage_start("sta-001", "2024-01-01T10:10:00")
#   ct.get_region_coverage("downtown", "2024-01-01T10:10:30", 120)
#   # -> { region: "downtown", total: 2, healthy: 1, stale: 1, has_coverage: true }
# =============================================================================

require "time"

class CoverageTracker
  def initialize
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 1 — Station registration and heartbeats (~10 min)
  # -------------------------------------------------------------------------

  # Register a new station and return it.
  # Raise ArgumentError if station_id already exists.
  def register_station(station_id, name, region)
    raise NotImplementedError
  end

  # Record a heartbeat for the station.
  # - Raise KeyError if station_id does not exist.
  # - Raise ArgumentError if ts is earlier than or equal to the station's
  #   most recent heartbeat (out-of-order and duplicate heartbeats are
  #   rejected).
  def record_heartbeat(station_id, ts)
    raise NotImplementedError
  end

  # Return the timestamp of the most recent heartbeat, or nil if the
  # station has never sent one.
  # Raise KeyError if station_id does not exist.
  def get_last_heartbeat(station_id)
    raise NotImplementedError
  end

  # Return all stations, optionally filtered to a specific region.
  # Sorted by station_id ascending.
  def get_stations(region: nil)
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 2 — Staleness detection and outage tracking (~15 min)
  # -------------------------------------------------------------------------

  # Return station hashes for all stations that are stale as of as_of_ts.
  # A station is stale if:
  #   - It has never sent a heartbeat, OR
  #   - Its last heartbeat was more than stale_after_secs seconds before
  #     as_of_ts (i.e. as_of_ts - last_heartbeat > stale_after_secs).
  #
  # Results are sorted by station_id ascending.
  def get_stale_stations(as_of_ts, stale_after_secs)
    raise NotImplementedError
  end

  # Open a new outage record for the station (end_ts = nil).
  # - Raise KeyError if station_id does not exist.
  # - Raise ArgumentError if the station already has an open outage
  #   (an outage with end_ts = nil).
  def record_outage_start(station_id, ts)
    raise NotImplementedError
  end

  # Close the most recent open outage for the station by setting its
  # end_ts = ts.
  # - Raise KeyError if station_id does not exist.
  # - Raise ArgumentError if the station has no open outage.
  def record_outage_end(station_id, ts)
    raise NotImplementedError
  end

  # Return all Outage hashes for the station, sorted by start_ts ascending.
  # Raise KeyError if station_id does not exist.
  def get_outages(station_id)
    raise NotImplementedError
  end

  # -------------------------------------------------------------------------
  # PART 3 — Coverage analysis (~20 min)
  # -------------------------------------------------------------------------

  # Return a coverage summary for the region:
  #   {
  #     region:       String,
  #     total:        Integer,  # stations in this region
  #     healthy:      Integer,  # stations NOT stale
  #     stale:        Integer,  # stations that ARE stale
  #     has_coverage: true | false,  # true if healthy >= 1
  #   }
  #
  # Use get_stations (Part 1) and get_stale_stations (Part 2) internally.
  def get_region_coverage(region, as_of_ts, stale_after_secs)
    raise NotImplementedError
  end

  # Return an outage summary for the station:
  #   {
  #     station_id:         String,
  #     total_outages:      Integer,  # number of outage records
  #     open_outage:        true | false,  # true if there is a current open outage
  #     total_outage_secs:  Integer,  # cumulative outage duration in seconds
  #   }
  #
  # For an open outage (end_ts is nil), count duration from start_ts up to
  # as_of_ts.
  #
  # Use get_outages (Part 2) internally.
  # Raise KeyError if station_id does not exist.
  def get_outage_summary(station_id, as_of_ts)
    raise NotImplementedError
  end
end
