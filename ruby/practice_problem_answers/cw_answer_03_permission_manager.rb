require "set"

class PermissionManager
  def initialize
    @roles = {}
    @user_roles = {}
  end

  def create_role(role_id, permissions = nil)
    raise ArgumentError, "duplicate role_id" if @roles.key?(role_id)

    @roles[role_id] = { permissions: Set.new(permissions), parent_id: nil }
  end

  def grant_permission(role_id, permission)
    fetch_role(role_id)[:permissions].add(permission)
  end

  def revoke_permission(role_id, permission)
    fetch_role(role_id)[:permissions].delete(permission)
  end

  def assign_role(user_id, role_id)
    fetch_role(role_id)
    @user_roles[user_id] ||= Set.new
    @user_roles[user_id].add(role_id)
  end

  def unassign_role(user_id, role_id)
    fetch_role(role_id)
    user_role_ids = @user_roles[user_id]
    raise KeyError, "user does not have role: #{role_id}" unless user_role_ids&.include?(role_id)

    user_role_ids.delete(role_id)
  end

  def has_permission(user_id, permission)
    get_all_permissions(user_id).include?(permission)
  end

  def get_all_permissions(user_id)
    role_ids = @user_roles[user_id]
    return Set.new if role_ids.nil?

    role_ids.each_with_object(Set.new) do |role_id, all|
      all.merge(get_role_permissions(role_id))
    end
  end

  def get_role_permissions(role_id)
    role = fetch_role(role_id)
    permissions = role[:permissions].dup
    parent_id = role[:parent_id]
    permissions.merge(get_role_permissions(parent_id)) if parent_id
    permissions
  end

  def set_parent_role(role_id, parent_role_id)
    fetch_role(parent_role_id)
    fetch_role(role_id)[:parent_id] = parent_role_id
  end

  def has_scoped_permission(user_id, resource, action)
    get_all_permissions(user_id).any? do |permission|
      next false unless permission.include?(":")

      perm_resource, perm_action = permission.split(":", 2)
      (perm_resource == resource || perm_resource == "*") &&
        (perm_action == action || perm_action == "*")
    end
  end

  private

  def fetch_role(role_id)
    @roles.fetch(role_id) { raise KeyError, "role not found: #{role_id}" }
  end
end
