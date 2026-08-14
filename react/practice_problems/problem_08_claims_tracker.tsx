/**
 * =============================================================================
 * INTERVIEW PROBLEM 08: Claims Tracker
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the claims tracker for a management liability insurance
 * platform. Adjusters monitor open claims, watch for status changes polled
 * from the server, and can leave notes directly in the UI.
 *
 * A mock polling API, seed data, and event helpers are provided.
 * Do NOT modify anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 * PROVIDED (do not modify)
 * =============================================================================
 *
 * SEED_CLAIMS        — 5 deterministic claims used in Playwright tests.
 * fetchClaims()      → Promise<Claim[]>
 *                    — mock REST GET /claims; resolves after ~400 ms.
 *                      Returns the current in-memory claim list, which may
 *                      include status changes applied by the mock.
 * addClaimNote(claimId, note) → Promise<ClaimNote>
 *                    — mock POST /claims/:id/notes; rejects ~15% of the time.
 *
 * =============================================================================
 * PART 1 — Claims list  (~15 min)
 * =============================================================================
 *
 * Render a <ClaimsTracker /> component that:
 *   - Initialises state with SEED_CLAIMS on mount.
 *   - Displays each claim in a row showing:
 *       Claim ID, coverage type badge, status badge, claimed amount
 *       (formatted as "$75,000"), policy ID, and filed date.
 *   - Shows a summary header with:
 *       total claim count ("5 claims") and
 *       total claimed amount ("$315,000 total claimed").
 *
 * Required data-testid attributes:
 *   data-testid="claim-row"   — wrapper element for each claim row
 *
 * TYPE CONTRACT
 * -------------
 * type ClaimStatus  = 'filed' | 'investigating' | 'evaluation' | 'settled' | 'denied'
 * type CoverageType = 'epl' | 'do' | 'fiduciary'
 *
 * interface Claim {
 *   id:             string
 *   policy_id:      string
 *   coverage_type:  CoverageType
 *   filed_date:     string        // ISO-8601 date
 *   status:         ClaimStatus
 *   claimed_amount: number        // dollars
 *   notes:          ClaimNote[]
 * }
 *
 * interface ClaimNote {
 *   id:         string
 *   author:     string
 *   text:       string
 *   created_at: string  // ISO-8601 datetime
 * }
 *
 * =============================================================================
 * PART 2 — Polling and status filter  (~15 min)
 * =============================================================================
 *
 * Add status polling:
 *   - Poll fetchClaims() every 5 seconds using setInterval inside useEffect.
 *   - On each successful poll, merge the result with current state:
 *       update status and notes for existing claims, keep any local state
 *       (e.g. open detail panel) intact.
 *   - Cancel the interval on unmount (cleanup function in useEffect).
 *   - Use useRef to hold the interval ID so it doesn't trigger re-renders.
 *
 * Add a status filter:
 *   - A row of clickable tab buttons, one per status plus "All".
 *   - Tabs: "All" | "filed" | "investigating" | "evaluation" | "settled" | "denied"
 *   - The active tab is visually distinct (border, colour, etc.).
 *   - Filtering must NOT restart the polling interval.
 *   - Derive the visible list with useMemo.
 *   - Show "Showing N claims".
 *
 * Required data-testid attributes:
 *   data-testid="tab-all"           — "All" tab button
 *   data-testid="tab-investigating" — "investigating" tab button
 *   data-testid="tab-settled"       — "settled" tab button
 *
 * =============================================================================
 * PART 3 — Claim detail panel with note submission  (~15 min)
 * =============================================================================
 *
 * Make claim rows expandable:
 *   - Clicking a claim row (or a dedicated "View" button) toggles an inline
 *     detail panel below that row (accordion style — only one open at a time).
 *   - The panel shows:
 *       Policy ID, filed date, full claim notes list (author + text + date).
 *       If no notes: "No notes yet."
 *   - At the bottom of the panel: a <textarea> for a new note and an
 *     "Add Note" button.
 *   - Clicking "Add Note" calls addClaimNote(claimId, noteText).
 *   - Show a spinner (data-testid="note-spinner") while the request is in-flight.
 *   - On success: prepend the returned ClaimNote to the claim's notes list
 *     and clear the textarea.
 *   - On failure: show an inline error ("Failed to save note — try again.")
 *     and keep the textarea text.
 *   - Disable the textarea and button while in-flight.
 *
 * Required data-testid attributes:
 *   data-testid="claim-detail"   — the expanded detail panel
 *   data-testid="note-spinner"   — loading indicator during note save
 *   data-testid="note-textarea"  — the textarea for new notes
 *
 * =============================================================================
 *
 * TEST CONTRACT
 * -------------
 * Playwright tests use the following query strategy (in priority order):
 *
 * 1. Seed data text — toContainText('CLM-2001'), toContainText('75,000') etc.
 * 2. Element type + visible text — locator('button').filter({ hasText: /add note/i })
 * 3. data-testid — required for: claim rows, tabs, detail panel, note spinner,
 *    note textarea (all listed above).
 *
 * =============================================================================
 */

import React, { useState, useEffect, useMemo, useRef } from 'react'

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

export type ClaimStatus  = 'filed' | 'investigating' | 'evaluation' | 'settled' | 'denied'
export type CoverageType = 'epl' | 'do' | 'fiduciary'

export interface ClaimNote {
  id:         string
  author:     string
  text:       string
  created_at: string
}

export interface Claim {
  id:             string
  policy_id:      string
  coverage_type:  CoverageType
  filed_date:     string
  status:         ClaimStatus
  claimed_amount: number
  notes:          ClaimNote[]
}

export const SEED_CLAIMS: Claim[] = [
  {
    id:             'CLM-2001',
    policy_id:      'POL-101',
    coverage_type:  'epl',
    filed_date:     '2025-03-10',
    status:         'investigating',
    claimed_amount: 75_000,
    notes: [
      {
        id:         'note-s1',
        author:     'adj.morgan',
        text:       'Initial review complete. Requesting HR records.',
        created_at: '2025-03-12T10:30:00',
      },
    ],
  },
  {
    id:             'CLM-2002',
    policy_id:      'POL-102',
    coverage_type:  'do',
    filed_date:     '2025-03-15',
    status:         'evaluation',
    claimed_amount: 150_000,
    notes: [],
  },
  {
    id:             'CLM-2003',
    policy_id:      'POL-101',
    coverage_type:  'epl',
    filed_date:     '2025-03-18',
    status:         'filed',
    claimed_amount: 30_000,
    notes: [],
  },
  {
    id:             'CLM-2004',
    policy_id:      'POL-103',
    coverage_type:  'fiduciary',
    filed_date:     '2025-02-20',
    status:         'settled',
    claimed_amount: 40_000,
    notes: [
      {
        id:         'note-s2',
        author:     'adj.chen',
        text:       'Settled for $32,000. Closing file.',
        created_at: '2025-04-01T14:00:00',
      },
    ],
  },
  {
    id:             'CLM-2005',
    policy_id:      'POL-104',
    coverage_type:  'do',
    filed_date:     '2025-04-02',
    status:         'denied',
    claimed_amount: 20_000,
    notes: [],
  },
]

// Internal mutable state so polls can surface simulated status changes.
let _claims: Claim[] = SEED_CLAIMS.map(c => ({ ...c, notes: [...c.notes] }))
let _noteId = 200

/**
 * fetchClaims() → Promise<Claim[]>
 * Simulates GET /claims. Resolves after ~400 ms.
 * On the 3rd+ call, advances CLM-2003 from "filed" → "investigating".
 */
let _fetchCount = 0
export function fetchClaims(): Promise<Claim[]> {
  return new Promise(resolve => {
    setTimeout(() => {
      _fetchCount++
      if (_fetchCount >= 3) {
        _claims = _claims.map(c =>
          c.id === 'CLM-2003' && c.status === 'filed'
            ? { ...c, status: 'investigating' }
            : c
        )
      }
      resolve(_claims.map(c => ({ ...c, notes: [...c.notes] })))
    }, 400)
  })
}

/**
 * addClaimNote(claimId, note) → Promise<ClaimNote>
 * Simulates POST /claims/:id/notes. Resolves after ~350 ms; rejects ~15%.
 */
export function addClaimNote(claimId: string, note: string): Promise<ClaimNote> {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.15) {
        reject(new Error(`Failed to save note for ${claimId}`))
      } else {
        const newNote: ClaimNote = {
          id:         `note-${_noteId++}`,
          author:     'current.user',
          text:       note,
          created_at: new Date().toISOString(),
        }
        _claims = _claims.map(c =>
          c.id === claimId ? { ...c, notes: [newNote, ...c.notes] } : c
        )
        resolve(newNote)
      }
    }, 350)
  })
}

export const COVERAGE_COLORS: Record<CoverageType, string> = {
  epl:       '#3b82f6',
  do:        '#8b5cf6',
  fiduciary: '#f59e0b',
}

export const STATUS_COLORS: Record<ClaimStatus, string> = {
  filed:         '#94a3b8',
  investigating: '#f59e0b',
  evaluation:    '#3b82f6',
  settled:       '#22c55e',
  denied:        '#ef4444',
}

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

// Suggested status tabs (feel free to use or ignore)
const ALL_STATUSES: ClaimStatus[] = ['filed', 'investigating', 'evaluation', 'settled', 'denied']

// ---------------------------------------------------------------------------
// PART 1 — Implement ClaimsTracker
// ---------------------------------------------------------------------------

export function ClaimsTracker() {
  // TODO Part 1: initialise state with SEED_CLAIMS.
  //              Render a summary header with total count and total claimed.
  //              Render each claim row: claim ID, coverage badge, status badge,
  //              claimed amount, policy ID, filed date.

  // TODO Part 2: poll fetchClaims() every 5 seconds; merge updates into state.
  //              Store interval ID in useRef so it doesn't trigger re-renders.
  //              Add status filter tabs (All + each status).
  //              Derive visible list with useMemo. Show "Showing N claims".

  // TODO Part 3: clicking a row toggles an inline detail panel (accordion, one open at a time).
  //              Panel shows policy ID, filed date, notes list.
  //              Add-note form: textarea + "Add Note" button.
  //              Call addClaimNote(); optimistically prepend returned note; revert on failure.

  return (
    <div style={{ fontFamily: 'sans-serif', maxWidth: 860, margin: '0 auto', padding: 24 }}>
      <h2 style={{ marginBottom: 16 }}>Claims Tracker</h2>

      {/* TODO: summary header */}

      {/* TODO: status filter tabs */}

      {/* TODO: claims list */}
      <p style={{ color: '#888' }}>No claims loaded.</p>
    </div>
  )
}

// ---------------------------------------------------------------------------
// App entry point
// ---------------------------------------------------------------------------

export default function App() {
  return <ClaimsTracker />
}
