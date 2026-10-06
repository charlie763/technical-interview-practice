# =============================================================================
# INTERVIEW PROBLEM 1: Geofence Alert Rule Engine
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# You're building a backend service for an IoT asset-tracking platform.
# Physical assets (forklifts, shipping containers, field equipment) carry GPS
# sensors that periodically report coordinates. The platform tracks which
# geographic "zone" (geofence) each asset is currently inside, and fires
# configured alert rules whenever an asset transitions between zones.
#
# For this problem, zones are axis-aligned bounding boxes — no geospatial
# libraries needed.
#
# You choose the internal data structures — the public interface below is
# what matters. Store all state in instance variables initialized in
# `initialize`. Class-level variables will bleed between tests and between
# instances — avoid them.
#
# DATA MODEL
# ----------
# Zone:
#   { id:, name:, bounds: { min_lat:, max_lat:, min_lng:, max_lng: } }
#
# Asset:
#   { id:, name:, lat: Float | nil, lng: Float | nil, zone_id: String | nil }
#   - lat/lng are nil until the first GPS update arrives.
#   - zone_id is the id of the zone the asset is currently in, or nil.
#
# AlertRule:
#   { id:, from_zone_id: String | nil, to_zone_id: String | nil, asset_id: String | nil }
#   - nil in from_zone_id matches ANY previous zone (including nil/"no zone").
#   - nil in to_zone_id   matches ANY new zone (including nil/"no zone").
#   - nil in asset_id     matches ANY asset.
#
# TriggeredAlert (appended to the alert log when a rule fires):
#   { rule_id:, asset_id:, from_zone_id:, to_zone_id:, timestamp: }
# =============================================================================

class GeofenceTracker
  def initialize
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Zone membership (warm-up, ~5 min)
  # ---------------------------------------------------------------------------

  # Return true if the asset's lat/lng falls inside the zone's bounding box.
  # - Bounds are inclusive on all edges.
  # - Return false if the asset has no location (lat or lng is nil).
  #
  # Stateless helper — doesn't touch instance state.
  def in_zone?(asset, zone)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Locate an asset (~5 min)
  # ---------------------------------------------------------------------------

  # Return the id of the first zone that contains the asset, or nil.
  # - Return nil if the asset doesn't exist or has no location.
  # - Iterate zones in insertion order.
  def current_zone_id(asset_id)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Process a location update (core logic, ~15 min)
  # ---------------------------------------------------------------------------

  # Handle a new GPS reading for an asset:
  #
  # 1. Update the asset's lat and lng.
  # 2. Recompute zone_id via current_zone_id and store it on the asset.
  # 3. If zone_id changed (including nil->zone or zone->nil), evaluate all
  #    alert rules and collect any that match.
  # 4. Append each matched rule as a TriggeredAlert to the alert log.
  # 5. Return the array of newly triggered alerts (empty array if no zone
  #    change or no rule matches).
  #
  # Raise KeyError if asset_id is not registered.
  #
  # Alert-rule matching — a rule matches when ALL three conditions hold:
  #   rule[:asset_id]     is nil  OR  rule[:asset_id]     == asset_id
  #   rule[:from_zone_id] is nil  OR  rule[:from_zone_id] == old_zone_id
  #   rule[:to_zone_id]   is nil  OR  rule[:to_zone_id]   == new_zone_id
  def process_location_update(asset_id, lat, lng, timestamp)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 4 — CRUD helpers (~15 min)
  # ---------------------------------------------------------------------------

  # Create a zone and return it. Raise ArgumentError if zone_id already exists.
  def add_zone(zone_id, name, min_lat, max_lat, min_lng, max_lng)
    raise NotImplementedError
  end

  # Remove a zone. Raise KeyError if not found.
  # Any asset currently assigned to the removed zone should have its zone_id
  # set to nil. Do NOT fire alert rules for this forced change.
  def remove_zone(zone_id)
    raise NotImplementedError
  end

  # Create an asset with no initial location (lat/lng/zone_id all nil) and
  # return it. Raise ArgumentError if asset_id already exists.
  def add_asset(asset_id, name)
    raise NotImplementedError
  end

  # Add an alert rule and return it. Raise ArgumentError if rule_id already exists.
  def add_alert_rule(rule_id, from_zone_id, to_zone_id, asset_id)
    raise NotImplementedError
  end
end
