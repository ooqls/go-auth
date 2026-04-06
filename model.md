# Authorization model

Users can be binded to a role using a role binding
Roles have permissions
roles have a heirarchy to determine the heirarchy of users
Users must have the permission to read/write to the target resource.
Resources are defined as roles, users, objects, and rolebindings
Permissions can be broad such as anything in this group, any resource of this kind, or a singular resource.

All resources have a Read,Update,Delete,Create action
The system must check that the user's role's permissions to ensure they have the right permission to do the action

checks for modifying users will include:
* Ensure they have permission to modify the user
* Ensure they have a higher heirarchy than the target user

Checks for creating/modifying roles:
* Ensure the permissions of the role they are updating/creating are encompassed by their permissions
* Ensure that the role heirarchy is not greater than their own
* Ensure they have permission to modify roles

Checks for creating/modifying role_bindings
* Ensure they have the permission to modify rolebindings
* ensure the user associated with that rolebinding is not a higher hierarchy than the user requesting the change.
* if they change who the rolebinding targets, ensure that the new user is not a higher heirarchy than the user requesting it.

checks for creating/modifying resources:
* Ensure the user has permissions to teh resource.

