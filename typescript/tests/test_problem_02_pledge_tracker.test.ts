/**
 * Tests for Problem 02 — Walkathon Pledge Tracker
 *
 * Run (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_02_pledge_tracker npm run test:02
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { PledgeTracker } from '../practice_problems/problem_02_pledge_tracker'

// ── Part 1 ─────────────────────────────────────────────────────────────────

describe('Part 1 — Participant registration and pledge collection', () => {
  let pt: PledgeTracker

  beforeEach(() => {
    pt = new PledgeTracker()
  })

  it('registerParticipant returns object with correct fields', () => {
    const p = pt.registerParticipant('p1', 'Emma Torres', 'class-5a')
    expect(p.id).toBe('p1')
    expect(p.name).toBe('Emma Torres')
    expect(p.classId).toBe('class-5a')
    expect(p.lapsCompleted).toBe(0)
  })

  it('registerParticipant throws on duplicate id', () => {
    pt.registerParticipant('p-dup', 'A', 'class-1')
    expect(() => pt.registerParticipant('p-dup', 'B', 'class-1')).toThrow()
  })

  it('addPledge returns object with correct fields', () => {
    pt.registerParticipant('p1', 'Emma', 'class-5a')
    const pledge = pt.addPledge('p1', 'Grandma Rose', 'rose@example.com', 'per_lap', 5)
    expect(pledge.participantId).toBe('p1')
    expect(pledge.sponsorName).toBe('Grandma Rose')
    expect(pledge.sponsorEmail).toBe('rose@example.com')
    expect(pledge.pledgeType).toBe('per_lap')
    expect(pledge.amount).toBe(5)
    expect(pledge.id).toBeTruthy()
  })

  it('addPledge assigns unique auto-generated IDs', () => {
    pt.registerParticipant('p1', 'Emma', 'class-5a')
    const a = pt.addPledge('p1', 'A', 'a@x.com', 'flat', 10)
    const b = pt.addPledge('p1', 'B', 'b@x.com', 'flat', 20)
    expect(a.id).not.toBe(b.id)
  })

  it('addPledge throws on unknown participant', () => {
    expect(() => pt.addPledge('no-such', 'A', 'a@x.com', 'flat', 10)).toThrow()
  })

  it('addPledge throws on non-positive amount', () => {
    pt.registerParticipant('p1', 'Emma', 'class-5a')
    expect(() => pt.addPledge('p1', 'A', 'a@x.com', 'flat', 0)).toThrow()
    expect(() => pt.addPledge('p1', 'A', 'a@x.com', 'flat', -5)).toThrow()
  })

  it('getParticipant returns the correct participant', () => {
    pt.registerParticipant('p1', 'Emma', 'class-5a')
    const p = pt.getParticipant('p1')
    expect(p.id).toBe('p1')
  })

  it('getParticipant throws on unknown id', () => {
    expect(() => pt.getParticipant('no-such')).toThrow()
  })
})

// ── Part 2 ─────────────────────────────────────────────────────────────────

describe('Part 2 — Lap recording and totals', () => {
  let pt: PledgeTracker

  beforeEach(() => {
    pt = new PledgeTracker()
    // p1: 2 pledges — $5/lap + $50 flat
    pt.registerParticipant('p1', 'Emma Torres', 'class-5a')
    pt.addPledge('p1', 'Grandma Rose', 'rose@example.com', 'per_lap', 5)
    pt.addPledge('p1', 'Uncle Joe',    'joe@example.com',  'flat', 50)
    // p2: 1 pledge — $10/lap
    pt.registerParticipant('p2', 'Liam Chen', 'class-5a')
    pt.addPledge('p2', 'Dad', 'dad@example.com', 'per_lap', 10)
    // p3: 1 pledge — $100 flat
    pt.registerParticipant('p3', 'Sofia Park', 'class-6b')
    pt.addPledge('p3', 'Family Friend', 'ff@example.com', 'flat', 100)
  })

  it('recordLaps updates lapsCompleted', () => {
    const p = pt.recordLaps('p1', 10)
    expect(p.lapsCompleted).toBe(10)
  })

  it('recordLaps throws on unknown participant', () => {
    expect(() => pt.recordLaps('no-such', 5)).toThrow()
  })

  it('recordLaps throws on negative laps', () => {
    expect(() => pt.recordLaps('p1', -1)).toThrow()
  })

  it('getParticipantTotal computes perLapTotal, flatTotal, totalRaised', () => {
    pt.recordLaps('p1', 10)
    const total = pt.getParticipantTotal('p1')
    expect(total.perLapTotal).toBe(50)   // 10 × $5
    expect(total.flatTotal).toBe(50)     // $50 flat
    expect(total.totalRaised).toBe(100)
    expect(total.lapsCompleted).toBe(10)
    expect(total.participantId).toBe('p1')
    expect(total.name).toBe('Emma Torres')
  })

  it('getParticipantTotal returns 0 for all when laps=0', () => {
    const total = pt.getParticipantTotal('p1')
    expect(total.perLapTotal).toBe(0)
    expect(total.totalRaised).toBe(50)   // flat still counts
  })

  it('getParticipantTotal handles only flat pledges', () => {
    pt.recordLaps('p3', 15)
    const total = pt.getParticipantTotal('p3')
    expect(total.perLapTotal).toBe(0)
    expect(total.flatTotal).toBe(100)
    expect(total.totalRaised).toBe(100)
  })

  it('getParticipantTotal throws on unknown participant', () => {
    expect(() => pt.getParticipantTotal('no-such')).toThrow()
  })

  it('getCampaignTotal returns sum across all participants', () => {
    pt.recordLaps('p1', 12) // 12×5 + 50 = 110
    pt.recordLaps('p2', 8)  // 8×10 = 80
    pt.recordLaps('p3', 15) // 100
    expect(pt.getCampaignTotal()).toBe(290)
  })
})

// ── Part 3 ─────────────────────────────────────────────────────────────────

describe('Part 3 — Class leaderboard and participant ranking', () => {
  let pt: PledgeTracker

  beforeEach(() => {
    pt = new PledgeTracker()
    pt.registerParticipant('p1', 'Emma Torres', 'class-5a')
    pt.addPledge('p1', 'Grandma', 'g@x.com', 'per_lap', 5)
    pt.addPledge('p1', 'Uncle',   'u@x.com', 'flat', 50)
    pt.recordLaps('p1', 12)  // perLap=60, flat=50 → 110

    pt.registerParticipant('p2', 'Liam Chen', 'class-5a')
    pt.addPledge('p2', 'Dad', 'd@x.com', 'per_lap', 10)
    pt.recordLaps('p2', 8)   // 80

    pt.registerParticipant('p3', 'Sofia Park', 'class-6b')
    pt.addPledge('p3', 'Friend', 'f@x.com', 'flat', 100)
    pt.recordLaps('p3', 1)   // 100

    // class-5a total: 190, class-6b total: 100
  })

  it('getClassLeaderboard sorts classes by totalRaised descending', () => {
    const lb = pt.getClassLeaderboard()
    expect(lb[0].classId).toBe('class-5a')
    expect(lb[0].totalRaised).toBe(190)
    expect(lb[1].classId).toBe('class-6b')
    expect(lb[1].totalRaised).toBe(100)
  })

  it('getClassLeaderboard includes participantCount', () => {
    const lb = pt.getClassLeaderboard()
    const a = lb.find(c => c.classId === 'class-5a')
    expect(a?.participantCount).toBe(2)
  })

  it('getClassLeaderboard includes topFundraiser for each class', () => {
    const lb = pt.getClassLeaderboard()
    const a = lb.find(c => c.classId === 'class-5a')
    expect(a?.topFundraiser.participantId).toBe('p1')
    expect(a?.topFundraiser.totalRaised).toBe(110)
  })

  it('getParticipantLeaderboard ranks by totalRaised descending', () => {
    const lb = pt.getParticipantLeaderboard()
    expect(lb[0].participantId).toBe('p1')
    expect(lb[0].rank).toBe(1)
    expect(lb[0].totalRaised).toBe(110)
    expect(lb[1].rank).toBe(2)
    expect(lb[2].rank).toBe(3)
  })

  it('getParticipantLeaderboard respects limit', () => {
    const lb = pt.getParticipantLeaderboard(2)
    expect(lb).toHaveLength(2)
  })

  it('getParticipantLeaderboard returns all when limit is undefined', () => {
    const lb = pt.getParticipantLeaderboard()
    expect(lb).toHaveLength(3)
  })
})
