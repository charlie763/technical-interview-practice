# =============================================================================
# INTERVIEW PROBLEM 3: Permission Manager (RBAC)
# Difficulty: Senior Software Engineer | Estimated time: 45 min
# =============================================================================
#
# CONTEXT
# -------
# Almost every production application has some form of Role-Based Access
# Control: a SaaS product with admin/member/viewer tiers, a dev tool with
# repo-level permissions, a document platform with edit/comment/view roles.
#
# You're implementing an in-memory RBAC engine from scratch. You choose the
# internal data structures — the class's public interface below is what
# matters. Store all state in instance variables initialized in `initialize`.
# Class-level variables will bleed between tests and between instances —
# avoid them.
#
# HOW IT WORKS
# ------------
#   - Roles hold a set of permission strings (e.g. "billing:read", "users:write").
#   - Users are assigned one or more roles.
#   - A user "has" a permission if any of their roles grant it.
#   - In Part 2, roles can inherit from a parent role (permissions flow downward).
#   - In Part 3, permission strings use "resource:action" format with wildcards.
#
# NOTES
# -----
#   - Users do not need to be pre-registered. Assigning a role to a user_id
#     that hasn't been seen before creates the user implicitly.
#   - You may assume no cycles will be introduced in the role hierarchy.
#   - role_id and user_id are caller-supplied string slugs (e.g. "admin", "alice").
#
# Example
# -------
#   pm = PermissionManager.new
#   pm.create_role("admin", ["users:write", "billing:read"])
#   pm.create_role("viewer", ["posts:read"])
#   pm.assign_role("alice", "admin")
#   pm.assign_role("bob", "viewer")
#
#   pm.has_permission("alice", "billing:read")   # -> true
#   pm.has_permission("alice", "posts:read")     # -> false
#   pm.has_permission("bob", "users:write")      # -> false
#   pm.get_all_permissions("alice")              # -> Set["users:write", "billing:read"]
#   pm.get_role_permissions("admin")             # -> Set["users:write", "billing:read"]
# =============================================================================

require "set"

class PermissionManager
  def initialize
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 1 — Flat role/permission model (~15 min)
  # ---------------------------------------------------------------------------

  # Create a role with an optional initial list of permission strings.
  # Raise ArgumentError if role_id already exists.
  # Permissions default to an empty set if not provided.
  def create_role(role_id, permissions = nil)
    raise NotImplementedError
  end

  # Add a permission string to a role.
  # Raise KeyError if role_id doesn't exist.
  # No-op if the role already has that permission.
  def grant_permission(role_id, permission)
    raise NotImplementedError
  end

  # Remove a permission string from a role.
  # Raise KeyError if role_id doesn't exist.
  # No-op if the permission wasn't on the role.
  def revoke_permission(role_id, permission)
    raise NotImplementedError
  end

  # Assign a role to a user. A user may hold multiple roles.
  # Raise KeyError if role_id doesn't exist.
  # No-op if the user already has that role.
  def assign_role(user_id, role_id)
    raise NotImplementedError
  end

  # Remove a role from a user.
  # Raise KeyError if role_id doesn't exist.
  # Raise KeyError if the user doesn't have that role.
  def unassign_role(user_id, role_id)
    raise NotImplementedError
  end

  # Return true if the user holds the given permission string through any of
  # their assigned roles. Return false if the user doesn't exist or no role
  # grants it.
  #
  # Parts 1 + 2: checks both direct and inherited permissions (once
  # set_parent_role is implemented).
  # Part 3 scoped wildcards are NOT applied here — only exact string match.
  def has_permission(user_id, permission)
    raise NotImplementedError
  end

  # Return the complete set of permission strings available to user_id,
  # across all their roles (and, after Part 2, all ancestor roles).
  # Return an empty set if the user doesn't exist.
  def get_all_permissions(user_id)
    raise NotImplementedError
  end

  # Return the set of permissions directly on a role.
  # Raise KeyError if role_id doesn't exist.
  #
  # Note: in Part 1 this returns only the role's own permissions. After
  # implementing Part 2 (set_parent_role), update this to also include
  # permissions inherited from ancestor roles.
  def get_role_permissions(role_id)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Role inheritance (~15 min)
  # ---------------------------------------------------------------------------

  # Make role_id inherit all permissions from parent_role_id, and
  # transitively from the parent's ancestors.
  #
  # Raise KeyError if either role doesn't exist.
  # A role may have at most one parent; calling this again replaces the
  # existing parent.
  #
  # After implementing this, update has_permission, get_all_permissions, and
  # get_role_permissions to include inherited permissions.
  def set_parent_role(role_id, parent_role_id)
    raise NotImplementedError
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Scoped permissions with wildcards (~15 min)
  # ---------------------------------------------------------------------------

  # Check whether the user has a permission that covers (resource, action).
  #
  # Permission strings are in "resource:action" format. A permission P covers
  # (resource, action) if any of the following match:
  #   - P == "resource:action"   (exact match)
  #   - P == "resource:*"        (wildcard action)
  #   - P == "*:action"          (wildcard resource)
  #   - P == "*:*"               (superadmin — grants everything)
  #
  # Only "resource:action" formatted permissions are evaluated here. Plain
  # strings without ":" are ignored.
  #
  # Inherited permissions (from Part 2) are included in the check.
  # Return false if no matching permission is found or user doesn't exist.
  def has_scoped_permission(user_id, resource, action)
    raise NotImplementedError
  end
end
