"""
=============================================================================
INTERVIEW PROBLEM 1: Geofence Alert Rule Engine
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
You're building a backend service for an IoT asset-tracking platform. Physical
assets (forklifts, shipping containers, field equipment) carry GPS sensors that
periodically report coordinates. The platform tracks which geographic "zone"
(geofence) each asset is currently inside, and fires configured alert rules
whenever an asset transitions between zones.

For this problem, zones are axis-aligned bounding boxes — no geospatial
libraries needed.

DATA MODEL
----------
All state lives in a single TrackerState dict (returned by make_tracker).

Zone:
  {"id": str, "name": str, "bounds": {"min_lat": float, "max_lat": float,
                                       "min_lng": float, "max_lng": float}}

Asset:
  {"id": str, "name": str, "lat": float | None, "lng": float | None, "zone_id": str | None}
  - lat/lng are None until the first GPS update arrives.
  - zone_id is the id of the zone the asset is currently in, or None.

AlertRule:
  {"id": str, "from_zone_id": str | None, "to_zone_id": str | None, "asset_id": str | None}
  - None in from_zone_id matches ANY previous zone (including None/"no zone").
  - None in to_zone_id  matches ANY new zone (including None/"no zone").
  - None in asset_id    matches ANY asset.

TriggeredAlert (appended to alert_log when a rule fires):
  {"rule_id": str, "asset_id": str, "from_zone_id": str | None,
   "to_zone_id": str | None, "timestamp": str}

TrackerState:
  {
    "zones":       {zone_id: Zone},
    "assets":      {asset_id: Asset},
    "alert_rules": [AlertRule],
    "alert_log":   [TriggeredAlert],
  }
=============================================================================
"""

from typing import Optional


def make_tracker() -> dict:
    """Return a fresh, empty TrackerState."""
    return {"zones": {}, "assets": {}, "alert_rules": [], "alert_log": []}


# ---------------------------------------------------------------------------
# PART 1 — Zone membership  (warm-up, ~5 min)
# ---------------------------------------------------------------------------

def is_in_zone(asset: dict, zone: dict) -> bool:
    """
    Return True if the asset's lat/lng falls inside the zone's bounding box.
    - Bounds are inclusive on all edges.
    - Return False if the asset has no location (lat or lng is None).
    """
    asset_lat = asset["lat"]
    asset_lng = asset["lng"]

    if (asset_lng is None) or (asset_lat is None):
        return False

    is_in_lat_bounds = (zone["bounds"]["min_lat"] <= asset_lat) and (asset_lat <= zone["bounds"]["max_lat"])
    is_in_lng_bounds = (zone["bounds"]["min_lng"] <= asset_lng) and (asset_lng <= zone["bounds"]["max_lng"])

    if is_in_lat_bounds and is_in_lng_bounds:
        return True
    else:
        return False

# ---------------------------------------------------------------------------
# PART 2 — Locate an asset  (~5 min)
# ---------------------------------------------------------------------------

def get_current_zone_id(state: dict, asset_id: str) -> Optional[str]:
    """
    Return the id of the first zone in state that contains the asset, or None.
    - Return None if the asset doesn't exist or has no location.
    - Iterate zones in insertion order (standard Python dict behaviour).
    """

    zone_id = None
    asset = state["assets"].get(asset_id)

    if asset:
        for zone in state["zones"].values():
            if is_in_zone(asset, zone):
                # update the asset's zone_id if the asset is in zone
                asset["zone_id"] = zone["id"]
                return zone["id"]
    return None


# ---------------------------------------------------------------------------
# PART 3 — Process a location update  (core logic, ~15 min)
# ---------------------------------------------------------------------------

def process_location_update(
    state: dict,
    asset_id: str,
    lat: float,
    lng: float,
    timestamp: str,
) -> list:
    """
    Handle a new GPS reading for an asset:

    1. Update the asset's lat and lng in state.
    2. Recompute zone_id via get_current_zone_id and store it on the asset.
    3. If zone_id changed (including None→zone or zone→None), evaluate all
       alert_rules and collect any that match.
    4. Append each matched rule as a TriggeredAlert to state["alert_log"].
    5. Return the list of newly triggered alerts (empty list if no zone change
       or no rule matches).

    Raise KeyError if asset_id is not in state["assets"].

    Alert-rule matching — a rule matches when ALL three conditions hold:
        rule["asset_id"]     is None  OR  rule["asset_id"]     == asset_id
        rule["from_zone_id"] is None  OR  rule["from_zone_id"] == old_zone_id
        rule["to_zone_id"]   is None  OR  rule["to_zone_id"]   == new_zone_id
    """
    '''
    TrackerState:
    {
        "zones":       {zone_id: Zone},
        "assets":      {asset_id: Asset},
        "alert_rules": [AlertRule],
        "alert_log":   [TriggeredAlert],
    }

    TriggeredAlert (appended to alert_log when a rule fires):
        {"rule_id": str, "asset_id": str, "from_zone_id": str | None,
        "to_zone_id": str | None, "timestamp": str}

    AlertRule:
        {"id": str, "from_zone_id": str | None, "to_zone_id": str | None, "asset_id": str | None}
        - None in from_zone_id matches ANY previous zone (including None/"no zone").
        - None in to_zone_id  matches ANY new zone (including None/"no zone").
        - None in asset_id    matches ANY asset.
    '''
    old_zone_id = state["assets"][asset_id]['zone_id']
    state["assets"][asset_id]["lat"] = lat
    state["assets"][asset_id]["lng"] = lng
    new_zone_id = get_current_zone_id(state, asset_id)

    triggered_alerts = []
    if old_zone_id != new_zone_id:
        for rule in state["alert_rules"]:
            if ((rule["asset_id"] is None or rule["asset_id"] == asset_id) and
                (rule["from_zone_id"] is None or rule["from_zone_id"] == old_zone_id) and
                (rule["to_zone_id"] is None or rule["to_zone_id"] == new_zone_id)
            ):
                trigger_alert = {
                    "rule_id": rule["id"],
                    "asset_id": asset_id,
                    "from_zone_id": rule["from_zone_id"],
                    "to_zone_id": rule["to_zone_id"],
                    "timestamp": timestamp
                }
                triggered_alerts.append(trigger_alert)
                state["alert_log"].append(trigger_alert)
    print(triggered_alerts)
    return triggered_alerts



# ---------------------------------------------------------------------------
# PART 4 — CRUD helpers  (~15 min)
# ---------------------------------------------------------------------------

def add_zone(
    state: dict,
    zone_id: str,
    name: str,
    min_lat: float,
    max_lat: float,
    min_lng: float,
    max_lng: float,
) -> dict:
    """
    Create a zone, store it in state, and return it.

    Zone:
        {"id": str, "name": str, "bounds": {"min_lat": float, "max_lat": float,
                                       "min_lng": float, "max_lng": float}}

    Raise ValueError if zone_id already exists.
    """
    try:
        if state["zones"].get(zone_id): # use get in case doesn't exist
            raise ValueError
        new_zone = {
            "id": zone_id,
            "name": name,
            "bounds": {
                "min_lat": min_lat,
                "max_lat": max_lat,
                "min_lng": min_lng,
                "max_lng": max_lng
            }
        }
        state["zones"][zone_id] = new_zone
        return new_zone

    # typically irl we should log where failures happen with more specific messaging
    except ValueError as error:
        print(f'zone_id already exists: {error}')
        raise

def remove_zone(state: dict, zone_id: str) -> None:
    """
    Remove a zone from state. Raise KeyError if not found.
    Any asset currently assigned to the removed zone should have its
    zone_id set to None. Do NOT fire alert rules for this forced change.
    """
    try:
        if not state["zones"].get(zone_id):  # use get in case doesn't exist
            raise KeyError

        # if asset in removed zone, update asset's zone_id to None
        for asset in state["assets"].values():
            if asset["zone_id"] == zone_id:
                asset["zone_id"] = None

        # alternatively use pop if we want to process the removed zone
        del state["zones"][zone_id]

    # typically irl we should log where failures happen with more specific messaging
    except KeyError as error:
        print(f'Zone does not exist in state: {error}')
        raise
    return None

def add_asset(state: dict, asset_id: str, name: str) -> dict:
    """
    Create an asset with no initial location (lat=None, lng=None, zone_id=None),
    store it in state, and return it.

    Asset:
        {"id": str, "name": str, "lat": float | None, "lng": float | None, "zone_id": str | None}

    Raise ValueError if asset_id already exists.
    """
    try:
        if state["assets"].get(asset_id): # use get in case doesn't exist
            raise ValueError
        new_asset = {
            "id" : asset_id,
            "name": name,
            "lat": None,
            "lng": None,
            "zone_id": None
        }
        state["assets"][asset_id] = new_asset
        return new_asset

    # typically irl we should log where failures happen with more specific messaging
    except ValueError as error:
        print(f'asset already exists: {error}')
        raise

def add_alert_rule(
    state: dict,
    rule_id: str,
    from_zone_id: Optional[str],
    to_zone_id: Optional[str],
    asset_id: Optional[str],
) -> dict:
    """
    Add an alert rule to state and return it.
    Raise ValueError if rule_id already exists.
    """
    try:
        target_rule = next((rule for rule in state["alert_rules"] if rule["id"] == rule_id), None)
        if target_rule:
            raise ValueError
        rule = {
            "id": rule_id,
            "from_zone_id": from_zone_id,
            "to_zone_id": to_zone_id,
            "asset_id": asset_id
        }
        state["alert_rules"].append(rule)
        return rule
    except ValueError as error:
        print(f'Rule already exists: {error}')
        raise
