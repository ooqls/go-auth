package userv1

import "fmt"

type Id int64

func (i *Id) String() string {
	return fmt.Sprintf("%d", i)
}

type User struct {
	UserId  Id     `db:"id"`
	Name    string `db:"name"`
	Email   string `db:"email"`
	Created string `db:"created"`
}
