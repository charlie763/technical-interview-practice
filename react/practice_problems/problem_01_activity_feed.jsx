/**
 * =============================================================================
 * INTERVIEW PROBLEM 3: Live Activity Feed with Filtering & Optimistic Updates
 * Difficulty: Senior Software Engineer | Estimated time: 45-60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the activity feed for a SaaS developer platform — think
 * Stripe's event log, Linear's notification inbox, or Grafana's alert history.
 *
 * The backend pushes events over a persistent connection (simulated below).
 * Users need to filter the feed and acknowledge events inline.
 *
 * A working mock event source and basic styles are provided. Do not modify
 * anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 *
 * PART 1 — Live event subscription  (~15 min)
 * ─────────────────────────────────────────────
 * Render an <ActivityFeed /> component that:
 *   - Subscribes to `eventSource` on mount and unsubscribes on unmount.
 *   - Displays incoming events in reverse-chronological order (newest on top).
 *   - Each event row shows: timestamp, type badge, actor, and message.
 *   - New events should appear at the top without losing scroll position for
 *     events already in view. (Hint: prepend, don't append.)
 *
 * PART 2 — Client-side filtering  (~15 min)
 * ──────────────────────────────────────────
 * Add filter controls above the feed:
 *   - A dropdown (or set of toggle buttons) for EVENT TYPE
 *     (values: "all" | "deploy" | "alert" | "payment" | "auth")
 *   - A dropdown for STATUS ("all" | "success" | "warning" | "error")
 *
 * Requirements:
 *   - Filtering is purely client-side — all events stay in memory.
 *   - Changing a filter must NOT restart the event subscription.
 *   - Use useMemo (or equivalent) so the filtered list is only recomputed
 *     when events or filter state actually change.
 *
 * PART 3 — Optimistic acknowledgement  (~15 min)
 * ────────────────────────────────────────────────
 * Add an "Ack" button to each unacknowledged event row.
 *
 * When clicked:
 *   1. Immediately mark the event as acknowledged in local state (optimistic).
 *   2. Call `acknowledgeEvent(eventId)` — it returns a Promise that resolves
 *      on success or rejects ~20% of the time to simulate flakiness.
 *   3. On rejection: revert the event to unacknowledged and display an inline
 *      error message on that row ("Failed — try again").
 *   4. While the request is in-flight, the button should show a loading state
 *      and be disabled to prevent double-submission.
 *
 * =============================================================================
 *
 * TEST CONTRACT
 * -------------
 * Playwright tests use the following query strategy (in priority order):
 *
 * 1. Text content from the live event stream (no testids needed for most assertions).
 * 2. Button text — the Ack button is found by /ack/i text. Name it "Ack" or similar.
 * 3. data-testid="event-row" (or a CSS class containing "event") — the ordering
 *    test needs to identify individual rows. Add data-testid="event-row" to each
 *    event row element, or give it a class name that includes "event".
 * 4. locator('select') — filter controls are found by element type; at least one
 *    <select> element must be present for the filter test to pass.
 *
 * =============================================================================
 */

import React, { useState, useEffect, useMemo, useRef, useCallback } from "react";

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

const EVENT_TYPES = ["deploy", "alert", "payment", "auth"];
const STATUSES = ["success", "warning", "error"];
const ACTORS = ["ci-bot", "alice@corp.com", "webhooks-service", "billing-worker", "bob@corp.com"];
const MESSAGES = {
  deploy: ["Deployed v2.4.1 to production", "Rollback triggered on staging", "Build pipeline completed"],
  alert:  ["CPU usage exceeded 90%", "Error rate spike detected", "Latency p99 > 2s"],
  payment: ["Invoice #8821 paid", "Subscription renewed", "Payment method declined"],
  auth:   ["API key created", "OAuth token revoked", "Login from new IP"],
};

let _eventId = 1;

/**
 * createEventSource() → { connect, disconnect, subscribe }
 *
 * connect()           — start emitting events (~every 1.5s)
 * disconnect()        — stop emitting
 * subscribe(handler)  — register a handler; returns an unsubscribe function
 *                       handler receives an event object (see shape below)
 *
 * Event shape:
 * {
 *   id:            string,   // unique identifier
 *   type:          string,   // "deploy" | "alert" | "payment" | "auth"
 *   status:        string,   // "success" | "warning" | "error"
 *   actor:         string,   // who/what triggered the event
 *   message:       string,
 *   timestamp:     string,   // ISO 8601
 *   acknowledged:  false,    // always false when emitted
 * }
 */
export function createEventSource() {
  let handlers = [];
  let timerId = null;

  function emit() {
    const type = EVENT_TYPES[Math.floor(Math.random() * EVENT_TYPES.length)];
    const event = {
      id:           `evt-${_eventId++}`,
      type,
      status:       STATUSES[Math.floor(Math.random() * STATUSES.length)],
      actor:        ACTORS[Math.floor(Math.random() * ACTORS.length)],
      message:      MESSAGES[type][Math.floor(Math.random() * MESSAGES[type].length)],
      timestamp:    new Date().toISOString(),
      acknowledged: false,
    };
    handlers.forEach((h) => h(event));
  }

  return {
    connect:    () => { timerId = setInterval(emit, 1500); },
    disconnect: () => { clearInterval(timerId); timerId = null; },
    subscribe:  (handler) => {
      handlers.push(handler);
      return () => { handlers = handlers.filter((h) => h !== handler); };
    },
  };
}

/**
 * acknowledgeEvent(eventId) → Promise<void>
 * Simulates a PATCH /events/:id API call.
 * Resolves after ~400ms. Rejects ~20% of the time.
 */
export function acknowledgeEvent(eventId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.2) {
        reject(new Error(`Failed to acknowledge ${eventId}`));
      } else {
        resolve();
      }
    }, 400);
  });
}

// Shared event source instance — import this in your component
export const eventSource = createEventSource();

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

/**
 * STATUS_COLORS and TYPE_LABELS are helpers you may use for styling.
 * Feel free to add your own or ignore these.
 */
const STATUS_COLORS = {
  success: "#22c55e",
  warning: "#f59e0b",
  error:   "#ef4444",
};

const TYPE_LABELS = {
  deploy:  "Deploy",
  alert:   "Alert",
  payment: "Payment",
  auth:    "Auth",
};

// ---------------------------------------------------------------------------
// PART 1 — Implement ActivityFeed
// ---------------------------------------------------------------------------

export function ActivityFeed() {
  // TODO Part 1: subscribe to eventSource on mount, unsubscribe on unmount.
  //              Store events in state and render them newest-first.

  // TODO Part 2: add filter state (type, status) and filter the event list
  //              with useMemo. Render filter controls above the list.
  //              The subscription must NOT restart when filters change.

  // TODO Part 3: implement optimistic acknowledgement.
  //              Track per-event loading and error state.
  //              Call acknowledgeEvent(), revert + show error on failure.

  return (
    <div style={{ fontFamily: "monospace", maxWidth: 700, margin: "0 auto", padding: 24 }}>
      <h2 style={{ marginBottom: 16 }}>Activity Feed</h2>

      {/* TODO: filter controls go here */}

      {/* TODO: event list goes here */}
      <p style={{ color: "#888" }}>No events yet.</p>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Suggested sub-components (optional — structure however you like)
// ---------------------------------------------------------------------------

// function FilterBar({ typeFilter, statusFilter, onTypeChange, onStatusChange }) { ... }

// function EventRow({ event, onAcknowledge }) { ... }

// ---------------------------------------------------------------------------
// App entry point — renders the feed
// ---------------------------------------------------------------------------

export default function App() {
  // Connect the event source once when the app mounts
  useEffect(() => {
    eventSource.connect();
    return () => eventSource.disconnect();
  }, []);

  return <ActivityFeed />;
}

/**
 * =============================================================================
 * SETUP NOTE
 * =============================================================================
 * To run this in a browser, create a Vite React project and drop this file in:
 *
 *   npm create vite@latest interview-react -- --template react
 *   cd interview-react
 *   # replace src/App.jsx with this file (or import ActivityFeed from it)
 *   npm install && npm run dev
 *
 * Or with Create React App:
 *   npx create-react-app interview-react
 *   # replace src/App.js with this file
 *   npm start
 * =============================================================================
 */
