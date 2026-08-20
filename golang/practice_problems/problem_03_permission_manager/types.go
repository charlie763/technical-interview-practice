package rbac

import "errors"

// ErrAlreadyExists is returned when creating a resource with a duplicate ID.
var ErrAlreadyExists = errors.New("already exists")

// ErrNotFound is returned when accessing a resource that does not exist.
var ErrNotFound = errors.New("not found")

// PermissionManager is the interface all implementations must satisfy.
//
// Implement it by defining your own struct and a NewPermissionManager constructor.
// Store all state in struct fields — do NOT use package-level variables, as they
// bleed state between instances and between test runs.
//
// Example skeleton:
//
//	type myPM struct {
//	    roles map[string]map[string]bool
//	    users map[string]map[string]bool
//	    // ... add whatever fields you need
//	}
//
//	func NewPermissionManager() PermissionManager {
//	    return &myPM{
//	        roles: make(map[string]map[string]bool),
//	        users: make(map[string]map[string]bool),
//	    }
//	}
type PermissionManager interface {
	// -------------------------------------------------------------------------
	// PART 1 — Flat role/permission model
	// -------------------------------------------------------------------------

	// CreateRole creates a role with an optional initial list of permissions.
	// Returns ErrAlreadyExists if roleID already exists.
	// Permissions default to an empty set if nil or empty.
	CreateRole(roleID string, permissions []string) error

	// GrantPermission adds a permission string to a role.
	// Returns ErrNotFound if roleID doesn't exist.
	// No-op if the role already has that permission.
	GrantPermission(roleID, permission string) error

	// RevokePermission removes a permission string from a role.
	// Returns ErrNotFound if roleID doesn't exist.
	// No-op if the permission wasn't on the role.
	RevokePermission(roleID, permission string) error

	// AssignRole assigns a role to a user. A user may hold multiple roles.
	// Returns ErrNotFound if roleID doesn't exist.
	// No-op if the user already has that role.
	AssignRole(userID, roleID string) error

	// UnassignRole removes a role from a user.
	// Returns ErrNotFound if roleID doesn't exist or the user doesn't have it.
	UnassignRole(userID, roleID string) error

	// HasPermission returns true if the user holds the given permission through
	// any of their assigned roles (and, after Part 2, inherited roles).
	// Returns false if the user doesn't exist or no role grants it.
	HasPermission(userID, permission string) bool

	// GetAllPermissions returns the complete set of permissions available to
	// userID across all their roles (and, after Part 2, all ancestor roles).
	// Returns an empty map if the user doesn't exist.
	GetAllPermissions(userID string) map[string]bool

	// GetRolePermissions returns the set of permissions on a role, including
	// inherited permissions (after Part 2).
	// Returns ErrNotFound if roleID doesn't exist.
	GetRolePermissions(roleID string) (map[string]bool, error)

	// -------------------------------------------------------------------------
	// PART 2 — Role inheritance
	// -------------------------------------------------------------------------

	// SetParentRole makes roleID inherit all permissions from parentRoleID
	// (and transitively from the parent's ancestors).
	// Returns ErrNotFound if either role doesn't exist.
	// A role may have at most one parent; calling this again replaces it.
	SetParentRole(roleID, parentRoleID string) error

	// -------------------------------------------------------------------------
	// PART 3 — Scoped permissions with wildcards
	// -------------------------------------------------------------------------

	// HasScopedPermission checks whether the user has a permission that covers
	// (resource, action). Permission strings use "resource:action" format.
	// A permission P covers (resource, action) if any of the following match:
	//   - P == "resource:action"  (exact match)
	//   - P == "resource:*"       (wildcard action)
	//   - P == "*:action"         (wildcard resource)
	//   - P == "*:*"              (superadmin — grants everything)
	// Only "resource:action" formatted permissions are evaluated; strings
	// without ":" are ignored. Inherited permissions are included.
	// Returns false if no matching permission is found or user doesn't exist.
	HasScopedPermission(userID, resource, action string) bool
}
