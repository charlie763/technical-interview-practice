"""
Notes

Questions
- whats the relationship from incidents to reports? one to one or can reports have multiple incidences
 -> one to many, incident -> reports
"""

"""
=============================================================================
INTERVIEW PROBLEM 9: Multi-Source Incident Aggregator
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building the incident aggregation layer for a public safety intelligence
platform. The platform ingests incident reports from multiple independent data
sources (radio dispatch transcriptions, field sensors, social media monitors).
Different sources frequently report the same real-world event, so the system
must deduplicate and group raw reports into unified incidents.

For this problem you are building an IncidentAggregator class.
Store all state in instance variables initialized in `__init__`.
Class-level variables will bleed between tests and between IncidentAggregator
instances — avoid them.
You choose the internal data structures; the public interface is what matters.

DATA MODEL
----------
Report:
  {
    "report_id":    str,
    "source_id":    str,
    "event_type":   str,        # e.g. "shooting", "car-crash", "fire"
    "location_key": str,        # opaque string, e.g. "downtown", "sector-7"
    "ts":           str,        # ISO-8601 timestamp (no timezone offset),
                                # e.g. "2024-01-01T10:00:00"
    "incident_id":  str | None  # None until assigned to an incident
  }

Incident:
  {
    "incident_id":  str,
    "event_type":   str,        # set at creation time
    "location_key": str,        # set at creation time
    "report_ids":   list[str],  # report IDs in ts-ascending order
    "report_count": int,
    "latest_ts":    str | None  # ts of the most recently added report, or None
  }

Timestamps are ISO-8601 strings without timezone offset. Use
datetime.fromisoformat() for arithmetic when comparing or computing durations.

# Example
# agg = IncidentAggregator()
# agg.ingest_report("r1", "radio-north", "shooting", "downtown", "2024-01-01T10:00:00")
# agg.ingest_report("r2", "radio-south", "shooting", "downtown", "2024-01-01T10:00:45")
# agg.ingest_report("r3", "social-feed", "car-crash", "midtown",  "2024-01-01T10:01:00")
# inc = agg.create_incident("inc-001", "shooting", "downtown")
# agg.add_report_to_incident("inc-001", "r1")
# agg.add_report_to_incident("inc-001", "r2")
# agg.get_incident("inc-001")["report_count"]           # -> 2
# agg.get_unassigned_reports()                           # -> [r3 report dict]
# agg.auto_ingest_report("r4", "radio-east", "shooting",
#                         "downtown", "2024-01-01T10:01:30",
#                         time_window_secs=120)
# # -> "inc-001"  (within 120 s window, same type + location)
=============================================================================
"""

from typing import Optional
from datetime import datetime
from uuid import uuid4


class IncidentAggregator:
    def __init__(self):
        self.reports: dict[str, dict] = {}  # {<report_id>: Report}
        self.incidents: dict[str, dict] = {}  # {<incident_id>: Incident}

    # -------------------------------------------------------------------------
    # PART 1 — Report ingestion  (~10 min)
    # -------------------------------------------------------------------------

    def ingest_report(
        self,
        report_id: str,
        source_id: str,
        event_type: str,
        location_key: str,
        ts: str,
    ) -> dict:
        """
        Store a new raw report and return it.
        The report's incident_id starts as None.
        Raise ValueError if report_id already exists.
        """
        existing_report = self.reports.get(report_id)
        if existing_report:
            raise ValueError
        new_report = {
            "report_id": report_id,
            "source_id": source_id,
            "event_type": event_type,
            "location_key": location_key,
            "ts": ts,
            "incident_id": None,
        }
        self.reports[report_id] = new_report
        return new_report

    #         Report:
    #   {
    #     "report_id":    str,
    #     "source_id":    str,
    #     "event_type":   str,        # e.g. "shooting", "car-crash", "fire"
    #     "location_key": str,        # opaque string, e.g. "downtown", "sector-7"
    #     "ts":           str,        # ISO-8601 timestamp (no timezone offset),
    #                                 # e.g. "2024-01-01T10:00:00"
    #     "incident_id":  str | None  # None until assigned to an incident
    #   }

    def get_report(self, report_id: str) -> Optional[dict]:
        """Return the report dict, or None if not found."""
        return self.reports.get(report_id)

    def get_reports(
        self,
        *,
        location_key: Optional[str] = None,
        event_type: Optional[str] = None,
    ) -> list:
        """
        Return all reports, optionally filtered by location_key and/or
        event_type (both filters applied when both are given).
        Results are sorted by ts ascending.
        """
        all_reports = self.reports.values()
        if location_key:
            all_reports = filter(
                lambda report: report["location_key"] == location_key, all_reports
            )
        if event_type:
            all_reports = filter(
                lambda report: report["event_type"] == event_type, all_reports
            )
        return sorted(
            all_reports, key=lambda report: datetime.fromisoformat(report["ts"])
        )

    # -------------------------------------------------------------------------
    # PART 2 — Manual incident grouping  (~15 min)
    # -------------------------------------------------------------------------

    def create_incident(
        self, incident_id: str, event_type: str, location_key: str
    ) -> dict:
        """
        Create and return a new, empty incident with the given event_type and
        location_key.
        Raise ValueError if incident_id already exists.
        """
        existing_incident = self.incidents.get(incident_id)
        if existing_incident:
            raise ValueError
        new_incident = {
            "incident_id": incident_id,
            "event_type": event_type,
            "location_key": location_key,
            "report_ids": [],
            "report_count": 0,
            "latest_ts": None,
        }
        self.incidents[incident_id] = new_incident
        return new_incident

    #         Incident:
    #   {
    #     "incident_id":  str,
    #     "event_type":   str,        # set at creation time
    #     "location_key": str,        # set at creation time
    #     "report_ids":   list[str],  # report IDs in ts-ascending order
    #     "report_count": int,
    #     "latest_ts":    str | None  # ts of the most recently added report, or None
    #   }

    def add_report_to_incident(self, incident_id: str, report_id: str) -> None:
        """
        Assign a report to an incident.
        - Raise KeyError  if incident_id or report_id does not exist.
        - Raise ValueError if the report is already assigned to any incident.
        - Sets report["incident_id"] = incident_id.
        - Updates the incident's report_ids (kept in ts-ascending order),
          report_count, and latest_ts.
        """
        existing_incident = self.incidents.get(incident_id)
        existing_report = self.reports.get(report_id)
        if not existing_incident or not existing_report:
            raise KeyError
        # print(f"incident_id: {incident_id}")
        # print(f"report_id: {report_id}")
        # print(f"self.incidents.values(): {self.incidents.values()}")
        # print(f"existing_report: {existing_report}")

        if existing_report["incident_id"] or any(
            [
                report_id in incident["report_ids"]
                for incident in self.incidents.values()
            ]
        ):  # do we also need to check incident records? -> didn't actually need, was just returning instead of raising originally
            raise ValueError
        existing_report["incident_id"] = incident_id
        # improvement: either very that additions happen later cronologically or don't do full sort everytime
        new_reports = [
            report
            for id, report in self.reports.items()
            if id in existing_incident["report_ids"]
        ]
        new_reports.append(existing_report)
        sorted_reports = sorted(
            new_reports, key=lambda report: datetime.fromisoformat(report["ts"])
        )
        existing_incident["report_ids"] = [
            report["report_id"] for report in sorted_reports
        ]
        existing_incident["report_count"] = existing_incident["report_count"] + 1
        existing_incident["latest_ts"] = existing_report["ts"]

    def get_incident(self, incident_id: str) -> Optional[dict]:
        """
        Return the incident dict (including up-to-date report_ids, report_count,
        and latest_ts), or None if not found.
        report_ids must be ordered by the corresponding report's ts, ascending.
        """
        return self.incidents.get(incident_id)

    def get_unassigned_reports(self) -> list:
        """
        Return all reports whose incident_id is still None, sorted by ts
        ascending.
        """
        unassigned_reports = filter(
            lambda report: report["incident_id"] == None, self.reports.values()
        )
        return sorted(
            unassigned_reports, key=lambda report: datetime.fromisoformat(report["ts"])
        )

    # -------------------------------------------------------------------------
    # PART 3 — Automatic deduplication  (~20 min)
    # -------------------------------------------------------------------------

    def auto_ingest_report(
        self,
        report_id: str,
        source_id: str,
        event_type: str,
        location_key: str,
        ts: str,
        time_window_secs: int,
    ) -> str:
        """
        Ingest a new report and automatically assign it to an incident:

        1. Call ingest_report to store the report.
        2. Find all *active* incidents whose event_type and location_key match
           the incoming report's.  An incident is "active" if its latest_ts is
           within time_window_secs of the new report's ts:
               latest_ts >= ts - time_window_secs
           Incidents with no reports (latest_ts is None) are not active.
        3. If one or more matches exist, pick the one whose latest_ts is closest
           to ts (i.e. most recently active).  Break ties by incident_id
           lexicographically ascending.
        4. If no active match exists, create a new incident (auto-generate a
           unique incident_id; any scheme is fine as long as it doesn't clash
           with existing IDs).
        5. Call add_report_to_incident to assign the report.
        6. Return the incident_id.
        """
        self.ingest_report(
            report_id=report_id,
            source_id=source_id,
            event_type=event_type,
            location_key=location_key,
            ts=ts,
        )
        # get_releveant incidents is copied from get rpoerts logic, might be godo to create helper method or something
        active_incidents = self.incidents.values()
        active_incidents = list(
            filter(
                lambda incident: incident["location_key"] == location_key
                and incident["event_type"] == event_type
                and incident["latest_ts"]
                and datetime.fromisoformat(incident["latest_ts"]).timestamp()
                >= datetime.fromisoformat(ts).timestamp() - time_window_secs,
                active_incidents,
            )
        )
        matching_incident = None
        # assume ts is only a variable for purposes of testing therefor can just find max latest ts
        if len(active_incidents) > 0:
            matching_incident = max(
                active_incidents,
                key=lambda incident: (incident["latest_ts"], incident["incident_id"]),
            )
        else:
            matching_incident = self.create_incident(
                incident_id=uuid4(), event_type=event_type, location_key=location_key
            )
        matching_incident_id = matching_incident["incident_id"]
        self.add_report_to_incident(
            incident_id=matching_incident_id, report_id=report_id
        )
        return matching_incident_id

    def get_active_incidents(self, as_of_ts: str, time_window_secs: int) -> list:
        """
        Return all incidents that have a latest_ts within time_window_secs of
        as_of_ts:
            latest_ts >= as_of_ts - time_window_secs

        Sorted by latest_ts descending (most-recently-active first).
        Incidents with no reports (latest_ts is None) are excluded.
        """
        active_incidents = self.incidents.values()
        active_incidents = list(
            filter(
                lambda incident: incident["latest_ts"]
                and datetime.fromisoformat(incident["latest_ts"]).timestamp()
                >= datetime.fromisoformat(as_of_ts).timestamp() - time_window_secs,
                active_incidents,
            )
        )
        return sorted(
            active_incidents,
            key=lambda incident: datetime.fromisoformat(incident["latest_ts"]),
            reverse=True,
        )
