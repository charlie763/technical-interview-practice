class CapacityError < StandardError; end

class CareTeamManager
  def initialize
    @members = {}
    @history = {}
  end

  def add_member(member_id, role, max_patients)
    @members[member_id] = { role: role, max_patients: max_patients }
  end

  def assign(patient_id, member_id, assigned_at)
    member = fetch_member(member_id)
    entries = (@history[patient_id] ||= [])
    current_entry = entries.find { |e| e[:role] == member[:role] && e[:unassigned_at].nil? }

    unless current_entry && current_entry[:member_id] == member_id
      raise CapacityError, "member at capacity: #{member_id}" if get_patients(member_id).size >= member[:max_patients]
    end

    current_entry[:unassigned_at] = assigned_at if current_entry
    entries << { member_id: member_id, role: member[:role], assigned_at: assigned_at, unassigned_at: nil }
  end

  def get_assignment(patient_id, role)
    entries = @history[patient_id] || []
    active = entries.find { |e| e[:role] == role && e[:unassigned_at].nil? }
    active && active[:member_id]
  end

  def get_patients(member_id)
    fetch_member(member_id)
    @history
      .select { |_patient_id, entries| entries.any? { |e| e[:member_id] == member_id && e[:unassigned_at].nil? } }
      .keys
      .sort
  end

  def available_members(role)
    @members
      .select { |_id, m| m[:role] == role }
      .select { |id, m| get_patients(id).size < m[:max_patients] }
      .keys
      .sort
  end

  def get_history(patient_id, role)
    entries = @history[patient_id] || []
    entries
      .select { |e| e[:role] == role }
      .sort_by { |e| e[:assigned_at] }
      .map { |e| [e[:member_id], e[:assigned_at], e[:unassigned_at]] }
  end

  def get_assignment_at(patient_id, role, timestamp)
    entry = get_history(patient_id, role).find do |_member_id, assigned_at, unassigned_at|
      assigned_at <= timestamp && (unassigned_at.nil? || timestamp < unassigned_at)
    end
    entry && entry[0]
  end

  private

  def fetch_member(member_id)
    @members.fetch(member_id) { raise KeyError, "member not found: #{member_id}" }
  end
end
