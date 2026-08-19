/**
 * =============================================================================
 * INTERVIEW PROBLEM 03: Donor Communication Suppressor
 * Difficulty: Senior Software Engineer | Estimated time: 45–60 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * You're building the donor suppression engine for a school fundraising
 * platform. Before each email campaign goes out, the platform filters the
 * send list to avoid contacting donors who have already given, recently gave,
 * or have opted out of communications. Getting this wrong wastes goodwill —
 * double-soliciting recent donors is a common and costly mistake.
 *
 * Store all state in instance variables initialised in the constructor.
 * You choose the internal data structures; the public interface is what matters.
 *
 * SUPPRESSION RULES (applied in priority order)
 * -----------------------------------------------
 * 1. opted_out       — donor explicitly unsubscribed
 * 2. already_donated — donor has a donation recorded for this specific campaignId
 * 3. recency_cooldown — donor donated to ANY campaign within the last `cooldownDays` days
 *                       relative to `asOf` (i.e. donatedAt >= asOf minus cooldownDays)
 *
 * A donor is suppressed if ANY rule applies. Return the highest-priority reason.
 *
 * // Example
 * // const s = new DonorSuppressor()
 * // s.optOut('alice@example.com')
 * // s.recordDonation('bob@example.com', 'camp-fall', '2025-08-01T10:00:00')
 * // s.checkSuppression('alice@example.com', 'camp-fall', '2025-08-20T00:00:00')
 * // → { email: 'alice@example.com', suppressed: true, reason: 'opted_out' }
 * // s.checkSuppression('bob@example.com', 'camp-fall', '2025-08-20T00:00:00')
 * // → { email: 'bob@example.com', suppressed: true, reason: 'already_donated' }
 * // s.checkSuppression('carol@example.com', 'camp-fall', '2025-08-20T00:00:00')
 * // → { email: 'carol@example.com', suppressed: false, reason: null }
 *
 * =============================================================================
 * PART 1 — Opt-out management and basic suppression check
 * =============================================================================
 *
 * Implement optOut, optIn, isOptedOut, and checkSuppression (opted_out rule only).
 */

// ── Types ──────────────────────────────────────────────────────────────────

export type SuppressReason =
  | "opted_out"
  | "already_donated"
  | "recency_cooldown";

export interface SuppressResult {
  email: string;
  suppressed: boolean;
  reason: SuppressReason | null;
}

export interface SuppressOptions {
  cooldownDays?: number; // suppress if donated within last N days (default: no cooldown)
}

export interface BulkFilterResult {
  eligible: string[]; // emails that passed all suppression rules
  suppressed: SuppressResult[]; // full result for each suppressed email
}

export interface SuppressedSummary {
  total: number;
  eligibleCount: number;
  suppressedCount: number;
  byReason: Partial<Record<SuppressReason, number>>;
}

interface Donation {
  email: string;
  donatedAt: Date;
}

// ── Class ──────────────────────────────────────────────────────────────────

export class DonorSuppressor {
  optedOutEmails: Set<string>;
  campaignDonations: Record<string, Record<string, Donation[]>>;
  constructor() {
    this.optedOutEmails = new Set();
    this.campaignDonations = {};
  }

  //   {
  //     campaignId: {email: Donation[]}
  //   }
  // ── Part 1 ───────────────────────────────────────────────────────────────

  /**
   * Add an email to the global opt-out list. Idempotent.
   */
  optOut(email: string): void {
    this.optedOutEmails.add(email);
  }

  /**
   * Remove an email from the opt-out list. Idempotent (no-op if not opted out).
   */
  optIn(email: string): void {
    this.optedOutEmails.delete(email);
  }

  /** Return true if the email is on the opt-out list. */
  isOptedOut(email: string): boolean {
    return this.optedOutEmails.has(email);
  }

  /**
   * Check whether a donor should be suppressed from a campaign send.
   *
   * Part 1: only applies the opted_out rule.
   * Part 2: also applies already_donated and recency_cooldown (via options).
   *
   * @param email       the donor's email
   * @param campaignId  the campaign the email would be sent for
   * @param asOf        ISO-8601 datetime representing "now" for date arithmetic
   * @param options     optional suppression config (Part 2)
   */
  checkSuppression(
    email: string,
    campaignId: string,
    asOf: string,
    options?: SuppressOptions,
  ): SuppressResult {
    //     export interface SuppressResult {
    //   email: string;
    //   suppressed: boolean;
    //   reason: SuppressReason | null;
    // }
    if (this.isOptedOut(email)) {
      return {
        email,
        suppressed: true,
        reason: "opted_out",
      };
    }

    // console.log("his.campaignDonations", this.campaignDonations);
    const donationsByCampaignAndEmail = this.campaignDonations[campaignId];
    // console.log("donationsByCampaignAndEmail", donationsByCampaignAndEmail);
    if (!donationsByCampaignAndEmail) {
      return {
        email,
        suppressed: false,
        reason: null,
      };
    }
    // assume campaign exists for now
    if (Object.keys(donationsByCampaignAndEmail).includes(email)) {
      return {
        email,
        suppressed: true,
        reason: "already_donated",
      };
    }

    const coolDown = options?.cooldownDays;
    // Assume donation array is pre-sorted ascending
    if (coolDown) {
      let hasDonatedWithinCoolDown = false;
      for (const donationRecord of Object.values(this.campaignDonations)) {
        for (const donationPair of Object.entries(donationRecord)) {
          const donationEmail = donationPair[0];
          const donationArr = donationPair[1];
          if (donationEmail !== email) {
            continue;
          }
          let coolDownDate = new Date(asOf);
          coolDownDate.setDate(coolDownDate.getDate() - coolDown);
          if (donationArr[donationArr.length - 1].donatedAt >= coolDownDate) {
            // console.log("donationRecord", donationRecord);
            // console.log("donationArr", donationArr);
            // console.log("coolDownDate", coolDownDate);
            // console.log("asOf", asOf);
            hasDonatedWithinCoolDown = true;
            break;
          }
        }
        if (hasDonatedWithinCoolDown) {
          break;
        }
      }
      console.log("hasDonatedWithinCoolDown", hasDonatedWithinCoolDown);
      if (hasDonatedWithinCoolDown) {
        return {
          email,
          suppressed: true,
          reason: "recency_cooldown",
        };
      }
    }

    //      * 2. already_donated — donor has a donation recorded for this specific campaignId
    //  * 3. recency_cooldown — donor donated to ANY campaign within the last `cooldownDays` days
    //  *                       relative to `asOf` (i.e. donatedAt >= asOf minus cooldownDays)

    return {
      email,
      suppressed: false,
      reason: null,
    };
  }

  // ── Part 2 ───────────────────────────────────────────────────────────────

  /**
   * Record that an email donated to a campaign at a given datetime.
   * Multiple calls for the same email/campaign pair are allowed (multiple
   * donations). The most recent donatedAt is used for recency calculations.
   *
   * @param donatedAt ISO-8601 datetime
   */
  recordDonation(email: string, campaignId: string, donatedAt: string): void {
    // Assume for now that it's called in cronological ordining so we don't need to manually sort
    const existingCampaignDonations = this.campaignDonations[campaignId];
    const newDonation: Donation = {
      email,
      donatedAt: new Date(donatedAt),
    };
    if (existingCampaignDonations) {
      const existingEmailDonations = existingCampaignDonations[email];
      if (existingEmailDonations) {
        existingEmailDonations.push(newDonation);
      } else {
        existingCampaignDonations[email] = [newDonation];
      }
    } else {
      this.campaignDonations[campaignId] = { [email]: [newDonation] };
    }
  }

  // Note: checkSuppression (above) is extended in Part 2 to use recordDonation
  // data when options.cooldownDays is provided.

  // ── Part 3 ───────────────────────────────────────────────────────────────

  /**
   * Filter a list of emails into eligible and suppressed groups.
   * Calls checkSuppression for each email — do not duplicate its logic.
   *
   * @param emails      array of donor emails to evaluate
   * @param campaignId  the campaign this send is for
   * @param asOf        ISO-8601 datetime representing "now"
   * @param options     forwarded to checkSuppression
   */
  filterCampaignList(
    emails: string[],
    campaignId: string,
    asOf: string,
    options?: SuppressOptions,
  ): BulkFilterResult {
    throw new Error("Not implemented");
  }

  /**
   * Return a summary of suppression reasons across a list of emails.
   * Calls filterCampaignList internally — do not duplicate its logic.
   */
  getSuppressedSummary(
    emails: string[],
    campaignId: string,
    asOf: string,
    options?: SuppressOptions,
  ): SuppressedSummary {
    throw new Error("Not implemented");
  }
}
