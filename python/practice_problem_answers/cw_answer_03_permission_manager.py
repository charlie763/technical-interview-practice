"""
=============================================================================
INTERVIEW PROBLEM 3: Permission Manager (RBAC)
Difficulty: Senior Software Engineer | Estimated time: 45 min
=============================================================================

CONTEXT
-------
Almost every production application has some form of Role-Based Access Control:
a SaaS product with admin/member/viewer tiers, a dev tool with repo-level
permissions, a document platform with edit/comment/view roles, etc.

You're implementing an in-memory RBAC engine from scratch. The data
structures are up to you — the class's public interface is what matters.

HOW IT WORKS
------------
  - Roles hold a set of permission strings (e.g. "billing:read", "users:write").
  - Users are assigned one or more roles.
  - A user "has" a permission if any of their roles grant it.
  - In Part 2, roles can inherit from a parent role (permissions flow downward).
  - In Part 3, permission strings use "resource:action" format with wildcards.

NOTES
-----
  - Users do not need to be pre-registered. Assigning a role to a user_id
    that hasn't been seen before creates the user implicitly.
  - You may assume no cycles will be introduced in the role hierarchy.
  - Choose whatever internal data structures you like (dicts, sets, etc.).
=============================================================================
"""

# Notes to self
# read through all problem specs, i.e. inheritance, only grabbing roles via user id first -> better long term planning


class Role:
    """
    i.e. something like "admin"
    """

    EXISTING_ROLES_IDS = set()

    def __init__(self, id: int, permissions: set[str], name: str | None = None):
        self.id = id
        self.permissions = permissions
        self.name = name


class User:
    def __init__(self, id: int, name: str = "jean-luc", roles: set[Role] = set()):
        self.id = id
        self.name = name
        self.roles = roles


class UserRole:
    CURRET_USER_ROLE_ID = 1

    def __init__(self, user_id: int, role_id: int):
        self.user_id = user_id
        self.role_id = role_id
        self.id = self.CURRET_USER_ROLE_ID
        self.CURRET_USER_ROLE_ID += 1


class PermissionManager:
    """In-memory Role-Based Access Control (RBAC) engine."""

    VALID_RECORD_TYPES = ["billing", "users"]
    VALID_PERMISSION_TYPES = ["read", "write", "update", "delete"]

    def __init__(self):
        self.users: dict[int, User] = {}
        self.roles: dict[int, Role] = {}
        self.user_roles: dict[int, UserRole] = {}  # int should represent user_id

    def get_role(self, role_id: int):
        role = self.roles.get(role_id)
        if not role:
            raise KeyError(f"role for role id, {role_id}, does not exist")
        return role

    def upsert_user(self, user_id: int):
        user = self.users.get(user_id)
        if not user:
            user = User(id=user_id)
        return user

    # -------------------------------------------------------------------------
    # PART 1 — Flat role/permission model  (~15 min)
    # -------------------------------------------------------------------------

    def create_role(self, role_id: str, permissions: list = None) -> None:
        """
        Create a role with an optional initial list of permission strings.
        Raise ValueError if role_id already exists.
        Permissions default to an empty set if not provided.
        """
        # NOTES for ai:
        #   - more realistic would be to have role id increment
        #   - tests are interdependent in a way that means I can't test this without test writing get_role permissions first
        #   - ai should start with some example data
        #   - put tests in order of functionality being built
        #   - tests shouldn't be susceptible to memory bleed issues
        if role_id in Role.EXISTING_ROLES_IDS:
            raise ValueError(f"role id, {role_id} already exists")
        role_permissions = set(permissions) if permissions else set()
        new_role = Role(id=role_id, permissions=role_permissions)
        self.roles[role_id] = new_role
        Role.EXISTING_ROLES_IDS.add(role_id)
        return new_role

    def grant_permission(self, role_id: str, permission: str) -> None:
        """
        Add a permission string to a role.
        Raise KeyError if role_id doesn't exist.
        No-op if the role already has that permission.
        """
        role = self.get_role(role_id=role_id)
        role.permissions.add(permission)

    def revoke_permission(self, role_id: str, permission: str) -> None:
        """
        Remove a permission string from a role.
        Raise KeyError if role_id doesn't exist.
        No-op if the permission wasn't on the role.
        """
        role = self.get_role(role_id=role_id)
        role.permissions.remove(permission)

    def assign_role(self, user_id: str, role_id: str) -> None:
        """
        Assign a role to a user. A user may hold multiple roles.
        Raise KeyError if role_id doesn't exist.
        No-op if the user already has that role.
        """
        role = self.get_role(role_id=role_id)
        new_user = self.upsert_user(user_id=user_id)
        new_user.roles.add(role)
        self.users[user_id] = new_user

    def unassign_role(self, user_id: str, role_id: str) -> None:
        """
        Remove a role from a user.
        Raise KeyError if role_id doesn't exist.
        Raise KeyError if the user doesn't have that role.
        """
        user = self.users.get(user_id)
        role = self.get_role(role_id=role_id)
        if not role in user.roles:
            raise KeyError("use does not have that role")
        user.roles.remove(role)

    def has_permission(self, user_id: str, permission: str) -> bool:
        """
        Return True if the user holds the given permission string through
        any of their assigned roles.
        Return False if the user doesn't exist or no role grants it.

        Parts 1 + 2: checks both direct and inherited permissions (once
        set_parent_role is implemented).
        Part 3 scoped wildcards are NOT applied here — only exact string match.
        """
        user = self.users.get(user_id)
        has_permission = False
        for role in user.roles:
            if permission in role.permissions:
                has_permission = True
                break
        return has_permission

    def get_all_permissions(self, user_id: str) -> set:
        """
        Return the complete set of permission strings available to user_id,
        across all their roles (and, after Part 2, all ancestor roles).
        Return an empty set if the user doesn't exist.
        """
        all_permissions = set()
        user = self.users.get(user_id)
        if not user:
            return set()
        for role in user.roles:
            for permission in role.permissions:
                all_permissions.add(permission)
        return all_permissions

    # -------------------------------------------------------------------------
    # PART 2 — Role inheritance  (~15 min)
    # -------------------------------------------------------------------------

    def set_parent_role(self, role_id: str, parent_role_id: str) -> None:
        """
        Make role_id inherit all permissions from parent_role_id, and
        transitively from the parent's ancestors.

        Raise KeyError if either role doesn't exist.
        A role may have at most one parent; calling this again replaces the
        existing parent.

        After this is implemented, has_permission and get_all_permissions
        must reflect inherited permissions automatically.
        """
        raise NotImplementedError

    def get_role_permissions(self, role_id: str) -> set:
        """
        Return the full set of permissions for a role, including those
        inherited from ancestor roles.
        Raise KeyError if role_id doesn't exist.
        """
        role = self.roles.get(role_id)
        if not role:
            raise KeyError(f"role for role id, {role_id}, does not exist")
        return role.permissions

    # -------------------------------------------------------------------------
    # PART 3 — Scoped permissions with wildcards  (~15 min)
    # -------------------------------------------------------------------------

    def has_scoped_permission(self, user_id: str, resource: str, action: str) -> bool:
        """
        Check whether the user has a permission that covers (resource, action).

        Permission strings are in "resource:action" format. A permission P
        covers (resource, action) if any of the following match:
          - P == "resource:action"   (exact match)
          - P == "resource:*"        (wildcard action)
          - P == "*:action"          (wildcard resource)
          - P == "*:*"               (superadmin — grants everything)

        Only "resource:action" formatted permissions are evaluated here.
        Plain strings without ":" are ignored.

        Inherited permissions (from Part 2) are included in the check.
        Return False if no matching permission is found or user doesn't exist.
        """
        raise NotImplementedError
