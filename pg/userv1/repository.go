package userv1

import (
	"context"
	"errors"
	"fmt"

	"github.com/braumsmilk/go-auth/pg"
	"github.com/braumsmilk/go-auth/pg/tables"
)

var (
	ErrUserNotFound error = errors.New("user not found")
)

var _ Repository = &PostgresRepository{}

type Repository interface {
	GetUserFromName(ctx context.Context, user string) (*User, error)
	GetUser(ctx context.Context, id Id) (*User, error)
	GetUserName(ctx context.Context, id Id) (string, error)
	CreateUser(ctx context.Context, email, name, pw string) (Id, error)
	DeleteUser(ctx context.Context, id Id) error
	Authenticate(ctx context.Context, user, pw string) (bool, Id, error)
	GetAllUsers(ctx context.Context, page, pagesize int) ([]User, error)
}

type PostgresRepository struct{}

func (r *PostgresRepository) GetUserFromName(ctx context.Context, user string) (*User, error) {
	rows, err := pg.Get().QueryxContext(ctx, fmt.Sprintf("SELECT * FROM "+tables.Users+" WHERE name = $1"), user)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %v", err)
	}

	var u User
	for rows.Next() {
		err = rows.StructScan(&u)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row into user: %v", err)
		}
	}

	return &u, nil
}

func (r *PostgresRepository) GetUser(ctx context.Context, id Id) (*User, error) {
	rows, err := pg.Get().QueryxContext(ctx, fmt.Sprintf("SELECT * FROM %s WHERE id = $1;", tables.Users), id)
	if err != nil {
		return nil, fmt.Errorf("failed to query for user: %v", err)
	}

	users, err := pg.ParseRows[User](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse row into struct: %v", err)
	}

	if len(users) > 1 {
		return nil, pg.ErrTooManyRows
	}

	if len(users) == 1 {
		return &users[0], nil
	}

	return nil, ErrUserNotFound
}

func (r *PostgresRepository) GetUserName(ctx context.Context, id Id) (string, error) {
	dbCon := pg.Get()
	query := fmt.Sprintf("SELECT name FROM %s WHERE userid = $1", tables.Users)
	rows, err := dbCon.QueryContext(ctx, query, id)
	if err != nil {
		return "", fmt.Errorf("failed to query for username: %v", err)
	}

	var name string
	err = rows.Scan(&name)
	if err != nil {
		return "", fmt.Errorf("failed to scan into name: %v", err)
	}

	return name, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, email, name, pwDigest string) (Id, error) {
	dbCon := pg.Get()
	insert := fmt.Sprintf("INSERT INTO %s (\"email\", \"name\", \"created\") VALUES ($1, $2, NOW()::date ) RETURNING id;", tables.Users)
	tx, err := dbCon.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin new user transaction: %v", err)
	}

	stmt, err := tx.Prepare(insert)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare new user statement: %v", err)
	}

	res := stmt.QueryRowContext(ctx, email, name)
	if res.Err() != nil {
		return 0, fmt.Errorf("failed to execute new user transaction: %v", res.Err())
	}

	var userid int
	err = res.Scan(&userid)
	if err != nil {
		return 0, fmt.Errorf("failed to get userid of new user: %v", err)
	}

	pwinsert := fmt.Sprintf("INSERT INTO %s (\"userid\", \"pw_digest\", \"created\") VALUES ($1, $2, NOW()::date);", tables.Passwords)
	pwstmt, err := tx.Prepare(pwinsert)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare new password statement")
	}

	_, err = pwstmt.ExecContext(ctx, userid, pwDigest)
	if err != nil {
		return 0, fmt.Errorf("failed to create new password entry: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit new user transaction: %v", err)
	}

	return Id(userid), nil
}

func (r *PostgresRepository) DeleteUser(ctx context.Context, id Id) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tables.Users)
	_, err := pg.Get().ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("failed to execute delete: %v", err)
	}

	return nil
}

func (r *PostgresRepository) Authenticate(ctx context.Context, email, pwDigest string) (bool, Id, error) {
	query := fmt.Sprintf(
		"SELECT pw_digest, id FROM " + tables.Users +
			" LEFT JOIN " + tables.Passwords +
			" ON " + tables.Users + ".id = " + tables.Passwords + ".userid WHERE email = $1;")
	rows, err := pg.Get().QueryxContext(ctx,
		query, email)
	if err != nil {
		return false, -1, fmt.Errorf("failed to query authentication tables: %v", err)
	}

	for rows.Next() {
		var currentPwDigest string
		var userId Id
		err = rows.Scan(&currentPwDigest, &userId)
		if err != nil {
			return false, -1, fmt.Errorf("failed to scan password digest: %v", err)
		}

		return currentPwDigest == pwDigest, userId, nil
	}

	return false, 0, nil
}

func (r *PostgresRepository) GetAllUsers(ctx context.Context, page, pagesize int) ([]User, error) {
	q := fmt.Sprintf("SELECT * FROM %s ORDER BY id LIMIT $1 OFFSET $2", tables.Users)
	rows, err := pg.Get().QueryxContext(ctx, q, pagesize, page*pagesize)
	if err != nil {
		return nil, fmt.Errorf("failed to query for all users: %v", err)
	}

	users, err := pg.ParseRows[User](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to parse rows into users: %v", err)
	}

	return users, nil
}
