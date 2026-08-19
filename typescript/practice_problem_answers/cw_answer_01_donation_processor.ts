/**
 * =============================================================================
 * INTERVIEW PROBLEM 01: Donation Processor
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the core donation module for a K-12 school fundraising
 * platform. Schools run campaigns — annual funds, giving days, capital
 * campaigns — and donations flow in from families, alumni, and community
 * supporters. Each campaign has a dollar goal, and you need to track progress,
 * identify top donors, and handle edge cases like refunds and matching gifts.
 *
 * Store all state in instance variables initialised in the constructor.
 * You choose the internal data structures; the public interface is what matters.
 *
 * DATA MODEL
 * ----------
 * Campaign: { id, name, goal, startDate }
 * Donation:
 *   {
 *     id:          string   — auto-generated (e.g. "don-1", "don-2", …)
 *     campaignId:  string
 *     donorName:   string
 *     donorEmail:  string
 *     amount:      number   — dollars, must be > 0
 *     timestamp:   string   — ISO-8601 datetime
 *     refunded:    boolean  — starts false
 *   }
 *
 * // Example
 * // const dp = new DonationProcessor()
 * // dp.createCampaign('camp-1', 'Annual Fund', 50_000, '2025-09-01')
 * // dp.addDonation('camp-1', 'Alice Smith', 'alice@school.org', 500, '2025-09-05T10:00:00')
 * // dp.addDonation('camp-1', 'Bob Jones',   'bob@school.org',  1_000, '2025-09-06T11:00:00')
 * // dp.getCampaignStats('camp-1')
 * // → { totalRaised: 1500, donorCount: 2, donationCount: 2, goalPercent: 3.0 }
 * // dp.getTopDonors('camp-1', 2)
 * // → [{ donorEmail: 'bob@school.org', donorName: 'Bob Jones', totalAmount: 1000, donationCount: 1 },
 * //    { donorEmail: 'alice@school.org', ..., totalAmount: 500, donationCount: 1 }]
 *
 * =============================================================================
 * PART 1 — Campaign setup, donation recording, and basic stats
 * =============================================================================
 *
 * Implement createCampaign, addDonation, getDonation, and getCampaignStats.
 */

// ── Types ──────────────────────────────────────────────────────────────────

export interface Campaign {
  id: string;
  name: string;
  goal: number; // dollars
  startDate: string; // ISO-8601 date
}

export interface Donation {
  id: string;
  campaignId: string;
  donorName: string;
  donorEmail: string;
  amount: number; // dollars
  timestamp: string; // ISO-8601 datetime
  refunded: boolean;
}

export interface CampaignStats {
  totalRaised: number; // sum of non-refunded donation amounts
  donorCount: number; // unique donor emails across non-refunded donations
  donationCount: number; // non-refunded donation count
  goalPercent: number; // (totalRaised / goal) * 100, rounded to 1 decimal place
}

export interface DonorSummary {
  donorEmail: string;
  donorName: string; // name from the most recent donation by this email
  totalAmount: number; // sum of non-refunded donations from this email for the campaign
  donationCount: number; // non-refunded count
}

interface CampaignWithDonations extends Campaign {
  donations: Record<string, Donation>;
}

// ── Class ──────────────────────────────────────────────────────────────────

export class DonationProcessor {
  campaignData: Record<string, CampaignWithDonations>;
  constructor() {
    this.campaignData = {};
  }

  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Create a new campaign.
   *
   * @throws {Error} if id already exists
   */
  createCampaign(
    id: string,
    name: string,
    goal: number,
    startDate: string,
  ): Campaign {
    const existingCampaign = this.campaignData[id];
    if (existingCampaign) {
      throw new Error("tbd");
    }
    const newCampaign = { id, name, goal, startDate };
    this.campaignData[id] = { ...newCampaign, donations: {} };
    return newCampaign;
  }

  /**
   * Record a donation to an existing campaign.
   * Donation IDs are auto-generated ("don-1", "don-2", …) in insertion order.
   *
   * @throws {Error} if campaignId does not exist
   * @throws {Error} if amount <= 0
   */
  addDonation(
    campaignId: string,
    donorName: string,
    donorEmail: string,
    amount: number,
    timestamp: string,
  ): Donation {
    if (amount <= 0) {
      throw new Error("amount must be greater than 0");
    }
    const existingCampaign = this.campaignData[campaignId];
    if (!existingCampaign) {
      throw new Error("campaign doesn't exist");
    }
    const currentDonations = existingCampaign.donations;
    const newDonationId = `don-${Object.values(currentDonations).length + 1}`;
    const newDonation = {
      id: newDonationId,
      campaignId,
      donorName,
      donorEmail,
      amount, // dollars
      timestamp, // ISO-8601 datetime
      refunded: false,
    };
    existingCampaign.donations = {
      ...currentDonations,
      [newDonationId]: newDonation,
    };
    return newDonation;
  }

  /**
   * Return a donation by ID.
   * @throws {Error} if donationId does not exist
   */
  getDonation(donationId: string): Donation {
    let donation;
    Object.values(this.campaignData).forEach((campaignDatum) => {
      const campaignDonation = campaignDatum.donations[donationId];
      if (campaignDatum) {
        donation = campaignDonation;
      }
    });
    if (!donation) {
      throw new Error("donation doesn't exist");
    }
    return donation;
  }

  /**
   * Return aggregate stats for a campaign, excluding refunded donations.
   * goalPercent = Math.round((totalRaised / goal) * 1000) / 10
   *
   * @throws {Error} if campaignId does not exist
   */
  getCampaignStats(campaignId: string): CampaignStats {
    const existingCampaign = this.campaignData[campaignId];
    if (!existingCampaign) {
      throw new Error("campaign doesn't exist");
    }
    const campaignStats = { totalRaised: 0, donorCount: 0, donationCount: 0 };
    const alreadyDonatedEmails: string[] = [];
    Object.values(existingCampaign.donations).forEach((donation) => {
      if (!donation.refunded) {
        campaignStats.totalRaised += donation.amount;
        campaignStats.donationCount += 1;
        if (!alreadyDonatedEmails.includes(donation.donorEmail)) {
          campaignStats.donorCount += 1;
          alreadyDonatedEmails.push(donation.donorEmail);
        }
      }
    });
    return {
      ...campaignStats,
      goalPercent:
        Math.round((campaignStats.totalRaised / existingCampaign.goal) * 1000) /
        10,
    };
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Return the top `limit` donors for a campaign, ranked by total non-refunded
   * amount (descending). Aggregate multiple donations from the same email.
   * On a tie in totalAmount, sort by donorEmail ascending.
   *
   * @throws {Error} if campaignId does not exist
   */
  getTopDonors(campaignId: string, limit: number): DonorSummary[] {
    const existingCampaign = this.campaignData[campaignId];
    if (!existingCampaign) {
      throw new Error("campaign doesn't exist");
    }
    const donorEmailToAmountMap: Record<string, DonorSummary> = {};
    Object.values(existingCampaign.donations).forEach((donation) => {
      if (!donation.refunded) {
        let existingDonationSummary: DonorSummary | undefined =
          donorEmailToAmountMap[donation.donorEmail];
        if (existingDonationSummary) {
          existingDonationSummary.totalAmount += donation.amount;
          existingDonationSummary.donationCount += 1;
        } else {
          existingDonationSummary = {
            donorEmail: donation.donorEmail,
            donorName: donation.donorName,
            totalAmount: donation.amount,
            donationCount: 1,
          };
        }
        donorEmailToAmountMap[donation.donorEmail] = existingDonationSummary;
      }
    });
    const donationSortFn = (
      donationSummaryA: DonorSummary,
      donationSummaryB: DonorSummary,
    ) => {
      if (donationSummaryA.totalAmount < donationSummaryB.totalAmount) {
        return 1;
      } else if (donationSummaryA.totalAmount > donationSummaryB.totalAmount) {
        return -1;
      } else {
        if (donationSummaryA.donorEmail < donationSummaryB.donorEmail) {
          return 1;
        } else {
          return -1;
        }
      }
    };
    const sortedDoantionSummarys = [
      ...Object.values(donorEmailToAmountMap),
    ].sort(donationSortFn);
    return sortedDoantionSummarys.slice(0, limit);
  }

  /**
   * Return all non-refunded donations for a campaign whose timestamp falls
   * within [from, to] (inclusive), sorted by timestamp ascending.
   *
   * @throws {Error} if campaignId does not exist
   */
  getDonationsInRange(
    campaignId: string,
    from: string,
    to: string,
  ): Donation[] {
    const existingCampaign = this.campaignData[campaignId];
    if (!existingCampaign) {
      throw new Error("campaign doesn't exist");
    }
    const fromDate = new Date(from);
    const toDate = new Date(to);
    const donationsInRanage = Object.values(existingCampaign.donations).filter(
      (donation) => {
        const donationDate = new Date(donation.timestamp);
        return donationDate <= toDate && donationDate >= fromDate;
      },
    );
    return [...donationsInRanage].sort(
      (donationA: Donation, donationB: Donation) =>
        new Date(donationA.timestamp).getTime() -
        new Date(donationB.timestamp).getTime(),
    );
  }

  // ── Part 3 ───────────────────────────────────────────────────────────────

  /**
   * Mark a donation as refunded.
   * Refunded donations are excluded from stats and leaderboards.
   *
   * @throws {Error} if donationId does not exist
   * @throws {Error} if the donation is already refunded
   */
  refundDonation(donationId: string): Donation {
    throw new Error("Not implemented");
  }

  /**
   * Create a matched donation in the same campaign as `sourceDonationId`.
   *
   * matchedAmount = Math.min(sourceDonation.amount * matchRatio, maxAmount)
   * Matched donations appear in stats and leaderboards like any other donation.
   * The source donation must not be refunded.
   *
   * @param sourceDonationId  the donation being matched
   * @param matcherName       name of the matching donor
   * @param matcherEmail      email of the matching donor
   * @param matchRatio        fraction of source amount to match (e.g. 0.5 = 50%)
   * @param maxAmount         ceiling on the matched amount in dollars
   * @throws {Error} if sourceDonationId does not exist
   * @throws {Error} if source donation is refunded
   */
  addMatchingGift(
    sourceDonationId: string,
    matcherName: string,
    matcherEmail: string,
    matchRatio: number,
    maxAmount: number,
  ): Donation {
    throw new Error("Not implemented");
  }
}
