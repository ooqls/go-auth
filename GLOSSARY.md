# Authentication and authorization

The authority for authenticating users and deciding which actions they may perform.

## Language

**Permission binding**:
An association granting a permission to a role; it is distinct from assigning a role to a user.
_Avoid_: User permission, role binding

**Role binding**:
An association assigning a role to a user.
_Avoid_: Permission binding

**Hierarchy**:
A role's relative authority, with larger values representing greater authority; a user's hierarchy is the highest among their assigned roles.
_Avoid_: Priority, strict superiority

**Permission-management authority**:
Authority to assign or unassign permissions on eligible roles, independent of personally possessing the permissions being delegated.
_Avoid_: Permission ownership
