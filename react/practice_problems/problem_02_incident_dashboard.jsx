/**
 * =============================================================================
 * INTERVIEW PROBLEM 2: Real-Time Incident Dashboard
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the live incident monitoring panel for a public safety
 * operations platform. Incidents are streamed in real time from multiple field
 * sensors and radio dispatch sources. Operators need to see active incidents
 * sorted by recency, filter by severity/type, and mark incidents as contained.
 *
 * A working mock incident feed, sample data, and basic styles are provided.
 * Do NOT modify anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 *
 * PART 1 — Live incident feed  (~15 min)
 * ───────────────────────────────────────
 * Render an <IncidentDashboard /> component that:
 *   - Subscribes to `incidentFeed` on mount and unsubscribes on unmount.
 *   - Displays active (uncontained) incidents, newest first.
 *   - Each row shows: severity badge (color-coded), incident type, location,
 *     and a relative timestamp ("2 min ago").
 *
 * Data-testid requirements (Playwright relies on these):
 *   data-testid="incident-row"    — wrapper element for each active incident
 *   data-testid="severity-badge"  — the colored severity indicator in each row
 *
 * PART 2 — Filtering  (~15 min)
 * ──────────────────────────────
 * Add filter controls above the incident list:
 *   - A <select data-testid="filter-severity"> for severity
 *     (options: "all", "critical", "high", "medium", "low")
 *   - A <select data-testid="filter-type"> for incident type
 *     (options: "all" + each value in INCIDENT_TYPES)
 *
 * Requirements:
 *   - Filtering is purely client-side — all incidents stay in memory.
 *   - Changing a filter must NOT restart the feed subscription.
 *   - Use useMemo (or equivalent) so the filtered list is only recomputed
 *     when incidents or filter state actually change.
 *   - Show a count of currently visible incidents next to the filters.
 *
 * PART 3 — Optimistic "Contain" action  (~15 min)
 * ─────────────────────────────────────────────────
 * Add a "Contain" button to each active incident row.
 *
 * When clicked:
 *   1. Immediately mark the incident as contained in local state (optimistic).
 *   2. Call `containIncident(incidentId)` — resolves on success, rejects ~20%.
 *   3. On success: move the row to a "Contained" section at the bottom of the
 *      page with data-testid="contained-row".
 *   4. On rejection: revert the incident to active, show an inline error
 *      ("Failed — retry") on that row.
 *   5. While the request is in-flight, the button (data-testid="contain-btn")
 *      must be disabled to prevent double-submission.
 *
 * =============================================================================
 */

import React, { useState, useEffect, useMemo, useCallback } from 'react'

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

export const INCIDENT_TYPES = ['shooting', 'car-crash', 'fire', 'break-in', 'flooding']
export const SEVERITIES     = ['critical', 'high', 'medium', 'low']
export const LOCATIONS      = [
  'Downtown', 'Midtown', 'Eastside', 'Westside', 'South Bay',
  'Harbor District', 'North Hills', 'Riverfront',
]

export const SEVERITY_COLORS = {
  critical: { background: '#fef2f2', color: '#991b1b', border: '#fca5a5' },
  high:     { background: '#fff7ed', color: '#9a3412', border: '#fdba74' },
  medium:   { background: '#fefce8', color: '#854d0e', border: '#fde047' },
  low:      { background: '#f0fdf4', color: '#166534', border: '#86efac' },
}

/**
 * Deterministic seed incidents — appear immediately on connect.
 * Playwright tests rely on the location names and severities here.
 */
export const SEED_INCIDENTS = [
  { id: 'inc-s1', type: 'shooting',  severity: 'critical', location: 'Downtown',       ts: new Date(Date.now() - 2  * 60_000).toISOString(), contained: false },
  { id: 'inc-s2', type: 'car-crash', severity: 'high',     location: 'Midtown',        ts: new Date(Date.now() - 5  * 60_000).toISOString(), contained: false },
  { id: 'inc-s3', type: 'fire',      severity: 'critical', location: 'Eastside',       ts: new Date(Date.now() - 8  * 60_000).toISOString(), contained: false },
  { id: 'inc-s4', type: 'break-in',  severity: 'medium',   location: 'Westside',       ts: new Date(Date.now() - 12 * 60_000).toISOString(), contained: false },
  { id: 'inc-s5', type: 'flooding',  severity: 'low',      location: 'South Bay',      ts: new Date(Date.now() - 20 * 60_000).toISOString(), contained: false },
]

let _incidentId = 100

function makeRandomIncident() {
  const type     = INCIDENT_TYPES[Math.floor(Math.random() * INCIDENT_TYPES.length)]
  const severity = SEVERITIES[Math.floor(Math.random() * SEVERITIES.length)]
  const location = LOCATIONS[Math.floor(Math.random() * LOCATIONS.length)]
  return {
    id:        `inc-${_incidentId++}`,
    type,
    severity,
    location,
    ts:        new Date().toISOString(),
    contained: false,
  }
}

/**
 * createIncidentFeed() → { connect, disconnect, subscribe }
 *
 * connect()           — start emitting; seed incidents arrive within 300ms,
 *                       then a new random incident every 2s
 * disconnect()        — stop emitting
 * subscribe(handler)  — handler receives an incident object; returns unsubscribe fn
 */
export function createIncidentFeed() {
  let handlers = []
  let timerId  = null

  function emit(incident) {
    handlers.forEach(h => h({ ...incident }))
  }

  return {
    connect() {
      // Emit seed incidents quickly for deterministic Playwright tests
      SEED_INCIDENTS.forEach((inc, i) => setTimeout(() => emit(inc), i * 50))
      // Then keep emitting random incidents
      timerId = setInterval(() => emit(makeRandomIncident()), 2000)
    },
    disconnect() {
      clearInterval(timerId)
      timerId = null
    },
    subscribe(handler) {
      handlers.push(handler)
      return () => { handlers = handlers.filter(h => h !== handler) }
    },
  }
}

/**
 * containIncident(incidentId) → Promise<void>
 * Simulates a PATCH /incidents/:id API call.
 * Resolves after ~500ms. Rejects ~20% of the time.
 */
export function containIncident(incidentId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      Math.random() < 0.2
        ? reject(new Error(`Failed to contain ${incidentId}`))
        : resolve()
    }, 500)
  })
}

// Shared singleton — import this in your component
export const incidentFeed = createIncidentFeed()

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

/**
 * SEVERITY_COLORS and INCIDENT_TYPES are available above for styling.
 * Feel free to add helpers or sub-components as needed.
 */

// ---------------------------------------------------------------------------
// PART 1 — Implement IncidentDashboard
// ---------------------------------------------------------------------------

export function IncidentDashboard() {
  // TODO Part 1: subscribe to incidentFeed on mount, unsubscribe on unmount.
  //              Store incidents in state; render active ones newest-first.
  //              Each row: data-testid="incident-row"
  //              Severity badge: data-testid="severity-badge"

  // TODO Part 2: add severity + type filter state with useMemo.
  //              Render <select data-testid="filter-severity"> and
  //              <select data-testid="filter-type">. No re-subscribe on filter change.

  // TODO Part 3: optimistic "Contain" action.
  //              Button: data-testid="contain-btn" (disabled while in-flight)
  //              Contained rows: data-testid="contained-row"

  return (
    <div style={{ fontFamily: 'system-ui', maxWidth: 800, margin: '0 auto', padding: 24 }}>
      <h2 style={{ marginBottom: 16 }}>Incident Dashboard</h2>

      {/* TODO: filter controls */}

      {/* TODO: active incident list */}
      <p style={{ color: '#94a3b8' }}>No incidents yet.</p>

      {/* TODO: contained section */}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Suggested sub-components (optional — structure however you like)
// ---------------------------------------------------------------------------

// function FilterBar({ severityFilter, typeFilter, onSeverityChange, onTypeChange, count }) { ... }
// function IncidentRow({ incident, onContain }) { ... }

// ---------------------------------------------------------------------------
// App entry point
// ---------------------------------------------------------------------------

export default function App() {
  useEffect(() => {
    incidentFeed.connect()
    return () => incidentFeed.disconnect()
  }, [])

  return <IncidentDashboard />
}
