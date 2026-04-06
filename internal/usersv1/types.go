package usersv1

import (
	"github.com/google/uuid"
	"github.com/ooqls/go-auth/internal/corev1"
	"github.com/ooqls/go-auth/internal/usersv1/datagen"
)

var (
	Metadata = corev1.Metadata{
		Group: corev1.Group,
		Kind:  "User",
	}
)

type User struct {
	corev1.Object
	Username string `json:"username"`
	Email    string `json:"email"`
	Key      []byte `json:"key"`
	Salt     []byte `json:"salt"`
}

func FromDatagenUser(user datagen.Usersv1) User {
	return User{
		Object: corev1.Object{
			Metadata: Metadata,
			Id:       user.ID,
			Name:     user.Username,
		},
		Username: user.Username,
		Email:    user.Email,
		Key:      user.Key,
		Salt:     user.Salt,
	}
}

func NewUser(id uuid.UUID, username, email, key, salt string) User {
	return User{
		Object: corev1.Object{
			Metadata: Metadata,
			Id:       id,
			Name:     username,
		},
		Username: username,
		Email:    email,
		Key:      []byte(key),
		Salt:     []byte(salt),
	}
}
