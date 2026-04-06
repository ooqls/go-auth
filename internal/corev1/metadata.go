package corev1

type Metadata struct {
	Group string `json:"group"`
	Kind  string `json:"kind"`
}

const (
	Group = "auth"
)

var UsersV1 = Metadata{
	Group: Group,
	Kind:  "user",
}

var RolesV1 = Metadata{
	Group: Group,
	Kind:  "role",
}

var RoleBindingsV1 = Metadata{
	Group: Group,
	Kind:  "role_binding",
}

var ResourcesV1 = Metadata{
	Group: Group,
	Kind:  "resource",
}

var PermissionsV1 = Metadata{
	Group: Group,
	Kind:  "permission",
}

var LoginChallengesV1 = Metadata{
	Group: Group,
	Kind:  "login_challenge",
}

var ChallengeAttemptsV1 = Metadata{
	Group: Group,
	Kind:  "challenge_attempt",
}

var AuditsV1 = Metadata{
	Group: Group,
	Kind:  "audit",
}

var ActionsV1 = Metadata{
	Group: Group,
	Kind:  "action",
}
