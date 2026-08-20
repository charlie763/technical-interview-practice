// =============================================================================
// INTERVIEW PROBLEM 3: Permission Manager (RBAC)
// Difficulty: Senior Software Engineer | Estimated time: 45 min
// =============================================================================
//
// CONTEXT
// -------
// Almost every production application has some form of Role-Based Access Control:
// a SaaS product with admin/member/viewer tiers, a dev tool with repo-level
// permissions, a document platform with edit/comment/view roles, etc.
//
// You're implementing an in-memory RBAC engine from scratch. The internal data
// structures are your choice — the public interface (PermissionManager) is what
// matters.
//
// HOW IT WORKS
// ------------
//   - Roles hold a set of permission strings (e.g. "billing:read", "users:write").
//   - Users are assigned one or more roles.
//   - A user "has" a permission if any of their roles grant it.
//   - In Part 2, roles can inherit from a parent role (permissions flow downward).
//   - In Part 3, permission strings use "resource:action" format with wildcards.
//
// NOTES
// -----
//   - Users do not need to be pre-registered. Assigning a role to a userID
//     that hasn't been seen before creates the user implicitly.
//   - You may assume no cycles will be introduced in the role hierarchy.
//   - Choose whatever internal data structures you like (maps, slices, etc.).
//   - roleID and userID are caller-supplied string slugs (e.g. "admin", "alice").
//   - All mutable state MUST be stored in struct fields set in your constructor.
//     Package-level variables will bleed state between instances and test runs.
//
// INTERFACE (see types.go)
// -------------------------
// The PermissionManager interface defines all methods you must implement.
//
// EXAMPLE
// -------
//   pm := NewPermissionManager()
//   pm.CreateRole("admin", []string{"users:write", "billing:read"})
//   pm.CreateRole("viewer", []string{"posts:read"})
//   pm.AssignRole("alice", "admin")
//   pm.AssignRole("bob", "viewer")
//
//   pm.HasPermission("alice", "billing:read")   // true
//   pm.HasPermission("alice", "posts:read")     // false
//   pm.HasPermission("bob", "users:write")      // false
//   perms, _ := pm.GetRolePermissions("admin")  // {"users:write": true, "billing:read": true}
// =============================================================================

package rbac

// NewPermissionManager returns a new PermissionManager implementation.
//
// You must define your own struct type that implements the PermissionManager
// interface (see types.go) and return it from this constructor.
//
// All state must be stored in struct fields — do NOT use package-level variables.
func NewPermissionManager() PermissionManager {
	panic("not implemented")
}
