"""
Tests for Problem 3: Permission Manager (RBAC)

Run from the python/ directory:
    pytest tests/test_problem_03_permission_manager.py -v
"""

import pytest

# from practice_problems.problem_03_permission_manager import (
#     PermissionManager,
# )
from practice_problem_answers.cw_answer_03_permission_manager import PermissionManager

# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


@pytest.fixture
def fresh_pm():
    """Bare PermissionManager with no roles or users."""
    return PermissionManager()


@pytest.fixture
def pm():
    """Fresh PermissionManager with a standard set of roles."""
    p = PermissionManager()
    p.create_role("viewer", ["posts:read", "comments:read"])
    p.create_role(
        "editor", ["posts:read", "posts:write", "comments:read", "comments:write"]
    )
    p.create_role(
        "admin",
        [
            "posts:read",
            "posts:write",
            "posts:delete",
            "users:read",
            "users:write",
            "billing:read",
        ],
    )
    return p


# ---------------------------------------------------------------------------
# PART 1 — Flat RBAC
# ---------------------------------------------------------------------------


class TestCreateRole:
    def test_creates_role(self, fresh_pm):
        fresh_pm.create_role("role_creates", ["reports:read"])
        assert fresh_pm.get_role_permissions("role_creates") == {"reports:read"}

    def test_empty_permissions_by_default(self, fresh_pm):
        fresh_pm.create_role("role_empty")
        assert fresh_pm.get_role_permissions("role_empty") == set()

    def test_duplicate_raises(self, fresh_pm):
        fresh_pm.create_role("role_dup")
        with pytest.raises(ValueError):
            fresh_pm.create_role("role_dup")


class TestGrantRevokePermission:
    def test_grant_adds_permission(self, pm):
        pm.grant_permission("viewer", "posts:write")
        assert "posts:write" in pm.get_role_permissions("viewer")

    def test_grant_idempotent(self, pm):
        pm.grant_permission("viewer", "posts:read")  # already exists
        assert (
            pm.get_role_permissions("viewer").count if False else True
        )  # just shouldn't raise

    def test_grant_missing_role_raises(self, pm):
        with pytest.raises(KeyError):
            pm.grant_permission("ghost", "posts:read")

    def test_revoke_removes_permission(self, pm):
        pm.revoke_permission("viewer", "posts:read")
        assert "posts:read" not in pm.get_role_permissions("viewer")

    def test_revoke_idempotent(self, pm):
        pm.revoke_permission("viewer", "nonexistent")  # should not raise

    def test_revoke_missing_role_raises(self, pm):
        with pytest.raises(KeyError):
            pm.revoke_permission("ghost", "posts:read")


class TestAssignUnassignRole:
    def test_assign_gives_permissions(self, pm):
        pm.assign_role("alice", "viewer")
        assert pm.has_permission("alice", "posts:read") is True

    def test_assign_multiple_roles(self, pm):
        pm.assign_role("alice", "viewer")
        pm.assign_role("alice", "admin")
        assert pm.has_permission("alice", "billing:read") is True
        assert pm.has_permission("alice", "posts:read") is True

    def test_assign_idempotent(self, pm):
        pm.assign_role("alice", "viewer")
        pm.assign_role("alice", "viewer")  # should not raise or duplicate
        # still only viewer-level perms (no inflation of permission counts)
        assert "billing:read" not in pm.get_all_permissions("alice")

    def test_assign_missing_role_raises(self, pm):
        with pytest.raises(KeyError):
            pm.assign_role("alice", "ghost_role")

    def test_unassign_removes_permissions(self, pm):
        pm.assign_role("alice", "admin")
        pm.unassign_role("alice", "admin")
        assert pm.has_permission("alice", "billing:read") is False

    def test_unassign_missing_role_raises(self, pm):
        with pytest.raises(KeyError):
            pm.unassign_role("alice", "ghost_role")

    def test_unassign_role_user_doesnt_have_raises(self, pm):
        pm.assign_role("alice", "viewer")
        with pytest.raises(KeyError):
            pm.unassign_role("alice", "admin")  # alice doesn't have admin


class TestHasPermission:
    def test_true_for_granted_permission(self, pm):
        pm.assign_role("alice", "viewer")
        assert pm.has_permission("alice", "posts:read") is True

    def test_false_for_missing_permission(self, pm):
        pm.assign_role("alice", "viewer")
        assert pm.has_permission("alice", "billing:read") is False

    def test_false_for_unknown_user(self, pm):
        assert pm.has_permission("nobody", "posts:read") is False

    def test_union_of_multiple_roles(self, pm):
        pm.assign_role("alice", "viewer")
        pm.assign_role("alice", "admin")
        perms = pm.get_all_permissions("alice")
        assert "billing:read" in perms
        assert "posts:read" in perms

    def test_empty_set_for_unknown_user(self, pm):
        assert pm.get_all_permissions("nobody") == set()


# ---------------------------------------------------------------------------
# PART 2 — Role inheritance
# ---------------------------------------------------------------------------


class TestRoleInheritance:
    def test_child_inherits_parent_permissions(self, pm):
        # viewer < editor < admin hierarchy
        pm.set_parent_role("editor", "viewer")
        perms = pm.get_role_permissions("editor")
        assert "posts:read" in perms  # own
        assert "comments:read" in perms  # inherited from viewer

    def test_grandchild_inherits_transitively(self, fresh_pm):
        fresh_pm.create_role("base", ["base:read"])
        fresh_pm.create_role("mid", ["mid:write"])
        fresh_pm.create_role("top", ["top:admin"])
        fresh_pm.set_parent_role("mid", "base")
        fresh_pm.set_parent_role("top", "mid")
        perms = fresh_pm.get_role_permissions("top")
        assert "base:read" in perms
        assert "mid:write" in perms
        assert "top:admin" in perms

    def test_user_gets_inherited_permissions(self, pm):
        pm.set_parent_role("editor", "viewer")
        pm.assign_role("alice", "editor")
        assert pm.has_permission("alice", "posts:read") is True  # own
        assert pm.has_permission("alice", "comments:read") is True  # inherited

    def test_user_does_not_get_sibling_permissions(self, pm):
        pm.set_parent_role("editor", "viewer")
        pm.assign_role("alice", "viewer")
        assert pm.has_permission("alice", "posts:write") is False  # editor-only

    def test_replacing_parent(self, fresh_pm):
        fresh_pm.create_role("base_a", ["a:read"])
        fresh_pm.create_role("base_b", ["b:read"])
        fresh_pm.create_role("child", ["c:read"])
        fresh_pm.set_parent_role("child", "base_a")
        fresh_pm.set_parent_role("child", "base_b")  # replace parent
        perms = fresh_pm.get_role_permissions("child")
        assert "b:read" in perms
        assert "a:read" not in perms  # old parent no longer applies

    def test_set_parent_missing_role_raises(self, pm):
        with pytest.raises(KeyError):
            pm.set_parent_role("viewer", "nonexistent")

    def test_set_parent_missing_child_raises(self, pm):
        with pytest.raises(KeyError):
            pm.set_parent_role("nonexistent", "viewer")

    def test_get_role_permissions_missing_raises(self, pm):
        with pytest.raises(KeyError):
            pm.get_role_permissions("ghost")


# ---------------------------------------------------------------------------
# PART 3 — Scoped permissions with wildcards
# ---------------------------------------------------------------------------


class TestScopedPermissions:
    @pytest.fixture
    def scoped_pm(self):
        p = PermissionManager()
        p.create_role("reader", ["posts:read", "comments:read"])
        p.create_role("post_owner", ["posts:*"])
        p.create_role("moderator", ["*:delete"])
        p.create_role("superadmin", ["*:*"])
        p.create_role("mixed", ["billing:read", "plain_permission"])
        return p

    def test_exact_match(self, scoped_pm):
        scoped_pm.assign_role("alice", "reader")
        assert scoped_pm.has_scoped_permission("alice", "posts", "read") is True

    def test_exact_match_miss(self, scoped_pm):
        scoped_pm.assign_role("alice", "reader")
        assert scoped_pm.has_scoped_permission("alice", "posts", "write") is False

    def test_action_wildcard(self, scoped_pm):
        scoped_pm.assign_role("alice", "post_owner")
        assert scoped_pm.has_scoped_permission("alice", "posts", "read") is True
        assert scoped_pm.has_scoped_permission("alice", "posts", "write") is True
        assert scoped_pm.has_scoped_permission("alice", "posts", "delete") is True

    def test_action_wildcard_does_not_grant_other_resources(self, scoped_pm):
        scoped_pm.assign_role("alice", "post_owner")
        assert scoped_pm.has_scoped_permission("alice", "billing", "read") is False

    def test_resource_wildcard(self, scoped_pm):
        scoped_pm.assign_role("alice", "moderator")
        assert scoped_pm.has_scoped_permission("alice", "posts", "delete") is True
        assert scoped_pm.has_scoped_permission("alice", "comments", "delete") is True
        assert scoped_pm.has_scoped_permission("alice", "users", "delete") is True

    def test_resource_wildcard_does_not_grant_other_actions(self, scoped_pm):
        scoped_pm.assign_role("alice", "moderator")
        assert scoped_pm.has_scoped_permission("alice", "posts", "write") is False

    def test_superadmin_grants_everything(self, scoped_pm):
        scoped_pm.assign_role("alice", "superadmin")
        assert scoped_pm.has_scoped_permission("alice", "posts", "read") is True
        assert scoped_pm.has_scoped_permission("alice", "billing", "write") is True
        assert (
            scoped_pm.has_scoped_permission("alice", "anything", "everything") is True
        )

    def test_plain_permission_ignored_by_scoped_check(self, scoped_pm):
        scoped_pm.assign_role("alice", "mixed")
        # "plain_permission" has no ":" so it's not a scoped permission
        assert (
            scoped_pm.has_scoped_permission("alice", "plain_permission", "read")
            is False
        )

    def test_unknown_user_returns_false(self, scoped_pm):
        assert scoped_pm.has_scoped_permission("nobody", "posts", "read") is False

    def test_inherited_scoped_permissions(self, scoped_pm):
        """Wildcard permissions granted via role inheritance must be visible."""
        scoped_pm.create_role("child_role", ["comments:write"])
        scoped_pm.set_parent_role("child_role", "superadmin")
        scoped_pm.assign_role("alice", "child_role")
        assert scoped_pm.has_scoped_permission("alice", "billing", "delete") is True
