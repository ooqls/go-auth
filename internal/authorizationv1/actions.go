package authorizationv1

type Action = string

const (
	CreateAction    Action = "create"
	ReadAction      Action = "read"
	UpdateAction    Action = "update"
	DeleteAction    Action = "delete"
	AssignAction    Action = "assign"
	UnassignAction  Action = "unassign"
	UpdateKeyAction Action = "update-key"
)
