/**
 * =============================================================================
 * INTERVIEW PROBLEM 05: Campaign Fundraising Dashboard
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the live campaign dashboard for a K-12 school fundraising
 * platform. Development and advancement teams use this view to monitor a
 * running annual-fund campaign in real time: how close they are to their goal,
 * who has given, and whether thank-you messages have been dispatched.
 *
 * A mock donation stream, campaign data, and API stubs are provided.
 * Do NOT modify anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 * PROVIDED (do not modify)
 * =============================================================================
 *
 * CAMPAIGN          — campaign metadata (name, goal, dates).
 * SEED_DONATIONS    — 6 deterministic donations used in Playwright tests.
 * createDonationStream() → { connect, disconnect, subscribe }
 *                   — mock real-time stream; emits random donations every ~2 s.
 * sendThankYou(donationId) → Promise<void>
 *                   — mock API call; rejects ~15% of the time.
 *
 * =============================================================================
 * PART 1 — Campaign progress + live donation feed  (~15 min)
 * =============================================================================
 *
 * Render a <CampaignDashboard /> component that:
 *   - Initialises state with SEED_DONATIONS (newest first based on timestamp).
 *   - Subscribes to `donationStream` on mount; unsubscribes on unmount.
 *   - Displays the campaign name, a progress bar showing % of goal reached,
 *     total raised (e.g. "$9,400"), and donor count ("6 donors").
 *   - Displays each donation in reverse-chronological order (newest on top).
 *   - Each row shows: donor name, formatted amount, donor type badge, and
 *     relative time ("5 min ago").
 *
 * Required data-testid attributes:
 *   data-testid="donation-row"  — wrapper element for each donation row
 *   data-testid="progress-bar" — the filled portion of the goal progress bar
 *
 * =============================================================================
 * PART 2 — Filter and search  (~15 min)
 * =============================================================================
 *
 * Add two controls above the donation feed:
 *
 * 1. Donor-type filter — a <select> that filters by donor type.
 *    Options: "All", "family", "alumni", "staff", "board"
 *    data-testid="filter-type"
 *
 * 2. Name search — an <input> that filters by donor name (case-insensitive
 *    substring match). Tests accept data-testid="search-input" OR a
 *    placeholder attribute containing "search" (case-insensitive).
 *
 * Requirements:
 *   - Changing a filter must NOT restart the donation stream subscription.
 *   - Derive the visible list with useMemo.
 *   - Show "Showing N donations" reflecting the filtered count.
 *
 * =============================================================================
 * PART 3 — Optimistic "Send Thanks" action  (~15 min)
 * =============================================================================
 *
 * Add a "Send Thanks" button to each donation row where thanksSent === false.
 *
 * When clicked:
 *   1. Immediately mark the donation as thanksSent = true (optimistic update).
 *   2. Call sendThankYou(donationId) and show a spinner
 *      (data-testid="thanks-spinner") while the request is in-flight.
 *   3. Disable the button while in-flight to prevent double-submission.
 *   4. On success: keep the sent state (show "Thanks sent" or similar).
 *   5. On failure: revert thanksSent to false and show an inline error
 *      ("Failed — try again").
 *
 * Required data-testid attributes:
 *   data-testid="thanks-spinner"  — loading indicator during send
 *
 * =============================================================================
 *
 * TYPE CONTRACT
 * -------------
 * The following types are exported from the PROVIDED section below.
 * Do not redefine them.
 *
 * interface Campaign   { id, name, goalAmount, startDate, endDate }
 * type DonorType       = 'family' | 'alumni' | 'staff' | 'board'
 * interface Donation   { id, donorName, amount, donorType, message,
 *                        timestamp, thanksSent }
 *
 * =============================================================================
 *
 * TEST CONTRACT
 * -------------
 * Playwright tests use the following query strategy (in priority order):
 *
 * 1. Seed data text — toContainText('Margaret Chen'), toContainText('9,400') etc.
 * 2. Element type + visible text — locator('button').filter({ hasText: /send thanks/i })
 * 3. Input attributes — getByPlaceholder(/search/i) for the name search input
 * 4. data-testid — required for: donation rows (counting/ordering), progress bar,
 *    filter-type select, and thanks-spinner (all listed above).
 *
 * =============================================================================
 */

import React, { useState, useEffect, useMemo } from 'react'

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

export interface Campaign {
  id: string
  name: string
  goalAmount: number
  startDate: string
  endDate: string
}

export type DonorType = 'family' | 'alumni' | 'staff' | 'board'

export interface Donation {
  id: string
  donorName: string
  amount: number
  donorType: DonorType
  message: string
  timestamp: string
  thanksSent: boolean
}

export const CAMPAIGN: Campaign = {
  id: 'camp-2026',
  name: 'Westfield Academy Annual Fund 2026',
  goalAmount: 50_000,
  startDate: '2026-08-01',
  endDate: '2026-09-30',
}

/**
 * Deterministic seed donations — appear immediately on load.
 * Playwright tests rely on donor names, amounts, and types here.
 * Total raised from seed data: $9,400 across 6 donors.
 */
export const SEED_DONATIONS: Donation[] = [
  {
    id: 'don-s1',
    donorName: 'Margaret Chen',
    amount: 1_000,
    donorType: 'alumni',
    message: 'Go Eagles!',
    timestamp: new Date(Date.now() - 5 * 60_000).toISOString(),
    thanksSent: false,
  },
  {
    id: 'don-s2',
    donorName: 'The Patel Family',
    amount: 500,
    donorType: 'family',
    message: '',
    timestamp: new Date(Date.now() - 12 * 60_000).toISOString(),
    thanksSent: false,
  },
  {
    id: 'don-s3',
    donorName: 'Robert Kim',
    amount: 250,
    donorType: 'staff',
    message: 'Proud to support our students.',
    timestamp: new Date(Date.now() - 20 * 60_000).toISOString(),
    thanksSent: false,
  },
  {
    id: 'don-s4',
    donorName: 'Westfield School Board',
    amount: 5_000,
    donorType: 'board',
    message: '',
    timestamp: new Date(Date.now() - 35 * 60_000).toISOString(),
    thanksSent: true,
  },
  {
    id: 'don-s5',
    donorName: 'Sarah Williams',
    amount: 150,
    donorType: 'family',
    message: 'Happy to help!',
    timestamp: new Date(Date.now() - 48 * 60_000).toISOString(),
    thanksSent: false,
  },
  {
    id: 'don-s6',
    donorName: 'James Park',
    amount: 2_500,
    donorType: 'alumni',
    message: 'Best school ever.',
    timestamp: new Date(Date.now() - 60 * 60_000).toISOString(),
    thanksSent: false,
  },
]

const RANDOM_NAMES = [
  'Alex Foster', 'Linda Zhao', 'Omar Hassan', 'Julia Reeves',
  'Chris Navarro', 'Anne Murphy', 'Sam Okafor', 'Diane Cheng',
]
const RANDOM_MESSAGES = [
  'Happy to give!', 'Keep up the great work.', '', 'Go team!', '',
]
const DONOR_TYPES: DonorType[] = ['family', 'alumni', 'staff', 'board']
let _donId = 100

/**
 * createDonationStream() → { connect, disconnect, subscribe }
 *
 * connect()          — start emitting random donations (~every 2 s)
 * disconnect()       — stop emitting
 * subscribe(handler) — register a handler; returns an unsubscribe function.
 *                      handler receives a Donation object (thanksSent: false).
 */
export function createDonationStream(): {
  connect: () => void
  disconnect: () => void
  subscribe: (handler: (donation: Donation) => void) => () => void
} {
  let handlers: Array<(d: Donation) => void> = []
  let timerId: ReturnType<typeof setInterval> | null = null

  function emit() {
    const donation: Donation = {
      id: `don-${_donId++}`,
      donorName: RANDOM_NAMES[Math.floor(Math.random() * RANDOM_NAMES.length)],
      amount: [25, 50, 100, 250, 500, 1_000][Math.floor(Math.random() * 6)],
      donorType: DONOR_TYPES[Math.floor(Math.random() * DONOR_TYPES.length)],
      message: RANDOM_MESSAGES[Math.floor(Math.random() * RANDOM_MESSAGES.length)],
      timestamp: new Date().toISOString(),
      thanksSent: false,
    }
    handlers.forEach(h => h(donation))
  }

  return {
    connect: () => { timerId = setInterval(emit, 2_000) },
    disconnect: () => { if (timerId !== null) { clearInterval(timerId); timerId = null } },
    subscribe: (handler) => {
      handlers.push(handler)
      return () => { handlers = handlers.filter(h => h !== handler) }
    },
  }
}

/**
 * sendThankYou(donationId) → Promise<void>
 * Simulates a POST /donations/:id/thank-you API call.
 * Resolves after ~350 ms. Rejects ~15% of the time.
 */
export function sendThankYou(donationId: string): Promise<void> {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.15) {
        reject(new Error(`Failed to send thank-you for ${donationId}`))
      } else {
        resolve()
      }
    }, 350)
  })
}

// Shared stream instance — subscribe to this in your component
export const donationStream = createDonationStream()

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

// Suggested donor-type colour helpers (feel free to use or ignore these)
const DONOR_TYPE_COLORS: Record<DonorType, string> = {
  family:  '#3b82f6',
  alumni:  '#8b5cf6',
  staff:   '#22c55e',
  board:   '#f59e0b',
}

// ---------------------------------------------------------------------------
// PART 1 — Implement CampaignDashboard
// ---------------------------------------------------------------------------

export function CampaignDashboard() {
  // TODO Part 1: initialise state with SEED_DONATIONS sorted newest-first.
  //              Subscribe to donationStream on mount; unsubscribe on unmount.
  //              Prepend incoming donations so new ones appear at the top.
  //              Render the campaign name, goal progress bar, total raised,
  //              and donor count.
  //              Render each donation row: donor name, formatted amount ($X,XXX),
  //              donor type badge, and relative time ("5 min ago").

  // TODO Part 2: add donorType filter (select) and donorName search (input).
  //              Derive the visible list with useMemo — do NOT restart the
  //              stream subscription when filters change.
  //              Show "Showing N donations".

  // TODO Part 3: implement optimistic "Send Thanks" per donation row.
  //              Track per-donation loading and error state.
  //              Call sendThankYou(), revert + show inline error on failure.

  return (
    <div style={{ fontFamily: 'sans-serif', maxWidth: 720, margin: '0 auto', padding: 24 }}>
      <h2 style={{ marginBottom: 16 }}>{CAMPAIGN.name}</h2>

      {/* TODO: goal progress bar */}

      {/* TODO: filter controls */}

      {/* TODO: donation list */}
      <p style={{ color: '#888' }}>No donations yet.</p>
    </div>
  )
}

// ---------------------------------------------------------------------------
// Suggested sub-components (optional — structure however you like)
// ---------------------------------------------------------------------------

// function ProgressBar({ raised, goal }: { raised: number; goal: number }) { ... }

// function DonationRow({
//   donation,
//   onSendThanks,
// }: {
//   donation: Donation
//   onSendThanks: (id: string) => void
// }) { ... }

// ---------------------------------------------------------------------------
// App entry point — connects the stream, renders the dashboard
// ---------------------------------------------------------------------------

export default function App() {
  useEffect(() => {
    donationStream.connect()
    return () => donationStream.disconnect()
  }, [])

  return <CampaignDashboard />
}
