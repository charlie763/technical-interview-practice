"""
=============================================================================
INTERVIEW PROBLEM 11: Sensor Coverage Tracker
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the health-monitoring subsystem for a platform that deploys
radio-receiver sensor stations across geographic regions. Each station
periodically sends a heartbeat. When a station falls silent, operators need
to know, and the platform's incident-detection coverage for that region may
be affected.

For this problem you are building a CoverageTracker class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between CoverageTracker
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Station:
  {
    "station_id": str,
    "name":       str,
    "region":     str,  # logical grouping, e.g. "downtown", "sector-7"
  }

Outage:
  {
    "station_id": str,
    "start_ts":   str,        # ISO-8601 when the outage began
    "end_ts":     str | None, # ISO-8601 when the outage ended; None = ongoing
  }

Timestamps are ISO-8601 strings without timezone offset (e.g.
"2024-01-01T10:00:00"). Use datetime.fromisoformat() for arithmetic.

# Example
# ct = CoverageTracker()
# ct.register_station("sta-001", "North Tower", region="downtown")
# ct.register_station("sta-002", "South Tower", region="downtown")
# ct.record_heartbeat("sta-001", "2024-01-01T10:00:00")
# ct.record_heartbeat("sta-002", "2024-01-01T10:00:05")
# ct.get_last_heartbeat("sta-001")              # -> "2024-01-01T10:00:00"
# ct.get_stale_stations("2024-01-01T10:05:00", stale_after_secs=120)  # -> []
# ct.record_outage_start("sta-001", "2024-01-01T10:10:00")
# ct.get_region_coverage("downtown", "2024-01-01T10:10:30", stale_after_secs=120)
# # -> {"region": "downtown", "total": 2, "healthy": 1, "stale": 1,
# #     "has_coverage": True}
=============================================================================
"""

from typing import Optional


class CoverageTracker:
    def __init__(self):
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 1 — Station registration and heartbeats  (~10 min)
    # -------------------------------------------------------------------------

    def register_station(self, station_id: str, name: str, region: str) -> dict:
        """
        Register a new station and return it.
        Raise ValueError if station_id already exists.
        """
        raise NotImplementedError

    def record_heartbeat(self, station_id: str, ts: str) -> None:
        """
        Record a heartbeat for the station.
        - Raise KeyError   if station_id does not exist.
        - Raise ValueError if ts is earlier than or equal to the station's most
          recent heartbeat (out-of-order and duplicate heartbeats are rejected).
        """
        raise NotImplementedError

    def get_last_heartbeat(self, station_id: str) -> Optional[str]:
        """
        Return the timestamp of the most recent heartbeat, or None if the
        station has never sent one.
        Raise KeyError if station_id does not exist.
        """
        raise NotImplementedError

    def get_stations(self, region: Optional[str] = None) -> list:
        """
        Return all stations, optionally filtered to a specific region.
        Sorted by station_id ascending.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 2 — Staleness detection and outage tracking  (~15 min)
    # -------------------------------------------------------------------------

    def get_stale_stations(self, as_of_ts: str, stale_after_secs: int) -> list:
        """
        Return station dicts for all stations that are stale as of as_of_ts.
        A station is stale if:
          - It has never sent a heartbeat, OR
          - Its last heartbeat was more than stale_after_secs seconds before
            as_of_ts  (i.e. as_of_ts - last_heartbeat > stale_after_secs).

        Results are sorted by station_id ascending.
        """
        raise NotImplementedError

    def record_outage_start(self, station_id: str, ts: str) -> None:
        """
        Open a new outage record for the station (end_ts = None).
        - Raise KeyError   if station_id does not exist.
        - Raise ValueError if the station already has an open outage
          (an outage with end_ts = None).
        """
        raise NotImplementedError

    def record_outage_end(self, station_id: str, ts: str) -> None:
        """
        Close the most recent open outage for the station by setting its
        end_ts = ts.
        - Raise KeyError   if station_id does not exist.
        - Raise ValueError if the station has no open outage.
        """
        raise NotImplementedError

    def get_outages(self, station_id: str) -> list:
        """
        Return all Outage dicts for the station, sorted by start_ts ascending.
        Raise KeyError if station_id does not exist.
        """
        raise NotImplementedError

    # -------------------------------------------------------------------------
    # PART 3 — Coverage analysis  (~20 min)
    # -------------------------------------------------------------------------

    def get_region_coverage(
        self, region: str, as_of_ts: str, stale_after_secs: int
    ) -> dict:
        """
        Return a coverage summary for the region:
          {
            "region":       str,
            "total":        int,   # stations in this region
            "healthy":      int,   # stations NOT stale
            "stale":        int,   # stations that ARE stale
            "has_coverage": bool,  # True if healthy >= 1
          }

        Use get_stations (Part 1) and get_stale_stations (Part 2) internally.
        """
        raise NotImplementedError

    def get_outage_summary(self, station_id: str, as_of_ts: str) -> dict:
        """
        Return an outage summary for the station:
          {
            "station_id":        str,
            "total_outages":     int,   # number of outage records
            "open_outage":       bool,  # True if there is a current open outage
            "total_outage_secs": int,   # cumulative outage duration in seconds
          }

        For an open outage (end_ts is None), count duration from start_ts up
        to as_of_ts.

        Use get_outages (Part 2) internally.
        Raise KeyError if station_id does not exist.
        """
        raise NotImplementedError
