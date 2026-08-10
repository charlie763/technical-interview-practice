/**
 * =============================================================================
 * INTERVIEW PROBLEM 1: Geofence Alert Rule Engine
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building a backend service for an IoT asset-tracking platform. Physical
 * assets (forklifts, shipping containers, field equipment) carry GPS sensors that
 * periodically report coordinates. The platform tracks which geographic "zone"
 * (geofence) each asset is currently inside, and fires configured alert rules
 * whenever an asset transitions between zones.
 *
 * For this problem, zones are axis-aligned bounding boxes — no geospatial
 * libraries needed.
 *
 * DATA MODEL
 * ----------
 * All state lives in a single TrackerState object (returned by makeTracker).
 * Store all state in the TrackerState object itself — there is no class here,
 * every function takes state in and mutates/reads it explicitly.
 *
 * # Example
 * const state = makeTracker();
 * addZone(state, "warehouse", "Warehouse A", 35.0, 35.1, -106.7, -106.6);
 * addAsset(state, "forklift_1", "Forklift #1");
 * addAlertRule(state, "rule_entry", undefined, "warehouse", undefined);
 * processLocationUpdate(state, "forklift_1", 35.05, -106.65, "t1");
 * // -> [{ ruleId: "rule_entry", assetId: "forklift_1", fromZoneId: undefined,
 * //       toZoneId: "warehouse", timestamp: "t1" }]
 * =============================================================================
 */

export type Zone = {
  id: string;
  name: string;
  bounds: {
    minLat: number;
    maxLat: number;
    minLng: number;
    maxLng: number;
  };
};

export type Asset = {
  id: string;
  name: string;
  lat?: number;
  lng?: number;
  zoneId?: string;
};

export type AlertRule = {
  id: string;
  fromZoneId?: string;
  toZoneId?: string;
  assetId?: string;
};

export type TriggeredAlert = {
  ruleId: string;
  assetId: string;
  fromZoneId?: string;
  toZoneId?: string;
  timestamp: string;
};

export type TrackerState = {
  zones: Record<string, Zone>;
  assets: Record<string, Asset>;
  alertRules: AlertRule[];
  alertLog: TriggeredAlert[];
};

/** Return a fresh, empty TrackerState. */
export function makeTracker(): TrackerState {
  return { zones: {}, assets: {}, alertRules: [], alertLog: [] };
}

// ---------------------------------------------------------------------------
// PART 1 — Zone membership  (warm-up, ~5 min)
// ---------------------------------------------------------------------------

/**
 * Return true if the asset's lat/lng falls inside the zone's bounding box.
 * - Bounds are inclusive on all edges.
 * - Return false if the asset has no location (lat or lng is undefined).
 */
export function isInZone(asset: Asset, zone: Zone): boolean {
  if (asset.lat === undefined || asset.lng === undefined) {
    return false;
  }
  const isInLat = zone.bounds.minLat <= asset.lat && asset.lat <= zone.bounds.maxLat;
  const isInLng = zone.bounds.minLng <= asset.lng && asset.lng <= zone.bounds.maxLng;
  return isInLat && isInLng;
}

// ---------------------------------------------------------------------------
// PART 2 — Locate an asset  (~5 min)
// ---------------------------------------------------------------------------

/**
 * Return the id of the first zone in state that contains the asset, or undefined.
 * - Return undefined if the asset doesn't exist or has no location.
 * - Iterate zones in insertion order (standard object key order).
 */
export function getCurrentZoneId(state: TrackerState, assetId: string): string | undefined {
  if (!(assetId in state.assets)) {
    return undefined;
  }
  const asset = state.assets[assetId];
  for (let zone of Object.values(state.zones)) {
    if (isInZone(asset, zone)) {
      return zone.id;
    }
  }
  return undefined;
}

// ---------------------------------------------------------------------------
// PART 3 — Process a location update  (core logic, ~15 min)
// ---------------------------------------------------------------------------

/**
 * Handle a new GPS reading for an asset:
 *
 * 1. Update the asset's lat and lng in state.
 * 2. Recompute zoneId via getCurrentZoneId and store it on the asset.
 * 3. If zoneId changed (including undefined→zone or zone→undefined), evaluate
 *    all alertRules and collect any that match.
 * 4. Append each matched rule as a TriggeredAlert to state.alertLog.
 * 5. Return the list of newly triggered alerts (empty array if no zone change
 *    or no rule matches).
 *
 * Throw an Error if assetId is not in state.assets.
 *
 * Alert-rule matching — a rule matches when ALL three conditions hold:
 *     rule.assetId     is undefined  OR  rule.assetId     === assetId
 *     rule.fromZoneId  is undefined  OR  rule.fromZoneId  === oldZoneId
 *     rule.toZoneId    is undefined  OR  rule.toZoneId    === newZoneId
 */
export function processLocationUpdate(state: TrackerState, assetId: string, lat: number, lng: number, timestamp: string): TriggeredAlert[] {
  if (!(assetId in state.assets)) {
    throw new Error("No assest found")
  }
  const asset = state.assets[assetId]
  const prevZone = getCurrentZoneId(state, asset.id)
  asset.lat = lat
  asset.lng = lng
  const currentZone = getCurrentZoneId(state, assetId)
  if (currentZone == prevZone) {
    return []
  }
  //throw new Error("processLocationUpdate not implemented");
  const rulesMatched = state.alertRules.filter((ar) => {
    let assetMatched = false
    let fromZoneMatched = false
    let toZoneMatched = false
    if (ar.assetId === undefined || ar.assetId === assetId) {
      assetMatched = true
    }
    if (ar.fromZoneId === undefined || ar.fromZoneId === prevZone) {
      fromZoneMatched = true
    }
    if (ar.toZoneId === undefined || ar.toZoneId === currentZone) {
      toZoneMatched = true
    }
    return assetMatched && fromZoneMatched && toZoneMatched
  })

  const tirggeredAlerts: TriggeredAlert[] = rulesMatched.map<TriggeredAlert>((rule) => ({
    ruleId: rule.id,
    assetId: assetId,
    fromZoneId: prevZone,
    toZoneId: currentZone,
    timestamp
  }))
  state.alertLog.push(...tirggeredAlerts)
  return tirggeredAlerts
}


// ---------------------------------------------------------------------------
// PART 4 — CRUD helpers  (~15 min)
// ---------------------------------------------------------------------------

/**
 * Create a zone, store it in state, and return it.
 * Throw an Error if zoneId already exists.
 */
export function addZone(state: TrackerState, zoneId: string, name: string, minLat: number, maxLat: number, minLng: number, maxLng: number): Zone {
  if (zoneId in state.zones) {
    throw new Error("Zoneid already exsists");
  }
  state.zones[zoneId] = {
    id: zoneId,
    name,
    bounds: {
      minLat,
      maxLat,
      minLng,
      maxLng,
    },
  };
  return state.zones[zoneId];
}

/**
 * Remove a zone from state. Throw an Error if not found.
 * Any asset currently assigned to the removed zone should have its
 * zoneId set to undefined. Do NOT fire alert rules for this forced change.
 */
export function removeZone(state: TrackerState, zoneId: string): void {
  if (!(zoneId in state.zones)) {
    throw new Error("zone id not foud")
  }
  const zone = state.zones[zoneId]
  for (let asset of Object.values(state.assets)) {
    if (asset.zoneId === zone.id) {
      asset.zoneId = undefined
    }
  }
  delete state.zones[zoneId]
}

/**
 * Create an asset with no initial location (lat=undefined, lng=undefined,
 * zoneId=undefined), store it in state, and return it.
 * Throw an Error if assetId already exists.
 */
export function addAsset(state: TrackerState, assetId: string, name: string): Asset {
  if (assetId in state.assets) {
    throw new Error("Asses id already exists");
  }
  const asset: Asset = {
    id: assetId,
    name,
  };
  state.assets[assetId] = asset
  return asset
}

/**
 * Add an alert rule to state and return it.
 * Throw an Error if ruleId already exists.
 */
export function addAlertRule(
  state: TrackerState,
  ruleId: string,
  fromZoneId: string | undefined,
  toZoneId: string | undefined,
  assetId: string | undefined,
): AlertRule {
  const rule = state.alertRules.find((ar) => ar.id === ruleId)
  if (rule !== undefined) {
    throw new Error('Rule already defined')
  }
  const alertRule: AlertRule = {
    id: ruleId,
    assetId,
    fromZoneId,
    toZoneId
  }
  state.alertRules.push(alertRule)
  return alertRule
}
