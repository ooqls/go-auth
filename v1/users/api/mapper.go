package usersapi

import (
	"github.com/ooqls/go-auth/v1/gen"
	"github.com/ooqls/go-auth/v1/users"
)

// toGenUser translates a domain User into the generated OpenAPI model. All
// users handlers must serialize through this function so the domain -> API
// mapping lives in exactly one place. It also keeps sensitive domain fields
// (key, salt) out of API responses.
func toGenUser(u users.User) gen.User {
	id := u.Id
	return gen.User{
		Id:        &id,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// toGenUserList translates a slice of domain users into generated OpenAPI
// models.
func toGenUserList(list []users.User) []gen.User {
	items := make([]gen.User, 0, len(list))
	for _, u := range list {
		items = append(items, toGenUser(u))
	}
	return items
}
