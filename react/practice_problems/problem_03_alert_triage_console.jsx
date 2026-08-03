/**
 * =============================================================================
 * INTERVIEW PROBLEM 3: Alert Triage Console
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the dispatch operations console for an emergency-response
 * platform. Incoming alerts from radio dispatch and field sensors stream in
 * continuously. Operators triage them by assigning to available field units
 * or acknowledging alerts that don't require dispatch.
 *
 * A working mock alert stream, available units, and API stubs are provided.
 * Do NOT modify anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 *
 * PART 1 — Alert queue  (~15 min)
 * ────────────────────────────────
 * Render an <AlertTriageConsole /> component that:
 *   - Subscribes to `alertStream` on mount and unsubscribes on unmount.
 *   - Displays unacknowledged, unassigned alerts sorted by severity descending
 *     (5 = highest), then by timestamp ascending (oldest first within same sev).
 *   - Each row shows: severity badge (1–5, color-coded), alert type, message,
 *     source, and relative time ("3 min ago").
 *
 * Data-testid requirements (Playwright relies on these):
 *   data-testid="alert-row"      — wrapper element for each alert row
 *   data-testid="severity-badge" — the severity indicator inside each row
 *
 * PART 2 — Assign to unit  (~15 min)
 * ────────────────────────────────────
 * Add an assign action to each unassigned, unacknowledged alert row:
 *   - A <select data-testid="unit-select"> populated with UNITS.
 *   - A <button data-testid="assign-btn"> that calls assignAlert(alertId, unitId).
 *
 * When the assignment succeeds:
 *   - Show the assigned unit name on the row (no separate section needed yet).
 *   - Mark the alert as assigned in local state.
 * On failure (~15% of the time):
 *   - Revert to unassigned; show an inline error ("Assignment failed — retry").
 * While in-flight:
 *   - Disable the assign button and the unit select.
 *
 * PART 3 — Tab navigation and bulk-ack  (~15 min)
 * ─────────────────────────────────────────────────
 * Add a tab bar at the top of the console with three tabs:
 *   Queued | Assigned | Acknowledged
 *
 * Data-testid requirements:
 *   data-testid="tab-queued"        — Queued tab button
 *   data-testid="tab-assigned"      — Assigned tab button
 *   data-testid="tab-acknowledged"  — Acknowledged tab button
 *
 * Each tab shows its count in parentheses, e.g. "Queued (3)".
 *
 * Switching tabs changes which alerts are shown:
 *   Queued       — unassigned AND not acknowledged
 *   Assigned     — assigned AND not acknowledged
 *   Acknowledged — acknowledged (by any means)
 *
 * On the Queued tab, add an "Ack All" button (data-testid="ack-all-btn") that
 * calls acknowledgeAlert(id) for every currently visible queued alert and marks
 * them all as acknowledged. Fire all calls in parallel (Promise.all or similar).
 *
 * =============================================================================
 */

import React, { useState, useEffect, useMemo, useCallback } from 'react'

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

export const ALERT_TYPES = ['radio-dispatch', 'sensor-alert', 'social-monitor', 'field-report']

export const SEVERITY_COLORS = {
  5: { background: '#fef2f2', color: '#991b1b', border: '#fca5a5', label: 'P1' },
  4: { background: '#fff7ed', color: '#9a3412', border: '#fdba74', label: 'P2' },
  3: { background: '#fefce8', color: '#854d0e', border: '#fde047', label: 'P3' },
  2: { background: '#eff6ff', color: '#1e40af', border: '#93c5fd', label: 'P4' },
  1: { background: '#f0fdf4', color: '#166534', border: '#86efac', label: 'P5' },
}

/**
 * Available field units that alerts can be dispatched to.
 */
export const UNITS = [
  { id: 'unit-12', name: 'Alpha Team',  type: 'patrol' },
  { id: 'unit-14', name: 'Beta Team',   type: 'patrol' },
  { id: 'unit-21', name: 'Fire Unit 3', type: 'fire'   },
  { id: 'unit-33', name: 'Medic 5',     type: 'ems'    },
  { id: 'unit-45', name: 'SWAT Alpha',  type: 'swat'   },
]

/**
 * Deterministic seed alerts — appear immediately on connect.
 * Playwright tests rely on the message text and severity values here.
 */
export const SEED_ALERTS = [
  {
    id:           'alt-s1',
    type:         'radio-dispatch',
    severity:     5,
    message:      'Reports of shots fired at Central Station',
    source:       'radio-north',
    ts:           new Date(Date.now() - 1  * 60_000).toISOString(),
    assignedTo:   null,
    acknowledged: false,
  },
  {
    id:           'alt-s2',
    type:         'sensor-alert',
    severity:     4,
    message:      'Collision detected on Highway 12',
    source:       'sensor-grid-east',
    ts:           new Date(Date.now() - 3  * 60_000).toISOString(),
    assignedTo:   null,
    acknowledged: false,
  },
  {
    id:           'alt-s3',
    type:         'radio-dispatch',
    severity:     3,
    message:      'Structural fire reported at Warehouse District',
    source:       'radio-south',
    ts:           new Date(Date.now() - 6  * 60_000).toISOString(),
    assignedTo:   null,
    acknowledged: false,
  },
  {
    id:           'alt-s4',
    type:         'sensor-alert',
    severity:     2,
    message:      'Flooding sensors triggered at Riverside Park',
    source:       'sensor-grid-west',
    ts:           new Date(Date.now() - 10 * 60_000).toISOString(),
    assignedTo:   null,
    acknowledged: false,
  },
]

let _alertId = 100

const MESSAGES = {
  'radio-dispatch': [
    'Officer requesting backup at intersection of 5th and Main',
    'Suspicious vehicle reported near school zone',
    'Domestic disturbance call in progress',
  ],
  'sensor-alert': [
    'Gunshot detection triggered — confidence 87%',
    'Traffic sensor anomaly on Bridge Road',
    'Environmental sensor spike at industrial zone',
  ],
  'social-monitor': [
    'Spike in 911-related social posts — downtown area',
    'Crowd incident reported via social media',
  ],
  'field-report': [
    'Officer on scene — requesting additional units',
    'Situation escalating — requesting supervisor',
  ],
}

function makeRandomAlert() {
  const type     = ALERT_TYPES[Math.floor(Math.random() * ALERT_TYPES.length)]
  const msgs     = MESSAGES[type]
  const severity = Math.ceil(Math.random() * 5)
  return {
    id:           `alt-${_alertId++}`,
    type,
    severity,
    message:      msgs[Math.floor(Math.random() * msgs.length)],
    source:       `source-${Math.floor(Math.random() * 10)}`,
    ts:           new Date().toISOString(),
    assignedTo:   null,
    acknowledged: false,
  }
}

/**
 * createAlertStream() → { connect, disconnect, subscribe }
 *
 * connect()           — seed alerts arrive within 300ms, then new random
 *                       alerts every 2.5s
 * disconnect()        — stop emitting
 * subscribe(handler)  — returns unsubscribe fn
 */
export function createAlertStream() {
  let handlers = []
  let timerId  = null

  function emit(alert) {
    handlers.forEach(h => h({ ...alert }))
  }

  return {
    connect() {
      SEED_ALERTS.forEach((alert, i) => setTimeout(() => emit(alert), i * 50))
      timerId = setInterval(() => emit(makeRandomAlert()), 2500)
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
 * assignAlert(alertId, unitId) → Promise<void>
 * Simulates a POST /alerts/:id/assign API call.
 * Resolves after ~400ms. Rejects ~15% of the time.
 */
export function assignAlert(alertId, unitId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      Math.random() < 0.15
        ? reject(new Error(`Failed to assign ${alertId} to ${unitId}`))
        : resolve()
    }, 400)
  })
}

/**
 * acknowledgeAlert(alertId) → Promise<void>
 * Always resolves after ~300ms.
 */
export function acknowledgeAlert(alertId) {
  return new Promise(resolve => setTimeout(resolve, 300))
}

// Shared singleton — import this in your component
export const alertStream = createAlertStream()

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

// ---------------------------------------------------------------------------
// PART 1 — Implement AlertTriageConsole
// ---------------------------------------------------------------------------

export function AlertTriageConsole() {
  // TODO Part 1: subscribe to alertStream on mount, unsubscribe on unmount.
  //              Store alerts in state. On the default view, display queued
  //              alerts sorted by severity desc, then ts asc.
  //              Each row: data-testid="alert-row"
  //              Severity badge: data-testid="severity-badge"

  // TODO Part 2: add assign action per queued row.
  //              Unit select: data-testid="unit-select"
  //              Assign button: data-testid="assign-btn"
  //              Disable both while request is in-flight.

  // TODO Part 3: add tab bar and Ack All button.
  //              Tabs: data-testid="tab-queued", "tab-assigned", "tab-acknowledged"
  //              Ack All: data-testid="ack-all-btn" (visible on Queued tab only)

  return (
    <div style={{ fontFamily: 'system-ui', maxWidth: 900, margin: '0 auto', padding: 24 }}>
      <h2 style={{ marginBottom: 16 }}>Alert Triage Console</h2>

      {/* TODO: tab bar (Part 3) */}

      {/* TODO: Ack All button (Part 3, Queued tab only) */}

      {/* TODO: alert list (Part 1) */}
      <p style={{ color: '#94a3b8' }}>No alerts yet.</p>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Suggested sub-components (optional — structure however you like)
// ---------------------------------------------------------------------------

// function TabBar({ activeTab, counts, onTabChange }) { ... }
// function AlertRow({ alert, units, onAssign, onAcknowledge }) { ... }

// ---------------------------------------------------------------------------
// App entry point
// ---------------------------------------------------------------------------

export default function App() {
  useEffect(() => {
    alertStream.connect()
    return () => alertStream.disconnect()
  }, [])

  return <AlertTriageConsole />
}
