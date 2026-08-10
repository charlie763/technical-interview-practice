/**
 * Tests for Problem 3: Permission Manager (RBAC)
 *
 * Run from the typescript/ directory:
 *   npm run test:03
 */

import { describe, expect, it, beforeEach } from "vitest";
import { PermissionManager } from "@problems/problem_03_permission_manager";

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** Bare PermissionManager with no roles or users. */
function freshPm(): PermissionManager {
  return new PermissionManager();
}

/** Fresh PermissionManager with a standard set of roles. */
function makePm(): PermissionManager {
  const p = new PermissionManager();
  p.createRole("viewer", ["posts:read", "comments:read"]);
  p.createRole("editor", ["posts:read", "posts:write", "comments:read", "comments:write"]);
  p.createRole("admin", ["posts:read", "posts:write", "posts:delete", "users:read", "users:write", "billing:read"]);
  return p;
}

// ---------------------------------------------------------------------------
// PART 1 — Flat RBAC
// ---------------------------------------------------------------------------

describe("createRole", () => {
  it("creates role", () => {
    const pm = freshPm();
    pm.createRole("role_creates", ["reports:read"]);
    expect(pm.getRolePermissions("role_creates")).toEqual(new Set(["reports:read"]));
  });

  it("empty permissions by default", () => {
    const pm = freshPm();
    pm.createRole("role_empty");
    expect(pm.getRolePermissions("role_empty")).toEqual(new Set());
  });

  it("duplicate throws", () => {
    const pm = freshPm();
    pm.createRole("role_dup");
    expect(() => pm.createRole("role_dup")).toThrow();
  });
});

describe("grantPermission / revokePermission", () => {
  let pm: PermissionManager;

  beforeEach(() => {
    pm = makePm();
  });

  it("grant adds permission", () => {
    pm.grantPermission("viewer", "posts:write");
    expect(pm.getRolePermissions("viewer").has("posts:write")).toBe(true);
  });

  it("grant is idempotent", () => {
    expect(() => pm.grantPermission("viewer", "posts:read")).not.toThrow(); // already exists
  });

  it("grant missing role throws", () => {
    expect(() => pm.grantPermission("ghost", "posts:read")).toThrow();
  });

  it("revoke removes permission", () => {
    pm.revokePermission("viewer", "posts:read");
    expect(pm.getRolePermissions("viewer").has("posts:read")).toBe(false);
  });

  it("revoke is idempotent", () => {
    expect(() => pm.revokePermission("viewer", "nonexistent")).not.toThrow();
  });

  it("revoke missing role throws", () => {
    expect(() => pm.revokePermission("ghost", "posts:read")).toThrow();
  });
});

describe("assignRole / unassignRole", () => {
  let pm: PermissionManager;

  beforeEach(() => {
    pm = makePm();
  });

  it("assign gives permissions", () => {
    pm.assignRole("alice", "viewer");
    expect(pm.hasPermission("alice", "posts:read")).toBe(true);
  });

  it("assign multiple roles", () => {
    pm.assignRole("alice", "viewer");
    pm.assignRole("alice", "admin");
    expect(pm.hasPermission("alice", "billing:read")).toBe(true);
    expect(pm.hasPermission("alice", "posts:read")).toBe(true);
  });

  it("assign is idempotent", () => {
    pm.assignRole("alice", "viewer");
    pm.assignRole("alice", "viewer"); // should not throw or duplicate
    // still only viewer-level perms (no inflation of permission counts)
    expect(pm.getAllPermissions("alice").has("billing:read")).toBe(false);
  });

  it("assign missing role throws", () => {
    expect(() => pm.assignRole("alice", "ghost_role")).toThrow();
  });

  it("unassign removes permissions", () => {
    pm.assignRole("alice", "admin");
    pm.unassignRole("alice", "admin");
    expect(pm.hasPermission("alice", "billing:read")).toBe(false);
  });

  it("unassign missing role throws", () => {
    expect(() => pm.unassignRole("alice", "ghost_role")).toThrow();
  });

  it("unassign role user doesn't have throws", () => {
    pm.assignRole("alice", "viewer");
    expect(() => pm.unassignRole("alice", "admin")).toThrow(); // alice doesn't have admin
  });
});

describe("hasPermission / getAllPermissions", () => {
  let pm: PermissionManager;

  beforeEach(() => {
    pm = makePm();
  });

  it("true for granted permission", () => {
    pm.assignRole("alice", "viewer");
    expect(pm.hasPermission("alice", "posts:read")).toBe(true);
  });

  it("false for missing permission", () => {
    pm.assignRole("alice", "viewer");
    expect(pm.hasPermission("alice", "billing:read")).toBe(false);
  });

  it("false for unknown user", () => {
    expect(pm.hasPermission("nobody", "posts:read")).toBe(false);
  });

  it("union of multiple roles", () => {
    pm.assignRole("alice", "viewer");
    pm.assignRole("alice", "admin");
    const perms = pm.getAllPermissions("alice");
    expect(perms.has("billing:read")).toBe(true);
    expect(perms.has("posts:read")).toBe(true);
  });

  it("empty set for unknown user", () => {
    expect(pm.getAllPermissions("nobody")).toEqual(new Set());
  });
});

// ---------------------------------------------------------------------------
// PART 2 — Role inheritance
// ---------------------------------------------------------------------------

describe("role inheritance", () => {
  let pm: PermissionManager;

  beforeEach(() => {
    pm = makePm();
  });

  it("child inherits parent permissions", () => {
    // viewer < editor < admin hierarchy
    pm.setParentRole("editor", "viewer");
    const perms = pm.getRolePermissions("editor");
    expect(perms.has("posts:read")).toBe(true); // own
    expect(perms.has("comments:read")).toBe(true); // inherited from viewer
  });

  it("grandchild inherits transitively", () => {
    const p = freshPm();
    p.createRole("base", ["base:read"]);
    p.createRole("mid", ["mid:write"]);
    p.createRole("top", ["top:admin"]);
    p.setParentRole("mid", "base");
    p.setParentRole("top", "mid");
    const perms = p.getRolePermissions("top");
    expect(perms.has("base:read")).toBe(true);
    expect(perms.has("mid:write")).toBe(true);
    expect(perms.has("top:admin")).toBe(true);
  });

  it("user gets inherited permissions", () => {
    pm.setParentRole("editor", "viewer");
    pm.assignRole("alice", "editor");
    expect(pm.hasPermission("alice", "posts:read")).toBe(true); // own
    expect(pm.hasPermission("alice", "comments:read")).toBe(true); // inherited
  });

  it("user does not get sibling permissions", () => {
    pm.setParentRole("editor", "viewer");
    pm.assignRole("alice", "viewer");
    expect(pm.hasPermission("alice", "posts:write")).toBe(false); // editor-only
  });

  it("replacing parent", () => {
    const p = freshPm();
    p.createRole("base_a", ["a:read"]);
    p.createRole("base_b", ["b:read"]);
    p.createRole("child", ["c:read"]);
    p.setParentRole("child", "base_a");
    p.setParentRole("child", "base_b"); // replace parent
    const perms = p.getRolePermissions("child");
    expect(perms.has("b:read")).toBe(true);
    expect(perms.has("a:read")).toBe(false); // old parent no longer applies
  });

  it("set parent missing role throws", () => {
    expect(() => pm.setParentRole("viewer", "nonexistent")).toThrow();
  });

  it("set parent missing child throws", () => {
    expect(() => pm.setParentRole("nonexistent", "viewer")).toThrow();
  });

  it("getRolePermissions missing role throws", () => {
    expect(() => pm.getRolePermissions("ghost")).toThrow();
  });
});

// ---------------------------------------------------------------------------
// PART 3 — Scoped permissions with wildcards
// ---------------------------------------------------------------------------

describe("hasScopedPermission", () => {
  function scopedPm(): PermissionManager {
    const p = new PermissionManager();
    p.createRole("reader", ["posts:read", "comments:read"]);
    p.createRole("post_owner", ["posts:*"]);
    p.createRole("moderator", ["*:delete"]);
    p.createRole("superadmin", ["*:*"]);
    p.createRole("mixed", ["billing:read", "plain_permission"]);
    return p;
  }

  it("exact match", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "reader");
    expect(pm.hasScopedPermission("alice", "posts", "read")).toBe(true);
  });

  it("exact match miss", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "reader");
    expect(pm.hasScopedPermission("alice", "posts", "write")).toBe(false);
  });

  it("action wildcard", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "post_owner");
    expect(pm.hasScopedPermission("alice", "posts", "read")).toBe(true);
    expect(pm.hasScopedPermission("alice", "posts", "write")).toBe(true);
    expect(pm.hasScopedPermission("alice", "posts", "delete")).toBe(true);
  });

  it("action wildcard does not grant other resources", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "post_owner");
    expect(pm.hasScopedPermission("alice", "billing", "read")).toBe(false);
  });

  it("resource wildcard", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "moderator");
    expect(pm.hasScopedPermission("alice", "posts", "delete")).toBe(true);
    expect(pm.hasScopedPermission("alice", "comments", "delete")).toBe(true);
    expect(pm.hasScopedPermission("alice", "users", "delete")).toBe(true);
  });

  it("resource wildcard does not grant other actions", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "moderator");
    expect(pm.hasScopedPermission("alice", "posts", "write")).toBe(false);
  });

  it("superadmin grants everything", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "superadmin");
    expect(pm.hasScopedPermission("alice", "posts", "read")).toBe(true);
    expect(pm.hasScopedPermission("alice", "billing", "write")).toBe(true);
    expect(pm.hasScopedPermission("alice", "anything", "everything")).toBe(true);
  });

  it("plain permission ignored by scoped check", () => {
    const pm = scopedPm();
    pm.assignRole("alice", "mixed");
    // "plain_permission" has no ":" so it's not a scoped permission
    expect(pm.hasScopedPermission("alice", "plain_permission", "read")).toBe(false);
  });

  it("unknown user returns false", () => {
    const pm = scopedPm();
    expect(pm.hasScopedPermission("nobody", "posts", "read")).toBe(false);
  });

  it("inherited scoped permissions (wildcard granted via inheritance must be visible)", () => {
    const pm = scopedPm();
    pm.createRole("child_role", ["comments:write"]);
    pm.setParentRole("child_role", "superadmin");
    pm.assignRole("alice", "child_role");
    expect(pm.hasScopedPermission("alice", "billing", "delete")).toBe(true);
  });
});
