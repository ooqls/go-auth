package userv1

import (
	"context"
)

var _ Repository = &MockUserRepository{}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: map[string]User{},
		pw:    map[string]string{},
	}
}

type MockUserRepository struct {
	users map[string]User
	pw    map[string]string
}

func (m *MockUserRepository) GetUserFromName(ctx context.Context, user string) (*User, error) {
	u, ok := m.users[user]
	if !ok {
		return nil, ErrUserNotFound
	}

	return &u, nil
}

func (m *MockUserRepository) GetUser(ctx context.Context, id Id) (*User, error) {
	for _, u := range m.users {
		if u.UserId == id {
			return &u, nil
		}
	}

	return nil, ErrUserNotFound
}

func (m *MockUserRepository) GetUserName(ctx context.Context, id Id) (string, error) {
	for _, u := range m.users {
		if u.UserId == id {
			return u.Name, nil
		}
	}

	return "", ErrUserNotFound
}

func (m *MockUserRepository) CreateUser(ctx context.Context, email, name, pw string) (Id, error) {
	u := User{
		UserId: Id(len(m.users)),
		Name:   name,
		Email: email,
	}

	m.pw[email] = pw
	m.users[email] = u
	return u.UserId, nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id Id) error {
	return nil
}

func (m *MockUserRepository) Authenticate(ctx context.Context, email string, pw string) (bool, Id, error) {
	return m.pw[email] == pw, 1, nil
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context, page, pagesize int) ([]User, error) {
	var users []User
	for _, u := range m.users {
		users = append(users, u)
	}

	return users, nil
}
