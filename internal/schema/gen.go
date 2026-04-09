package schema

import _ "embed"

//go:embed 20260114144208_auth.sql
var Migration20260114144208Auth string

//go:embed 20260309021330_user_highest_role_view.sql
var Migration20260309021330UserHighestRoleView string

func GetSchemaMigrations() []string {
	return []string{
		Migration20260114144208Auth,
	}
}

func GetViewMigrations() []string {
	return []string{
		Migration20260309021330UserHighestRoleView,
	}
}
