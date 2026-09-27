class ContractAmendmentManager
  def initialize
    @contracts = {}
    @amendments = {}
  end

  def add_contract(contract_id, title, fields)
    raise ArgumentError, "duplicate contract_id: #{contract_id}" if @contracts.key?(contract_id)

    contract = { contract_id: contract_id, title: title, fields: fields.dup }
    @contracts[contract_id] = contract
    @amendments[contract_id] = []
    contract
  end

  def get_base_contract(contract_id)
    fetch_contract(contract_id)
  end

  def add_amendment(amendment_id, contract_id, effective_on, overrides, note)
    fetch_contract(contract_id)
    raise ArgumentError, "duplicate amendment_id: #{amendment_id}" if all_amendment_ids.include?(amendment_id)

    amendment = {
      amendment_id: amendment_id,
      contract_id: contract_id,
      effective_on: effective_on,
      overrides: overrides.dup,
      note: note,
    }
    @amendments[contract_id] << amendment
    amendment
  end

  def get_amendments(contract_id)
    fetch_contract(contract_id)
    @amendments[contract_id].sort_by { |a| [a[:effective_on], a[:amendment_id]] }
  end

  def get_effective_contract(contract_id, as_of_date)
    fields = get_base_contract(contract_id)[:fields].dup
    get_amendments(contract_id)
      .select { |a| a[:effective_on] <= as_of_date }
      .each { |a| fields.merge!(a[:overrides]) }
    fields
  end

  def get_value_history(contract_id, field)
    base = get_base_contract(contract_id)
    entries = []
    entries << { effective_on: "base", value: base[:fields][field], source: "base" } if base[:fields].key?(field)

    get_amendments(contract_id).each do |a|
      next unless a[:overrides].key?(field)

      entries << { effective_on: a[:effective_on], value: a[:overrides][field], source: a[:amendment_id] }
    end

    raise KeyError, "field not found: #{field}" if entries.empty?

    entries
  end

  def get_amendment_summary(contract_id)
    fetch_contract(contract_id)
    amendments = get_amendments(contract_id)
    fields_amended = amendments.flat_map { |a| a[:overrides].keys }.uniq.sort
    latest = amendments.empty? ? nil : amendments.last[:effective_on]

    {
      contract_id: contract_id,
      amendment_count: amendments.size,
      fields_amended: fields_amended,
      latest_amendment: latest,
      current_fields: get_effective_contract(contract_id, latest || "2099-12-31"),
    }
  end

  private

  def fetch_contract(contract_id)
    @contracts.fetch(contract_id) { raise KeyError, "contract not found: #{contract_id}" }
  end

  def all_amendment_ids
    @amendments.values.flatten.map { |a| a[:amendment_id] }
  end
end
