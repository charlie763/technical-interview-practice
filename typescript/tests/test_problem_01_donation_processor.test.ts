/**
 * Tests for Problem 01 — Donation Processor
 *
 * Run (from typescript/):
 *   PRACTICE_ANSWER=cw_answer_01_donation_processor npm run test:01
 */

import { describe, it, expect, beforeEach } from "vitest";
import { DonationProcessor } from "../practice_problems/problem_01_donation_processor";

// ── Shared timestamps ───────────────────────────────────────────────────────
const T1 = "2025-09-05T10:00:00";
const T2 = "2025-09-06T11:00:00";
const T3 = "2025-09-07T09:00:00";
const T4 = "2025-09-08T14:00:00";

// ── Part 1 ─────────────────────────────────────────────────────────────────

describe("Part 1 — Campaign setup, donation recording, and stats", () => {
  let dp: DonationProcessor;

  beforeEach(() => {
    dp = new DonationProcessor();
  });

  // createCampaign
  it("createCampaign returns a campaign with correct fields", () => {
    const c = dp.createCampaign("c1", "Annual Fund", 50_000, "2025-09-01");
    expect(c.id).toBe("c1");
    expect(c.name).toBe("Annual Fund");
    expect(c.goal).toBe(50_000);
    expect(c.startDate).toBe("2025-09-01");
  });

  it("createCampaign throws on duplicate id", () => {
    dp.createCampaign("c-dup", "A", 10_000, "2025-09-01");
    expect(() =>
      dp.createCampaign("c-dup", "B", 20_000, "2025-09-01"),
    ).toThrow();
  });

  // addDonation
  it("addDonation returns a donation with correct fields", () => {
    dp.createCampaign("c1", "Fund", 10_000, "2025-09-01");
    const d = dp.addDonation("c1", "Alice Smith", "alice@school.org", 500, T1);
    expect(d.campaignId).toBe("c1");
    expect(d.donorName).toBe("Alice Smith");
    expect(d.donorEmail).toBe("alice@school.org");
    expect(d.amount).toBe(500);
    expect(d.timestamp).toBe(T1);
    expect(d.refunded).toBe(false);
  });

  it("addDonation assigns unique auto-generated IDs", () => {
    dp.createCampaign("c1", "Fund", 10_000, "2025-09-01");
    const d1 = dp.addDonation("c1", "A", "a@x.com", 100, T1);
    const d2 = dp.addDonation("c1", "B", "b@x.com", 100, T2);
    expect(d1.id).toBeTruthy();
    expect(d2.id).toBeTruthy();
    expect(d1.id).not.toBe(d2.id);
  });

  it("addDonation throws on unknown campaign", () => {
    expect(() => dp.addDonation("no-such", "A", "a@x.com", 100, T1)).toThrow();
  });

  it("addDonation throws on non-positive amount", () => {
    dp.createCampaign("c1", "Fund", 10_000, "2025-09-01");
    expect(() => dp.addDonation("c1", "A", "a@x.com", 0, T1)).toThrow();
    expect(() => dp.addDonation("c1", "A", "a@x.com", -50, T1)).toThrow();
  });

  // getDonation
  it("getDonation returns the correct donation", () => {
    dp.createCampaign("c1", "Fund", 10_000, "2025-09-01");
    const d = dp.addDonation("c1", "Alice", "alice@x.com", 500, T1);
    expect(dp.getDonation(d.id)).toEqual(d);
  });

  it("getDonation throws on unknown id", () => {
    expect(() => dp.getDonation("no-such")).toThrow();
  });

  // getCampaignStats
  it("getCampaignStats returns correct totals", () => {
    dp.createCampaign("camp-seed", "Annual Fund", 100_000, "2025-09-01");
    dp.addDonation("camp-seed", "Alice", "alice@x.com", 500, T1);
    dp.addDonation("camp-seed", "Bob", "bob@x.com", 1_000, T2);
    dp.addDonation("camp-seed", "Alice", "alice@x.com", 200, T3);
    dp.addDonation("camp-seed", "Carol", "carol@x.com", 750, T4);
    const stats = dp.getCampaignStats("camp-seed");
    expect(stats.totalRaised).toBe(2_450);
    expect(stats.donationCount).toBe(4);
    expect(stats.donorCount).toBe(3); // alice, bob, carol
    expect(stats.goalPercent).toBe(2.5); // 2450/100000 * 100
  });

  it("getCampaignStats on empty campaign returns zeros", () => {
    dp.createCampaign("empty", "Empty", 10_000, "2025-09-01");
    const stats = dp.getCampaignStats("empty");
    expect(stats.totalRaised).toBe(0);
    expect(stats.donorCount).toBe(0);
    expect(stats.donationCount).toBe(0);
    expect(stats.goalPercent).toBe(0);
  });

  it("getCampaignStats throws on unknown campaign", () => {
    expect(() => dp.getCampaignStats("no-such")).toThrow();
  });
});

// ── Part 2 ─────────────────────────────────────────────────────────────────

describe("Part 2 — Top donors and date-range search", () => {
  let dp: DonationProcessor;

  beforeEach(() => {
    dp = new DonationProcessor();
    dp.createCampaign("camp-seed", "Annual Fund", 100_000, "2025-09-01");
    // alice total: 700, bob: 1000, carol: 750
    dp.addDonation("camp-seed", "Alice Smith", "alice@x.com", 500, T1);
    dp.addDonation("camp-seed", "Bob Jones", "bob@x.com", 1_000, T2);
    dp.addDonation("camp-seed", "Alice Smith", "alice@x.com", 200, T3);
    dp.addDonation("camp-seed", "Carol Lee", "carol@x.com", 750, T4);
  });

  it("getTopDonors returns donors sorted by total amount descending", () => {
    const top = dp.getTopDonors("camp-seed", 3);
    expect(top[0].donorEmail).toBe("bob@x.com");
    expect(top[0].totalAmount).toBe(1_000);
    expect(top[1].donorEmail).toBe("carol@x.com");
    expect(top[1].totalAmount).toBe(750);
    expect(top[2].donorEmail).toBe("alice@x.com");
    expect(top[2].totalAmount).toBe(700);
  });

  it("getTopDonors aggregates multiple donations from the same email", () => {
    const top = dp.getTopDonors("camp-seed", 5);
    const alice = top.find((d) => d.donorEmail === "alice@x.com");
    expect(alice?.totalAmount).toBe(700);
    expect(alice?.donationCount).toBe(2);
  });

  it("getTopDonors respects the limit", () => {
    const top = dp.getTopDonors("camp-seed", 2);
    expect(top).toHaveLength(2);
  });

  it("getTopDonors throws on unknown campaign", () => {
    expect(() => dp.getTopDonors("no-such", 5)).toThrow();
  });

  it("getDonationsInRange returns donations within [from, to] inclusive", () => {
    const inRange = dp.getDonationsInRange("camp-seed", T1, T3);
    const amounts = inRange.map((d) => d.amount).sort((a, b) => a - b);
    expect(amounts).toEqual([200, 500, 1_000]);
  });

  it("getDonationsInRange excludes donations outside the range", () => {
    const inRange = dp.getDonationsInRange("camp-seed", T1, T2);
    expect(inRange.every((d) => d.timestamp <= T2 && d.timestamp >= T1)).toBe(
      true,
    );
  });

  it("getDonationsInRange returns donations sorted by timestamp ascending", () => {
    const inRange = dp.getDonationsInRange("camp-seed", T1, T4);
    const timestamps = inRange.map((d) => d.timestamp);
    expect(timestamps).toEqual([...timestamps].sort());
  });

  it("getDonationsInRange throws on unknown campaign", () => {
    expect(() => dp.getDonationsInRange("no-such", T1, T4)).toThrow();
  });
});

// ── Part 3 ─────────────────────────────────────────────────────────────────

describe("Part 3 — Refunds and matching gifts", () => {
  let dp: DonationProcessor;
  let d1Id: string;
  let d2Id: string;

  beforeEach(() => {
    dp = new DonationProcessor();
    dp.createCampaign("camp-seed", "Annual Fund", 100_000, "2025-09-01");
    const d1 = dp.addDonation("camp-seed", "Alice", "alice@x.com", 500, T1);
    const d2 = dp.addDonation("camp-seed", "Bob", "bob@x.com", 1_000, T2);
    d1Id = d1.id;
    d2Id = d2.id;
  });

  // refundDonation
  it("refundDonation marks the donation as refunded", () => {
    const refunded = dp.refundDonation(d1Id);
    expect(refunded.refunded).toBe(true);
  });

  it("refundDonation excludes the donation from stats", () => {
    dp.refundDonation(d1Id);
    const stats = dp.getCampaignStats("camp-seed");
    expect(stats.totalRaised).toBe(1_000);
    expect(stats.donationCount).toBe(1);
    expect(stats.donorCount).toBe(1);
  });

  it("refundDonation throws on unknown id", () => {
    expect(() => dp.refundDonation("no-such")).toThrow();
  });

  it("refundDonation throws if already refunded", () => {
    dp.refundDonation(d1Id);
    expect(() => dp.refundDonation(d1Id)).toThrow();
  });

  // addMatchingGift
  it("addMatchingGift creates a donation with min(amount*ratio, max)", () => {
    // source: 500, ratio: 0.5, max: 400 → match = min(250, 400) = 250
    const match = dp.addMatchingGift(d1Id, "Corp Fund", "corp@x.com", 0.5, 400);
    expect(match.amount).toBe(250);
    expect(match.donorName).toBe("Corp Fund");
    expect(match.donorEmail).toBe("corp@x.com");
    expect(match.campaignId).toBe("camp-seed");
  });

  it("addMatchingGift respects the maxAmount cap", () => {
    // source: 1000, ratio: 1.0, max: 300 → match = min(1000, 300) = 300
    const match = dp.addMatchingGift(d2Id, "Matcher", "m@x.com", 1.0, 300);
    expect(match.amount).toBe(300);
  });

  it("addMatchingGift adds the matched amount to campaign stats", () => {
    dp.addMatchingGift(d1Id, "Corp", "corp@x.com", 0.5, 400);
    const stats = dp.getCampaignStats("camp-seed");
    expect(stats.totalRaised).toBe(1_750); // 500 + 1000 + 250
  });

  it("addMatchingGift throws if source donation is refunded", () => {
    dp.refundDonation(d1Id);
    expect(() =>
      dp.addMatchingGift(d1Id, "Corp", "c@x.com", 0.5, 400),
    ).toThrow();
  });

  it("addMatchingGift throws on unknown source donation id", () => {
    expect(() =>
      dp.addMatchingGift("no-such", "Corp", "c@x.com", 0.5, 400),
    ).toThrow();
  });
});
