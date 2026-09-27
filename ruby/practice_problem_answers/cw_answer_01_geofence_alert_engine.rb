class GeofenceTracker
  def initialize
    @zones = {}
    @assets = {}
    @alert_rules = []
    @alert_log = []
  end

  def in_zone?(asset, zone)
    return false if asset[:lat].nil? || asset[:lng].nil?

    bounds = zone[:bounds]
    asset[:lat].between?(bounds[:min_lat], bounds[:max_lat]) &&
      asset[:lng].between?(bounds[:min_lng], bounds[:max_lng])
  end

  def current_zone_id(asset_id)
    asset = @assets[asset_id]
    return nil if asset.nil?

    zone = @zones.values.find { |z| in_zone?(asset, z) }
    zone && zone[:id]
  end

  def process_location_update(asset_id, lat, lng, timestamp)
    asset = fetch_asset(asset_id)
    old_zone_id = asset[:zone_id]

    asset[:lat] = lat
    asset[:lng] = lng
    new_zone_id = current_zone_id(asset_id)
    asset[:zone_id] = new_zone_id

    return [] if old_zone_id == new_zone_id

    triggered = @alert_rules.select do |rule|
      (rule[:asset_id].nil? || rule[:asset_id] == asset_id) &&
        (rule[:from_zone_id].nil? || rule[:from_zone_id] == old_zone_id) &&
        (rule[:to_zone_id].nil? || rule[:to_zone_id] == new_zone_id)
    end.map do |rule|
      {
        rule_id: rule[:id],
        asset_id: asset_id,
        from_zone_id: old_zone_id,
        to_zone_id: new_zone_id,
        timestamp: timestamp,
      }
    end

    @alert_log.concat(triggered)
    triggered
  end

  def add_zone(zone_id, name, min_lat, max_lat, min_lng, max_lng)
    raise ArgumentError, "duplicate zone_id" if @zones.key?(zone_id)

    zone = {
      id: zone_id,
      name: name,
      bounds: { min_lat: min_lat, max_lat: max_lat, min_lng: min_lng, max_lng: max_lng },
    }
    @zones[zone_id] = zone
  end

  def remove_zone(zone_id)
    raise KeyError, "zone not found: #{zone_id}" unless @zones.key?(zone_id)

    @zones.delete(zone_id)
    @assets.each_value { |a| a[:zone_id] = nil if a[:zone_id] == zone_id }
  end

  def add_asset(asset_id, name)
    raise ArgumentError, "duplicate asset_id" if @assets.key?(asset_id)

    asset = { id: asset_id, name: name, lat: nil, lng: nil, zone_id: nil }
    @assets[asset_id] = asset
  end

  def add_alert_rule(rule_id, from_zone_id, to_zone_id, asset_id)
    raise ArgumentError, "duplicate rule_id" if @alert_rules.any? { |r| r[:id] == rule_id }

    rule = { id: rule_id, from_zone_id: from_zone_id, to_zone_id: to_zone_id, asset_id: asset_id }
    @alert_rules << rule
    rule
  end

  private

  def fetch_asset(asset_id)
    @assets.fetch(asset_id) { raise KeyError, "asset not found: #{asset_id}" }
  end
end
