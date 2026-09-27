CLAIM_VALID_TRANSITIONS = {
  "filed" => ["investigating"],
  "investigating" => ["evaluation"],
  "evaluation" => %w[settled denied],
  "settled" => ["closed"],
  "denied" => ["closed"],
  "closed" => [],
}.freeze

CLAIM_TERMINAL_STATES = ["closed"].freeze

class ClaimsPipeline
  def initialize
    @claims = {}
  end

  def file_claim(
    claim_id, policy_id:, coverage_type:, incident_date:, filed_at:,
    claimed_amount:, reserve_amount:, actor:
  )
    raise ArgumentError, "duplicate claim_id: #{claim_id}" if @claims.key?(claim_id)
    raise ArgumentError, "claimed_amount must be positive" if claimed_amount <= 0
    raise ArgumentError, "reserve_amount must be positive" if reserve_amount <= 0

    claim = {
      claim_id: claim_id,
      policy_id: policy_id,
      coverage_type: coverage_type,
      incident_date: incident_date,
      filed_at: filed_at,
      status: "filed",
      claimed_amount: claimed_amount,
      reserve_amount: reserve_amount,
      approved_amount: nil,
      events: [],
    }
    claim[:events] << build_event(filed_at, actor, "filed", {})
    @claims[claim_id] = claim
    claim
  end

  def get_claim(claim_id)
    fetch_claim(claim_id)
  end

  def advance_status(claim_id, to_status, at:, actor:)
    claim = fetch_claim(claim_id)
    raise ArgumentError, "use settle_claim/deny_claim for #{to_status}" if %w[settled denied].include?(to_status)

    from_status = claim[:status]
    unless CLAIM_VALID_TRANSITIONS.fetch(from_status, []).include?(to_status)
      raise ArgumentError, "invalid transition: #{from_status} -> #{to_status}"
    end

    claim[:status] = to_status
    claim[:events] << build_event(at, actor, "status_change", { from_status: from_status, to_status: to_status })
    claim
  end

  def update_reserve(claim_id, new_reserve, at:, actor:)
    claim = fetch_claim(claim_id)
    raise ArgumentError, "new_reserve must be positive" if new_reserve <= 0
    raise ArgumentError, "claim is in a terminal state" if CLAIM_TERMINAL_STATES.include?(claim[:status])

    old_reserve = claim[:reserve_amount]
    claim[:reserve_amount] = new_reserve
    claim[:events] << build_event(at, actor, "reserve_update", { old_reserve: old_reserve, new_reserve: new_reserve })
    claim
  end

  def settle_claim(claim_id, approved_amount:, settled_at:, actor:)
    claim = fetch_claim(claim_id)
    raise ArgumentError, "claim is not in evaluation" unless claim[:status] == "evaluation"
    raise ArgumentError, "approved_amount must be positive" if approved_amount <= 0
    raise ArgumentError, "approved_amount exceeds claimed_amount" if approved_amount > claim[:claimed_amount]

    claim[:status] = "settled"
    claim[:approved_amount] = approved_amount
    claim[:events] << build_event(settled_at, actor, "settled", { approved_amount: approved_amount })
    claim
  end

  def deny_claim(claim_id, reason:, denied_at:, actor:)
    claim = fetch_claim(claim_id)
    raise ArgumentError, "claim is not in evaluation" unless claim[:status] == "evaluation"

    claim[:status] = "denied"
    claim[:events] << build_event(denied_at, actor, "denied", { reason: reason })
    claim
  end

  def get_claims_by_policy(policy_id)
    @claims.values.select { |c| c[:policy_id] == policy_id }.sort_by { |c| c[:filed_at] }
  end

  def get_open_claims
    @claims.values
      .reject { |c| CLAIM_TERMINAL_STATES.include?(c[:status]) || c[:status] == "denied" }
      .sort_by { |c| c[:filed_at] }
  end

  def get_reserve_adequacy
    total_reserves = @claims.values.sum { |c| c[:reserve_amount] }
    settled = @claims.values.select { |c| c[:status] == "settled" }
    total_approved = settled.sum { |c| c[:approved_amount] }
    under_reserved = settled.select { |c| c[:reserve_amount] < c[:approved_amount] }

    {
      total_reserves: total_reserves,
      total_approved: total_approved,
      under_reserved_count: under_reserved.size,
      under_reserved_gap: under_reserved.sum { |c| c[:approved_amount] - c[:reserve_amount] },
    }
  end

  def get_claims_metrics
    by_status = @claims.values.group_by { |c| c[:status] }.transform_values(&:size)
    settled = @claims.values.select { |c| c[:status] == "settled" }
    ratios = settled.map { |c| c[:approved_amount].to_f / c[:claimed_amount] }

    {
      total: @claims.size,
      by_status: by_status,
      total_claimed: @claims.values.sum { |c| c[:claimed_amount] },
      total_paid: settled.sum { |c| c[:approved_amount] },
      avg_settlement_ratio: ratios.empty? ? 0.0 : (ratios.sum / ratios.size).round(4),
    }
  end

  def get_policy_loss_history(policy_id)
    claims = get_claims_by_policy(policy_id)
    total_claimed = claims.sum { |c| c[:claimed_amount] }
    total_paid = claims.select { |c| c[:status] == "settled" }.sum { |c| c[:approved_amount] }
    loss_ratio = total_claimed.zero? ? 0.0 : (total_paid.to_f / total_claimed).round(4)

    {
      policy_id: policy_id,
      claim_count: claims.size,
      total_claimed: total_claimed,
      total_paid: total_paid,
      loss_ratio: loss_ratio,
    }
  end

  private

  def build_event(at, actor, action, payload)
    { at: at, actor: actor, action: action, payload: payload }
  end

  def fetch_claim(claim_id)
    @claims.fetch(claim_id) { raise KeyError, "claim not found: #{claim_id}" }
  end
end
