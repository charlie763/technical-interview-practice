require "time"

VALID_TRANSITIONS = {
  "draft" => ["in_review"],
  "in_review" => %w[approved draft],
  "approved" => ["executed"],
  "executed" => ["active"],
  "active" => %w[expiring_soon terminated],
  "expiring_soon" => %w[expired active terminated],
  "expired" => [],
  "terminated" => [],
}.freeze

CONTRACT_TERMINAL_STATES = %w[expired terminated].freeze

class ContractLifecycle
  def initialize
    @contracts = {}
    @audit_trails = {}
  end

  def create_contract(contract_id, title, created_at, actor)
    raise ArgumentError, "duplicate contract_id: #{contract_id}" if @contracts.key?(contract_id)

    contract = { contract_id: contract_id, title: title, state: "draft", created_at: created_at, fields: {} }
    @contracts[contract_id] = contract
    @audit_trails[contract_id] = [
      { contract_id: contract_id, from_state: nil, to_state: "draft", at: created_at, actor: actor },
    ]
    contract
  end

  def set_field(contract_id, key, value)
    contract = fetch_contract(contract_id)
    contract[:fields][key] = value
    contract
  end

  def get_contract(contract_id)
    fetch_contract(contract_id)
  end

  def transition(contract_id, to_state, at, actor)
    contract = fetch_contract(contract_id)
    from_state = contract[:state]
    unless VALID_TRANSITIONS.fetch(from_state, []).include?(to_state)
      raise ArgumentError, "invalid transition: #{from_state} -> #{to_state}"
    end

    contract[:state] = to_state
    @audit_trails[contract_id] << { contract_id: contract_id, from_state: from_state, to_state: to_state, at: at, actor: actor }
    contract
  end

  def get_audit_trail(contract_id)
    fetch_contract(contract_id)
    @audit_trails[contract_id]
  end

  def get_contracts_by_state(state)
    @contracts.values.select { |c| c[:state] == state }.sort_by { |c| c[:contract_id] }
  end

  def bulk_advance(contract_ids, to_state, at, actor)
    succeeded = []
    failed = []

    contract_ids.each do |contract_id|
      transition(contract_id, to_state, at, actor)
      succeeded << contract_id
    rescue KeyError, ArgumentError => e
      failed << { contract_id: contract_id, reason: e.message }
    end

    { succeeded: succeeded, failed: failed }
  end

  def get_lifecycle_metrics
    by_state = @contracts.values.group_by { |c| c[:state] }.transform_values(&:size)
    terminal_count = @contracts.values.count { |c| CONTRACT_TERMINAL_STATES.include?(c[:state]) }

    { total: @contracts.size, by_state: by_state, terminal_count: terminal_count }
  end

  def get_overdue_contracts(as_of)
    @contracts.values
      .reject { |c| CONTRACT_TERMINAL_STATES.include?(c[:state]) }
      .filter_map do |c|
        stuck_since = get_audit_trail(c[:contract_id]).last[:at]
        days_stuck = ((Time.parse(as_of) - Time.parse(stuck_since)) / 86_400).to_i
        next if days_stuck <= 30

        {
          contract_id: c[:contract_id],
          title: c[:title],
          state: c[:state],
          stuck_since: stuck_since,
          days_stuck: days_stuck,
        }
      end
      .sort_by { |e| -e[:days_stuck] }
  end

  private

  def fetch_contract(contract_id)
    @contracts.fetch(contract_id) { raise KeyError, "contract not found: #{contract_id}" }
  end
end
