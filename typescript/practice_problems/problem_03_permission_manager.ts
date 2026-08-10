/**
 * =============================================================================
 * INTERVIEW PROBLEM 3: Permission Manager (RBAC)
 * Difficulty: Senior Software Engineer | Estimated time: 45 min
 * =============================================================================
 *
 * CONTEXT
 * -------
 * Almost every production application has some form of Role-Based Access Control:
 * a SaaS product with admin/member/viewer tiers, a dev tool with repo-level
 * permissions, a document platform with edit/comment/view roles, etc.
 *
 * You're implementing an in-memory RBAC engine from scratch. You choose the
 * internal data structures — the public interface (this class's methods) is
 * what matters.
 *
 * HOW IT WORKS
 * ------------
 *   - Roles hold a set of permission strings (e.g. "billing:read", "users:write").
 *   - Users are assigned one or more roles.
 *   - A user "has" a permission if any of their roles grant it.
 *   - In Part 2, roles can inherit from a parent role (permissions flow downward).
 *   - In Part 3, permission strings use "resource:action" format with wildcards.
 *
 * NOTES
 * -----
 *   - Users do not need to be pre-registered. Assigning a role to a userId
 *     that hasn't been seen before creates the user implicitly.
 *   - You may assume no cycles will be introduced in the role hierarchy.
 *   - Choose whatever internal data structures you like (Map, Set, etc.).
 *   - roleId and userId are caller-supplied string slugs (e.g. "admin", "alice").
 *   - Store all state in instance properties initialized in the constructor.
 *     Class-level (static) fields will bleed state between PermissionManager
 *     instances and between test runs — avoid them.
 *
 * # Example
 * const pm = new PermissionManager();
 * pm.createRole("admin", ["users:write", "billing:read"]);
 * pm.createRole("viewer", ["posts:read"]);
 * pm.assignRole("alice", "admin");
 * pm.assignRole("bob", "viewer");
 *
 * pm.hasPermission("alice", "billing:read");   // -> true
 * pm.hasPermission("alice", "posts:read");     // -> false
 * pm.hasPermission("bob", "users:write");      // -> false
 * pm.getAllPermissions("alice");               // -> Set {"users:write", "billing:read"}
 * pm.getRolePermissions("admin");              // -> Set {"users:write", "billing:read"}
 * =============================================================================
 */

export class PermissionManager {
  constructor() {
    // TODO: initialize your internal state here.
    // All state must be instance properties (not static/class-level fields).
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 1 — Flat role/permission model  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Create a role with an optional initial list of permission strings.
   * Throw an Error if roleId already exists.
   * Permissions default to an empty set if not provided.
   */
  createRole(roleId: string, permissions?: string[]): void {
    throw new Error("Not implemented");
  }

  /**
   * Add a permission string to a role.
   * Throw an Error if roleId doesn't exist.
   * No-op if the role already has that permission.
   */
  grantPermission(roleId: string, permission: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Remove a permission string from a role.
   * Throw an Error if roleId doesn't exist.
   * No-op if the permission wasn't on the role.
   */
  revokePermission(roleId: string, permission: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Assign a role to a user. A user may hold multiple roles.
   * Throw an Error if roleId doesn't exist.
   * No-op if the user already has that role.
   */
  assignRole(userId: string, roleId: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Remove a role from a user.
   * Throw an Error if roleId doesn't exist.
   * Throw an Error if the user doesn't have that role.
   */
  unassignRole(userId: string, roleId: string): void {
    throw new Error("Not implemented");
  }

  /**
   * Return true if the user holds the given permission string through
   * any of their assigned roles.
   * Return false if the user doesn't exist or no role grants it.
   *
   * Parts 1 + 2: checks both direct and inherited permissions (once
   * setParentRole is implemented).
   * Part 3 scoped wildcards are NOT applied here — only exact string match.
   */
  hasPermission(userId: string, permission: string): boolean {
    throw new Error("Not implemented");
  }

  /**
   * Return the complete set of permission strings available to userId,
   * across all their roles (and, after Part 2, all ancestor roles).
   * Return an empty set if the user doesn't exist.
   */
  getAllPermissions(userId: string): Set<string> {
    throw new Error("Not implemented");
  }

  /**
   * Return the set of permissions directly on a role.
   * Throw an Error if roleId doesn't exist.
   *
   * Note: in Part 1 this returns only the role's own permissions.
   * After implementing Part 2 (setParentRole), update this to also
   * include permissions inherited from ancestor roles.
   */
  getRolePermissions(roleId: string): Set<string> {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 2 — Role inheritance  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Make roleId inherit all permissions from parentRoleId, and
   * transitively from the parent's ancestors.
   *
   * Throw an Error if either role doesn't exist.
   * A role may have at most one parent; calling this again replaces the
   * existing parent.
   *
   * After implementing this, update hasPermission, getAllPermissions,
   * and getRolePermissions to include inherited permissions.
   */
  setParentRole(roleId: string, parentRoleId: string): void {
    throw new Error("Not implemented");
  }

  // ---------------------------------------------------------------------------
  // PART 3 — Scoped permissions with wildcards  (~15 min)
  // ---------------------------------------------------------------------------

  /**
   * Check whether the user has a permission that covers (resource, action).
   *
   * Permission strings are in "resource:action" format. A permission P
   * covers (resource, action) if any of the following match:
   *   - P === "resource:action"   (exact match)
   *   - P === "resource:*"        (wildcard action)
   *   - P === "*:action"          (wildcard resource)
   *   - P === "*:*"               (superadmin — grants everything)
   *
   * Only "resource:action" formatted permissions are evaluated here.
   * Plain strings without ":" are ignored.
   *
   * Inherited permissions (from Part 2) are included in the check.
   * Return false if no matching permission is found or user doesn't exist.
   */
  hasScopedPermission(userId: string, resource: string, action: string): boolean {
    throw new Error("Not implemented");
  }
}
