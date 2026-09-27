class ApiGateway
  DEFAULT_PLANS = {
    "free" => { rpm: 60, rpd: 1_000 },
    "starter" => { rpm: 300, rpd: 25_000 },
    "pro" => { rpm: 1_000, rpd: 200_000 },
    "enterprise" => { rpm: nil, rpd: nil },
  }.freeze

  def initialize(plans = nil)
    @plans = plans || DEFAULT_PLANS
    @keys = {}
  end

  def create_key(key_id, owner, plan)
    raise ArgumentError, "duplicate key_id" if @keys.key?(key_id)
    raise ArgumentError, "unknown plan" unless @plans.key?(plan)

    @keys[key_id] = {
      id: key_id,
      owner: owner,
      plan: plan,
      enabled: true,
      request_log: [],
    }
  end

  def revoke_key(key_id)
    fetch_key(key_id)[:enabled] = false
  end

  def update_plan(key_id, new_plan)
    raise ArgumentError, "unknown plan" unless @plans.key?(new_plan)

    fetch_key(key_id)[:plan] = new_plan
  end

  def count_in_window(request_log, now, window_seconds)
    request_log.count { |t| t > now - window_seconds && t <= now }
  end

  def is_allowed(key_id, now)
    key = @keys[key_id]
    return false if key.nil? || !key[:enabled]

    limits = @plans[key[:plan]]
    return false if limits[:rpm] && count_in_window(key[:request_log], now, 60) >= limits[:rpm]
    return false if limits[:rpd] && count_in_window(key[:request_log], now, 86_400) >= limits[:rpd]

    true
  end

  def record_request(key_id, now)
    key = fetch_key(key_id)
    key[:request_log] << now
    key[:request_log].select! { |t| t > now - 90_000 }
  end

  def handle_request(key_id, now)
    key = @keys[key_id]
    return { allowed: false, reason: :key_not_found } if key.nil?
    return { allowed: false, reason: :key_disabled } unless key[:enabled]

    limits = @plans[key[:plan]]
    if limits[:rpm] && count_in_window(key[:request_log], now, 60) >= limits[:rpm]
      return { allowed: false, reason: :rpm_exceeded }
    end
    if limits[:rpd] && count_in_window(key[:request_log], now, 86_400) >= limits[:rpd]
      return { allowed: false, reason: :rpd_exceeded }
    end

    record_request(key_id, now)
    { allowed: true, key_id: key_id }
  end

  def get_usage(key_id, now)
    key = fetch_key(key_id)
    limits = @plans[key[:plan]]
    {
      key_id: key_id,
      plan: key[:plan],
      rpm_used: count_in_window(key[:request_log], now, 60),
      rpm_limit: limits[:rpm],
      rpd_used: count_in_window(key[:request_log], now, 86_400),
      rpd_limit: limits[:rpd],
    }
  end

  private

  def fetch_key(key_id)
    @keys.fetch(key_id) { raise KeyError, "key not found: #{key_id}" }
  end
end
