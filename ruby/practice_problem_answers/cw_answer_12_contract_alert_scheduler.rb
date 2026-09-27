require "date"

class ContractAlertScheduler
  def initialize
    @contracts = {}
    @alert_configs = {}
    @sent_records = []
  end

  def add_contract(contract_id, title, owner_email, expires_on)
    raise ArgumentError, "duplicate contract_id: #{contract_id}" if @contracts.key?(contract_id)

    contract = { contract_id: contract_id, title: title, owner_email: owner_email, expires_on: expires_on }
    @contracts[contract_id] = contract
    contract
  end

  def add_alert_config(config_id, days_before, label)
    raise ArgumentError, "duplicate config_id: #{config_id}" if @alert_configs.key?(config_id)

    config = { config_id: config_id, days_before: days_before, label: label }
    @alert_configs[config_id] = config
    config
  end

  def get_contracts_expiring_between(start_date, end_date)
    @contracts.values
      .select { |c| c[:expires_on] >= start_date && c[:expires_on] <= end_date }
      .sort_by { |c| c[:expires_on] }
  end

  def compute_alert_schedule(contract_id)
    contract = fetch_contract(contract_id)
    expires_on = Date.parse(contract[:expires_on])

    @alert_configs.values
      .map do |cfg|
        {
          config_id: cfg[:config_id],
          label: cfg[:label],
          alert_on: (expires_on - cfg[:days_before]).iso8601,
        }
      end
      .sort_by { |e| e[:alert_on] }
  end

  def get_due_alerts(as_of_date)
    @contracts.values
      .flat_map do |contract|
        compute_alert_schedule(contract[:contract_id])
          .select { |e| e[:alert_on] <= as_of_date }
          .map do |e|
            {
              contract_id: contract[:contract_id],
              config_id: e[:config_id],
              label: e[:label],
              alert_on: e[:alert_on],
              owner_email: contract[:owner_email],
              expires_on: contract[:expires_on],
            }
          end
      end
      .sort_by { |e| [e[:alert_on], e[:contract_id]] }
  end

  def record_alert_sent(contract_id, config_id, sent_on)
    fetch_contract(contract_id)
    fetch_alert_config(config_id)

    record = { contract_id: contract_id, config_id: config_id, sent_on: sent_on }
    @sent_records << record
    record
  end

  def get_upcoming_alerts(contract_id, as_of_date)
    compute_alert_schedule(contract_id)
      .select { |e| e[:alert_on] >= as_of_date }
      .map do |e|
        sent = @sent_records.any? { |r| r[:contract_id] == contract_id && r[:config_id] == e[:config_id] }
        { config_id: e[:config_id], label: e[:label], alert_on: e[:alert_on], sent: sent }
      end
      .sort_by { |e| e[:alert_on] }
  end

  private

  def fetch_contract(contract_id)
    @contracts.fetch(contract_id) { raise KeyError, "contract not found: #{contract_id}" }
  end

  def fetch_alert_config(config_id)
    @alert_configs.fetch(config_id) { raise KeyError, "alert config not found: #{config_id}" }
  end
end
