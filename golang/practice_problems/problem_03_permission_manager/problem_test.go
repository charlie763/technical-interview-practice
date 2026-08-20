// Tests for Problem 3: Permission Manager (RBAC)
//
// Run the stub (expect all panics):
//
//	cd golang && go test ./practice_problems/problem_03_permission_manager/ -v
//
// Run against your answer via run_tests.sh:
//
//	./run_tests.sh \
//	  -f golang/practice_problem_answers/cw_answer_03_permission_manager.go \
//	  -c go test -v .
package rbac

import (
	"errors"
	"testing"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newPM(t *testing.T) PermissionManager {
	t.Helper()
	return NewPermissionManager()
}

// seededPM returns a PermissionManager with standard roles pre-loaded.
func seededPM(t *testing.T) PermissionManager {
	t.Helper()
	pm := NewPermissionManager()
	mustOK(t, pm.CreateRole("viewer", []string{"posts:read", "comments:read"}))
	mustOK(t, pm.CreateRole("editor", []string{"posts:read", "posts:write", "comments:read", "comments:write"}))
	mustOK(t, pm.CreateRole("admin", []string{"posts:read", "posts:write", "posts:delete", "users:read", "users:write", "billing:read"}))
	return pm
}

func mustOK(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertContains(t *testing.T, set map[string]bool, key string) {
	t.Helper()
	if !set[key] {
		t.Errorf("expected set to contain %q", key)
	}
}

func assertNotContains(t *testing.T, set map[string]bool, key string) {
	t.Helper()
	if set[key] {
		t.Errorf("expected set to NOT contain %q", key)
	}
}

// ---------------------------------------------------------------------------
// PART 1 — Flat RBAC
// ---------------------------------------------------------------------------

func TestCreateRole(t *testing.T) {
	t.Run("creates_role_with_permissions", func(t *testing.T) {
		pm := newPM(t)
		mustOK(t, pm.CreateRole("role_creates", []string{"reports:read"}))
		perms, err := pm.GetRolePermissions("role_creates")
		mustOK(t, err)
		assertContains(t, perms, "reports:read")
	})

	t.Run("empty_permissions_by_default", func(t *testing.T) {
		pm := newPM(t)
		mustOK(t, pm.CreateRole("role_empty", nil))
		perms, err := pm.GetRolePermissions("role_empty")
		mustOK(t, err)
		if len(perms) != 0 {
			t.Errorf("expected empty permissions, got %v", perms)
		}
	})

	t.Run("duplicate_returns_error", func(t *testing.T) {
		pm := newPM(t)
		mustOK(t, pm.CreateRole("role_dup", nil))
		if err := pm.CreateRole("role_dup", nil); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("expected ErrAlreadyExists, got %v", err)
		}
	})
}

func TestGrantRevokePermission(t *testing.T) {
	t.Run("grant_adds_permission", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.GrantPermission("viewer", "posts:write"))
		perms, _ := pm.GetRolePermissions("viewer")
		assertContains(t, perms, "posts:write")
	})

	t.Run("grant_idempotent", func(t *testing.T) {
		pm := seededPM(t)
		// should not return an error even if already present
		if err := pm.GrantPermission("viewer", "posts:read"); err != nil {
			t.Errorf("grant idempotent: unexpected error: %v", err)
		}
	})

	t.Run("grant_missing_role_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.GrantPermission("ghost", "posts:read"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("revoke_removes_permission", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.RevokePermission("viewer", "posts:read"))
		perms, _ := pm.GetRolePermissions("viewer")
		assertNotContains(t, perms, "posts:read")
	})

	t.Run("revoke_idempotent", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.RevokePermission("viewer", "nonexistent"); err != nil {
			t.Errorf("revoke idempotent: unexpected error: %v", err)
		}
	})

	t.Run("revoke_missing_role_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.RevokePermission("ghost", "posts:read"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestAssignUnassignRole(t *testing.T) {
	t.Run("assign_gives_permissions", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.AssignRole("alice", "viewer"))
		if !pm.HasPermission("alice", "posts:read") {
			t.Error("expected alice to have posts:read")
		}
	})

	t.Run("assign_multiple_roles", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.AssignRole("alice", "viewer"))
		mustOK(t, pm.AssignRole("alice", "admin"))
		if !pm.HasPermission("alice", "billing:read") {
			t.Error("expected alice to have billing:read from admin")
		}
		if !pm.HasPermission("alice", "posts:read") {
			t.Error("expected alice to have posts:read from viewer")
		}
	})

	t.Run("assign_idempotent", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.AssignRole("alice", "viewer"))
		mustOK(t, pm.AssignRole("alice", "viewer")) // duplicate, should not error
		perms := pm.GetAllPermissions("alice")
		assertNotContains(t, perms, "billing:read") // no admin perms
	})

	t.Run("assign_missing_role_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.AssignRole("alice", "ghost_role"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unassign_removes_permissions", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.AssignRole("alice", "admin"))
		mustOK(t, pm.UnassignRole("alice", "admin"))
		if pm.HasPermission("alice", "billing:read") {
			t.Error("alice should not have billing:read after unassign")
		}
	})

	t.Run("unassign_missing_role_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.UnassignRole("alice", "ghost_role"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("unassign_role_user_does_not_have_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.AssignRole("alice", "viewer"))
		if err := pm.UnassignRole("alice", "admin"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

func TestHasPermission(t *testing.T) {
	t.Run("true_for_granted_permission", func(t *testing.T) {
		pm := seededPM(t)
		pm.AssignRole("alice", "viewer")
		if !pm.HasPermission("alice", "posts:read") {
			t.Error("expected true")
		}
	})

	t.Run("false_for_missing_permission", func(t *testing.T) {
		pm := seededPM(t)
		pm.AssignRole("alice", "viewer")
		if pm.HasPermission("alice", "billing:read") {
			t.Error("expected false")
		}
	})

	t.Run("false_for_unknown_user", func(t *testing.T) {
		pm := seededPM(t)
		if pm.HasPermission("nobody", "posts:read") {
			t.Error("expected false for unknown user")
		}
	})

	t.Run("union_of_multiple_roles", func(t *testing.T) {
		pm := seededPM(t)
		pm.AssignRole("alice", "viewer")
		pm.AssignRole("alice", "admin")
		perms := pm.GetAllPermissions("alice")
		assertContains(t, perms, "billing:read")
		assertContains(t, perms, "posts:read")
	})

	t.Run("empty_map_for_unknown_user", func(t *testing.T) {
		pm := seededPM(t)
		if perms := pm.GetAllPermissions("nobody"); len(perms) != 0 {
			t.Errorf("expected empty map, got %v", perms)
		}
	})
}

func TestGetRolePermissions(t *testing.T) {
	t.Run("missing_role_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if _, err := pm.GetRolePermissions("ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 2 — Role inheritance
// ---------------------------------------------------------------------------

func TestRoleInheritance(t *testing.T) {
	t.Run("child_inherits_parent_permissions", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.SetParentRole("editor", "viewer"))
		perms, _ := pm.GetRolePermissions("editor")
		assertContains(t, perms, "posts:read")    // own
		assertContains(t, perms, "comments:read") // inherited from viewer
	})

	t.Run("grandchild_inherits_transitively", func(t *testing.T) {
		pm := newPM(t)
		mustOK(t, pm.CreateRole("base", []string{"base:read"}))
		mustOK(t, pm.CreateRole("mid", []string{"mid:write"}))
		mustOK(t, pm.CreateRole("top", []string{"top:admin"}))
		mustOK(t, pm.SetParentRole("mid", "base"))
		mustOK(t, pm.SetParentRole("top", "mid"))
		perms, _ := pm.GetRolePermissions("top")
		assertContains(t, perms, "base:read")
		assertContains(t, perms, "mid:write")
		assertContains(t, perms, "top:admin")
	})

	t.Run("user_gets_inherited_permissions", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.SetParentRole("editor", "viewer"))
		mustOK(t, pm.AssignRole("alice", "editor"))
		if !pm.HasPermission("alice", "posts:read") {
			t.Error("expected alice to have posts:read (own)")
		}
		if !pm.HasPermission("alice", "comments:read") {
			t.Error("expected alice to have comments:read (inherited)")
		}
	})

	t.Run("user_does_not_get_sibling_permissions", func(t *testing.T) {
		pm := seededPM(t)
		mustOK(t, pm.SetParentRole("editor", "viewer"))
		mustOK(t, pm.AssignRole("alice", "viewer"))
		if pm.HasPermission("alice", "posts:write") {
			t.Error("alice (viewer) should not have posts:write (editor only)")
		}
	})

	t.Run("replacing_parent", func(t *testing.T) {
		pm := newPM(t)
		mustOK(t, pm.CreateRole("base_a", []string{"a:read"}))
		mustOK(t, pm.CreateRole("base_b", []string{"b:read"}))
		mustOK(t, pm.CreateRole("child", []string{"c:read"}))
		mustOK(t, pm.SetParentRole("child", "base_a"))
		mustOK(t, pm.SetParentRole("child", "base_b")) // replace parent
		perms, _ := pm.GetRolePermissions("child")
		assertContains(t, perms, "b:read")
		assertNotContains(t, perms, "a:read") // old parent no longer applies
	})

	t.Run("set_parent_missing_parent_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.SetParentRole("viewer", "nonexistent"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("set_parent_missing_child_returns_error", func(t *testing.T) {
		pm := seededPM(t)
		if err := pm.SetParentRole("nonexistent", "viewer"); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// PART 3 — Scoped permissions with wildcards
// ---------------------------------------------------------------------------

func scopedPM(t *testing.T) PermissionManager {
	t.Helper()
	pm := newPM(t)
	mustOK(t, pm.CreateRole("reader", []string{"posts:read", "comments:read"}))
	mustOK(t, pm.CreateRole("post_owner", []string{"posts:*"}))
	mustOK(t, pm.CreateRole("moderator", []string{"*:delete"}))
	mustOK(t, pm.CreateRole("superadmin", []string{"*:*"}))
	mustOK(t, pm.CreateRole("mixed", []string{"billing:read", "plain_permission"}))
	return pm
}

func TestHasScopedPermission(t *testing.T) {
	t.Run("exact_match", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "reader")
		if !pm.HasScopedPermission("alice", "posts", "read") {
			t.Error("expected true for exact match")
		}
	})

	t.Run("exact_match_miss", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "reader")
		if pm.HasScopedPermission("alice", "posts", "write") {
			t.Error("expected false — reader has no posts:write")
		}
	})

	t.Run("action_wildcard", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "post_owner")
		for _, action := range []string{"read", "write", "delete"} {
			if !pm.HasScopedPermission("alice", "posts", action) {
				t.Errorf("expected true for posts:%s via posts:*", action)
			}
		}
	})

	t.Run("action_wildcard_does_not_grant_other_resources", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "post_owner")
		if pm.HasScopedPermission("alice", "billing", "read") {
			t.Error("posts:* should not grant billing:read")
		}
	})

	t.Run("resource_wildcard", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "moderator")
		for _, resource := range []string{"posts", "comments", "users"} {
			if !pm.HasScopedPermission("alice", resource, "delete") {
				t.Errorf("expected true for %s:delete via *:delete", resource)
			}
		}
	})

	t.Run("resource_wildcard_does_not_grant_other_actions", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "moderator")
		if pm.HasScopedPermission("alice", "posts", "write") {
			t.Error("*:delete should not grant posts:write")
		}
	})

	t.Run("superadmin_grants_everything", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "superadmin")
		pairs := [][2]string{{"posts", "read"}, {"billing", "write"}, {"anything", "everything"}}
		for _, p := range pairs {
			if !pm.HasScopedPermission("alice", p[0], p[1]) {
				t.Errorf("superadmin should grant %s:%s", p[0], p[1])
			}
		}
	})

	t.Run("plain_permission_ignored_by_scoped_check", func(t *testing.T) {
		pm := scopedPM(t)
		pm.AssignRole("alice", "mixed")
		if pm.HasScopedPermission("alice", "plain_permission", "read") {
			t.Error("plain strings without ':' should not satisfy scoped check")
		}
	})

	t.Run("unknown_user_returns_false", func(t *testing.T) {
		pm := scopedPM(t)
		if pm.HasScopedPermission("nobody", "posts", "read") {
			t.Error("expected false for unknown user")
		}
	})

	t.Run("inherited_scoped_permissions", func(t *testing.T) {
		pm := scopedPM(t)
		mustOK(t, pm.CreateRole("child_role", []string{"comments:write"}))
		mustOK(t, pm.SetParentRole("child_role", "superadmin"))
		mustOK(t, pm.AssignRole("alice", "child_role"))
		if !pm.HasScopedPermission("alice", "billing", "delete") {
			t.Error("inherited *:* from superadmin should grant billing:delete")
		}
	})
}
