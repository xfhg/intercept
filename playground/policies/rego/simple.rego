package rbac

import future.keywords

default allow = false

allow {
    # Get the user's roles
    roles := data.user_roles[input.user]
    # Check if any of the user's roles have permission for the action
    allowed_roles := data.action_permissions[input.action]
    count([role | role = roles[_]; role = allowed_roles[_]]) > 0
}

violations[msg] {
    not allow
    msg := sprintf("User %s is not allowed to perform action %s", [input.user, input.action])
}

# Debug rules (optional, can be removed in production)
debug_user_roles = data.user_roles
debug_action_permissions = data.action_permissions
debug_input = input
debug_user_roles_for_input = data.user_roles[input.user]
debug_allowed_roles_for_action = data.action_permissions[input.action]
debug_role_check = [role |
    role = data.user_roles[input.user][_]
    role = data.action_permissions[input.action][_]
]
debug_allow = count(debug_role_check) > 0