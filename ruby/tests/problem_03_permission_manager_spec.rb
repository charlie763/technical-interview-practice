load_problem("03_permission_manager")

# Shared setup used by several describe blocks below.
def seed_standard_roles(pm)
  pm.create_role("viewer", ["posts:read", "comments:read"])
  pm.create_role(
    "editor",
    ["posts:read", "posts:write", "comments:read", "comments:write"]
  )
  pm.create_role(
    "admin",
    [
      "posts:read",
      "posts:write",
      "posts:delete",
      "users:read",
      "users:write",
      "billing:read",
    ]
  )
end

RSpec.describe PermissionManager do
  let(:pm) { described_class.new }

  # ---------------------------------------------------------------------------
  # PART 1 — Flat RBAC
  # ---------------------------------------------------------------------------
  describe "#create_role" do
    it "creates a role with the given permissions" do
      pm.create_role("role_creates", ["reports:read"])
      expect(pm.get_role_permissions("role_creates")).to eq(Set["reports:read"])
    end

    it "defaults to an empty permission set" do
      pm.create_role("role_empty")
      expect(pm.get_role_permissions("role_empty")).to eq(Set.new)
    end

    it "raises ArgumentError on a duplicate role_id" do
      pm.create_role("role_dup")
      expect { pm.create_role("role_dup") }.to raise_error(ArgumentError)
    end
  end

  describe "#grant_permission and #revoke_permission" do
    before { seed_standard_roles(pm) }

    it "adds a permission to a role" do
      pm.grant_permission("viewer", "posts:write")
      expect(pm.get_role_permissions("viewer")).to include("posts:write")
    end

    it "is idempotent when granting an existing permission" do
      expect { pm.grant_permission("viewer", "posts:read") }.not_to raise_error
    end

    it "raises KeyError when granting to a missing role" do
      expect { pm.grant_permission("ghost", "posts:read") }.to raise_error(KeyError)
    end

    it "removes a permission from a role" do
      pm.revoke_permission("viewer", "posts:read")
      expect(pm.get_role_permissions("viewer")).not_to include("posts:read")
    end

    it "is idempotent when revoking a permission the role doesn't have" do
      expect { pm.revoke_permission("viewer", "nonexistent") }.not_to raise_error
    end

    it "raises KeyError when revoking from a missing role" do
      expect { pm.revoke_permission("ghost", "posts:read") }.to raise_error(KeyError)
    end
  end

  describe "#assign_role and #unassign_role" do
    before { seed_standard_roles(pm) }

    it "gives the user the role's permissions" do
      pm.assign_role("part1_alice", "viewer")
      expect(pm.has_permission("part1_alice", "posts:read")).to be true
    end

    it "unions permissions across multiple assigned roles" do
      pm.assign_role("part1_bob", "viewer")
      pm.assign_role("part1_bob", "admin")
      expect(pm.has_permission("part1_bob", "billing:read")).to be true
      expect(pm.has_permission("part1_bob", "posts:read")).to be true
    end

    it "is idempotent when assigning the same role twice" do
      pm.assign_role("part1_carol", "viewer")
      pm.assign_role("part1_carol", "viewer")
      expect(pm.get_all_permissions("part1_carol")).not_to include("billing:read")
    end

    it "raises KeyError when assigning a missing role" do
      expect { pm.assign_role("part1_dave", "ghost_role") }.to raise_error(KeyError)
    end

    it "removes the role's permissions on unassign" do
      pm.assign_role("part1_erin", "admin")
      pm.unassign_role("part1_erin", "admin")
      expect(pm.has_permission("part1_erin", "billing:read")).to be false
    end

    it "raises KeyError when unassigning a missing role" do
      expect { pm.unassign_role("part1_frank", "ghost_role") }.to raise_error(KeyError)
    end

    it "raises KeyError when unassigning a role the user doesn't have" do
      pm.assign_role("part1_grace", "viewer")
      expect { pm.unassign_role("part1_grace", "admin") }.to raise_error(KeyError)
    end
  end

  describe "#has_permission and #get_all_permissions" do
    before { seed_standard_roles(pm) }

    it "returns true for a granted permission" do
      pm.assign_role("part1_helen", "viewer")
      expect(pm.has_permission("part1_helen", "posts:read")).to be true
    end

    it "returns false for a permission none of the user's roles grant" do
      pm.assign_role("part1_ivan", "viewer")
      expect(pm.has_permission("part1_ivan", "billing:read")).to be false
    end

    it "returns false for an unknown user" do
      expect(pm.has_permission("nobody", "posts:read")).to be false
    end

    it "returns the union of permissions across roles" do
      pm.assign_role("part1_jane", "viewer")
      pm.assign_role("part1_jane", "admin")
      perms = pm.get_all_permissions("part1_jane")
      expect(perms).to include("billing:read")
      expect(perms).to include("posts:read")
    end

    it "returns an empty set for an unknown user" do
      expect(pm.get_all_permissions("nobody")).to eq(Set.new)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 2 — Role inheritance
  # ---------------------------------------------------------------------------
  describe "role inheritance" do
    before { seed_standard_roles(pm) }

    it "gives a role its parent's permissions" do
      pm.set_parent_role("editor", "viewer")
      perms = pm.get_role_permissions("editor")
      expect(perms).to include("posts:read")     # own
      expect(perms).to include("comments:read")  # inherited from viewer
    end

    it "inherits transitively through a grandparent" do
      pm.create_role("base", ["base:read"])
      pm.create_role("mid", ["mid:write"])
      pm.create_role("top", ["top:admin"])
      pm.set_parent_role("mid", "base")
      pm.set_parent_role("top", "mid")
      perms = pm.get_role_permissions("top")
      expect(perms).to include("base:read", "mid:write", "top:admin")
    end

    it "gives an assigned user the inherited permissions" do
      pm.set_parent_role("editor", "viewer")
      pm.assign_role("part2_alice", "editor")
      expect(pm.has_permission("part2_alice", "posts:read")).to be true     # own
      expect(pm.has_permission("part2_alice", "comments:read")).to be true  # inherited
    end

    it "does not leak a child role's permissions to a sibling" do
      pm.set_parent_role("editor", "viewer")
      pm.assign_role("part2_bob", "viewer")
      expect(pm.has_permission("part2_bob", "posts:write")).to be false  # editor-only
    end

    it "replaces the existing parent when called again" do
      pm.create_role("base_a", ["a:read"])
      pm.create_role("base_b", ["b:read"])
      pm.create_role("child", ["c:read"])
      pm.set_parent_role("child", "base_a")
      pm.set_parent_role("child", "base_b")
      perms = pm.get_role_permissions("child")
      expect(perms).to include("b:read")
      expect(perms).not_to include("a:read")
    end

    it "raises KeyError when the parent role doesn't exist" do
      expect { pm.set_parent_role("viewer", "nonexistent") }.to raise_error(KeyError)
    end

    it "raises KeyError when the child role doesn't exist" do
      expect { pm.set_parent_role("nonexistent", "viewer") }.to raise_error(KeyError)
    end

    it "raises KeyError from get_role_permissions on a missing role" do
      expect { pm.get_role_permissions("ghost") }.to raise_error(KeyError)
    end
  end

  # ---------------------------------------------------------------------------
  # PART 3 — Scoped permissions with wildcards
  # ---------------------------------------------------------------------------
  describe "#has_scoped_permission" do
    let(:pm) do
      p = described_class.new
      p.create_role("reader", ["posts:read", "comments:read"])
      p.create_role("post_owner", ["posts:*"])
      p.create_role("moderator", ["*:delete"])
      p.create_role("superadmin", ["*:*"])
      p.create_role("mixed", ["billing:read", "plain_permission"])
      p
    end

    it "matches an exact resource:action permission" do
      pm.assign_role("part3_alice", "reader")
      expect(pm.has_scoped_permission("part3_alice", "posts", "read")).to be true
    end

    it "does not match a different action" do
      pm.assign_role("part3_alice", "reader")
      expect(pm.has_scoped_permission("part3_alice", "posts", "write")).to be false
    end

    it "matches an action wildcard for the same resource" do
      pm.assign_role("part3_bob", "post_owner")
      expect(pm.has_scoped_permission("part3_bob", "posts", "read")).to be true
      expect(pm.has_scoped_permission("part3_bob", "posts", "write")).to be true
      expect(pm.has_scoped_permission("part3_bob", "posts", "delete")).to be true
    end

    it "does not let an action wildcard grant other resources" do
      pm.assign_role("part3_bob", "post_owner")
      expect(pm.has_scoped_permission("part3_bob", "billing", "read")).to be false
    end

    it "matches a resource wildcard for the same action" do
      pm.assign_role("part3_carol", "moderator")
      expect(pm.has_scoped_permission("part3_carol", "posts", "delete")).to be true
      expect(pm.has_scoped_permission("part3_carol", "comments", "delete")).to be true
      expect(pm.has_scoped_permission("part3_carol", "users", "delete")).to be true
    end

    it "does not let a resource wildcard grant other actions" do
      pm.assign_role("part3_carol", "moderator")
      expect(pm.has_scoped_permission("part3_carol", "posts", "write")).to be false
    end

    it "lets *:* grant everything" do
      pm.assign_role("part3_dave", "superadmin")
      expect(pm.has_scoped_permission("part3_dave", "posts", "read")).to be true
      expect(pm.has_scoped_permission("part3_dave", "billing", "write")).to be true
      expect(pm.has_scoped_permission("part3_dave", "anything", "everything")).to be true
    end

    it "ignores permission strings without a colon" do
      pm.assign_role("part3_erin", "mixed")
      expect(pm.has_scoped_permission("part3_erin", "plain_permission", "read")).to be false
    end

    it "returns false for an unknown user" do
      expect(pm.has_scoped_permission("nobody", "posts", "read")).to be false
    end

    it "includes wildcard permissions inherited from a parent role" do
      pm.create_role("child_role", ["comments:write"])
      pm.set_parent_role("child_role", "superadmin")
      pm.assign_role("part3_frank", "child_role")
      expect(pm.has_scoped_permission("part3_frank", "billing", "delete")).to be true
    end
  end
end
