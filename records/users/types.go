package users

import (
	"github.com/ooqls/go-auth/records"
	"github.com/ooqls/go-auth/records/gen"
)

type User = gen.Authv1User
type UserId = records.UserId // or the correct type from your gen package