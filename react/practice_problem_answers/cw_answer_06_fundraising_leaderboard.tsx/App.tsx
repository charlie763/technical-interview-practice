/**
 * =============================================================================
 * INTERVIEW PROBLEM 06: Fundraising Class Leaderboard
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the alumni class leaderboard for a K-12 school fundraising
 * platform. Schools run peer-to-peer campaigns where alumni classes compete to
 * see which graduation year can raise the most. Advancement teams and class
 * volunteers track real-time progress on this leaderboard throughout a campaign.
 *
 * Seed data and API stubs are provided below.
 * Do NOT modify anything in the "PROVIDED — DO NOT MODIFY" sections.
 *
 * =============================================================================
 * PROVIDED (do not modify)
 * =============================================================================
 *
 * SEED_COHORTS      — 6 deterministic class cohorts used in Playwright tests.
 * addPledge(cohortId, amount) → Promise
 *                   — mock API call; rejects ~15% of the time.
 *
 * =============================================================================
 * PART 1 — Ranked leaderboard  (~15 min)
 * =============================================================================
 *
 * Render a <Leaderboard /> component that displays SEED_COHORTS sorted by
 * amountRaised descending (highest raised = Rank #1).
 *
 * Each row must include:
 *   • Rank number (#1, #2, …)
 *   • Class name (e.g. "Class of 1985")
 *   • Lead volunteer name
 *   • Amount raised (formatted, e.g. "$22,150")
 *   • Donor count
 *   • Percentage of goal reached (e.g. "88%" — may exceed 100%)
 *
 * Required data-testid attributes:
 *   data-testid="cohort-row"   — wrapper element for each class row
 *
 * =============================================================================
 * PART 2 — Sorting and search  (~15 min)
 * =============================================================================
 *
 * Add sorting controls and a search input above the leaderboard:
 *
 * 1. Sort by Amount Raised — a button that toggles asc/desc by amountRaised.
 *    Tests accept data-testid="sort-raised" OR a button whose visible text
 *    contains "raised" (case-insensitive).
 *
 * 2. Sort by Donor Count — a button that toggles asc/desc by donorCount.
 *    Tests accept data-testid="sort-donors" OR a button whose visible text
 *    contains "donor" (case-insensitive).
 *
 * 3. Name search — an <input> that filters rows by class name (case-insensitive
 *    substring). Tests accept data-testid="search-input" OR a placeholder
 *    attribute containing "search" (case-insensitive).
 *
 * Requirements:
 *   - Derive the displayed list with useMemo.
 *   - Show "Showing N classes" reflecting the filtered count.
 *
 * =============================================================================
 * PART 3 — Optimistic pledge entry  (~15 min)
 * =============================================================================
 *
 * Add a "Pledge" button to each cohort row. When clicked:
 *   - Reveal a number <input> (data-testid="pledge-input") and a "Submit"
 *     button inline on that row.
 *   - On Submit: call addPledge(cohortId, pledgeAmount) and apply an optimistic
 *     update that immediately adds the entered amount to that cohort's
 *     amountRaised in local state. Show a spinner (data-testid="pledge-spinner")
 *     while the request is in-flight.
 *   - On success: keep the updated total and hide the pledge form.
 *   - On failure (~15%): revert amountRaised to its previous value and show
 *     an inline error on that row ("Pledge failed — try again").
 *
 * Required data-testid attributes:
 *   data-testid="pledge-input"    — the pledge amount <input>
 *   data-testid="pledge-spinner"  — loading indicator during submit
 *
 * =============================================================================
 *
 * TYPE CONTRACT
 * -------------
 * The following types are exported from the PROVIDED section below.
 * Do not redefine them.
 *
 * interface Cohort { id, name, graduationYear, goalAmount, amountRaised,
 *                    donorCount, leadVolunteer }
 * type SortField   = 'raised' | 'donors'
 * type SortDir     = 'asc' | 'desc'
 *
 * =============================================================================
 *
 * TEST CONTRACT
 * -------------
 * Playwright tests use the following query strategy (in priority order):
 *
 * 1. Seed data text — toContainText('Class of 1985'), toContainText('22,150') etc.
 * 2. Element type + visible text — locator('button').filter({ hasText: /pledge/i })
 *    or locator('button').filter({ hasText: /submit/i })
 * 3. Input attributes — getByPlaceholder(/search/i) for the name search input
 * 4. data-testid — required for: cohort rows (counting), pledge-input, and
 *    pledge-spinner (listed above). Sort buttons also accept data-testid.
 *
 * =============================================================================
 */

import React, { useState, useMemo } from "react";

// =============================================================================
// PROVIDED — DO NOT MODIFY
// =============================================================================

export interface Cohort {
  id: string;
  name: string;
  graduationYear: number;
  goalAmount: number;
  amountRaised: number;
  donorCount: number;
  leadVolunteer: string;
}

export type SortField = "raised" | "donors";
export type SortDir = "asc" | "desc";

/**
 * Deterministic seed cohorts — used directly by Playwright tests.
 *
 * Default sort (amountRaised desc) produces this ranking:
 *   #1  Class of 1985  $22,150  47 donors  (88% of $25k goal)
 *   #2  Class of 1995  $18,900  62 donors  (94% of $20k goal)
 *   #3  Class of 2010  $11,200  54 donors  (112% of $10k goal — over!)
 *   #4  Class of 2005   $8,400  31 donors  (56% of $15k goal)
 *   #5  Class of 2015   $5,600  28 donors  (70% of $8k goal)
 *   #6  Class of 2020   $2,100  19 donors  (42% of $5k goal)
 *
 * By donorCount desc:
 *   #1  Class of 1995 (62)  #2  Class of 2010 (54)  #3  Class of 1985 (47) …
 */
export const SEED_COHORTS: Cohort[] = [
  {
    id: "cls-1985",
    name: "Class of 1985",
    graduationYear: 1985,
    goalAmount: 25_000,
    amountRaised: 22_150,
    donorCount: 47,
    leadVolunteer: "Patricia Nguyen",
  },
  {
    id: "cls-1995",
    name: "Class of 1995",
    graduationYear: 1995,
    goalAmount: 20_000,
    amountRaised: 18_900,
    donorCount: 62,
    leadVolunteer: "David Torres",
  },
  {
    id: "cls-2005",
    name: "Class of 2005",
    graduationYear: 2005,
    goalAmount: 15_000,
    amountRaised: 8_400,
    donorCount: 31,
    leadVolunteer: "Aisha Johnson",
  },
  {
    id: "cls-2010",
    name: "Class of 2010",
    graduationYear: 2010,
    goalAmount: 10_000,
    amountRaised: 11_200,
    donorCount: 54,
    leadVolunteer: "Marcus Lee",
  },
  {
    id: "cls-2015",
    name: "Class of 2015",
    graduationYear: 2015,
    goalAmount: 8_000,
    amountRaised: 5_600,
    donorCount: 28,
    leadVolunteer: "Priya Sharma",
  },
  {
    id: "cls-2020",
    name: "Class of 2020",
    graduationYear: 2020,
    goalAmount: 5_000,
    amountRaised: 2_100,
    donorCount: 19,
    leadVolunteer: "Tyler Brooks",
  },
];

/**
 * addPledge(cohortId, amount) → Promise<{ cohortId, amount }>
 * Simulates a POST /cohorts/:id/pledges API call.
 * Resolves after ~400 ms. Rejects ~15% of the time.
 */
export function addPledge(
  cohortId: string,
  amount: number,
): Promise<{ cohortId: string; amount: number }> {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      if (Math.random() < 0.15) {
        reject(new Error("Pledge failed — please try again."));
      } else {
        resolve({ cohortId, amount });
      }
    }, 400);
  });
}

// =============================================================================
// YOUR WORK STARTS HERE
// =============================================================================

// ---------------------------------------------------------------------------
// PART 1 — Implement Leaderboard
// ---------------------------------------------------------------------------

export function Leaderboard() {
  const [cohorts, setCohorts] = useState<Cohort[]>(SEED_COHORTS);
  const [sortField, setSortField] = useState<SortField>("raised");
  const [sortDir, setSortDir] = useState<SortDir>("desc");
  const [nameSearch, setNameSearch] = useState("");
  const sortedFilteredCohorts = useMemo(() => {
    let sortFunc = (cohortA: Cohort, cohortB: Cohort) =>
      cohortB.amountRaised - cohortA.amountRaised;
    if (sortField == "raised" && sortDir == "asc") {
      sortFunc = (cohortA: Cohort, cohortB: Cohort) =>
        cohortA.amountRaised - cohortB.amountRaised;
    } else if (sortField == "donors" && sortDir == "desc") {
      sortFunc = (cohortA: Cohort, cohortB: Cohort) =>
        cohortB.donorCount - cohortA.donorCount;
    } else if (sortField == "donors" && sortDir == "asc") {
      sortFunc = (cohortA: Cohort, cohortB: Cohort) =>
        cohortA.donorCount - cohortB.donorCount;
    }
    const sorted = cohorts.sort(sortFunc);
    return sorted.filter((cohort: Cohort) => cohort.name.includes(nameSearch));
  }, [cohorts, sortField, sortDir, nameSearch]);

  const handleSortButtonClick = (sortField: SortField) => {
    setSortField(sortField);
    setSortDir((prevDir) => {
      if (prevDir == "desc") {
        return "asc";
      } else {
        return "desc";
      }
    });
  };

  const handleNameSearchChange = (name: string) => {
    setNameSearch(name);
  };

  // TODO Part 2: add sort buttons (raised, donors) that each toggle asc/desc.
  //              Add a name search input that filters rows by class name.
  //              Derive the displayed list with useMemo.
  //              Show "Showing N classes".

  // TODO Part 3: add a "Pledge" button per row.
  //              On click, reveal a pledge-input and a "Submit" button.
  //              Apply an optimistic update to amountRaised; show pledge-spinner.
  //              Revert on failure and show an inline error.

  return (
    <div
      style={{
        fontFamily: "sans-serif",
        maxWidth: 800,
        margin: "0 auto",
        padding: 24,
      }}
    >
      <h2 style={{ marginBottom: 16 }}>Class Fundraising Leaderboard</h2>

      <div
        style={{
          display: "flex",

          gap: "6px",
        }}
      >
        <button onClick={() => handleSortButtonClick("raised")}>
          {`Amount Raised${sortField == "raised" ? (sortDir == "desc" ? " desc" : " asc") : ""}`}
        </button>
        <button onClick={() => handleSortButtonClick("donors")}>
          {`Donor Count${sortField == "donors" ? (sortDir == "desc" ? " desc" : " asc") : ""}`}
        </button>
        <input
          placeholder="search by name"
          type="text"
          value={nameSearch}
          onChange={(event) => handleNameSearchChange(event.target.value)}
        />
      </div>

      {sortedFilteredCohorts.length > 0 ? (
        sortedFilteredCohorts.map((cohort, idx) => (
          <div
            key={cohort.id}
            data-testid="cohort-row"
            style={{
              display: "flex",
              justifyContent: "space-between",
              gap: "6px",
            }}
          >
            <span>{`#${idx + 1}`}</span>
            <span>{cohort.name}</span>
            <span>{cohort.leadVolunteer}</span>
            <span>{`$${cohort.amountRaised}`}</span>
            <span>{cohort.donorCount}</span>
            <span>{`${((cohort.amountRaised / cohort.goalAmount) * 100).toFixed(2)}%`}</span>
          </div>
        ))
      ) : (
        <p style={{ color: "#888" }}>No classes found.</p>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Suggested sub-components (optional — structure however you like)
// ---------------------------------------------------------------------------

// function CohortRow({
//   cohort,
//   rank,
//   onPledge,
// }: {
//   cohort: Cohort
//   rank: number
//   onPledge: (cohortId: string, amount: number) => Promise<void>
// }) { ... }

// ---------------------------------------------------------------------------
// App entry point
// ---------------------------------------------------------------------------

export default function App() {
  return <Leaderboard />;
}
