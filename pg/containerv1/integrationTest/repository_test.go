package integrationtest

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/braumsmilk/go-auth/pg/containerv1"
	"github.com/braumsmilk/go-auth/pg/userv1"
	"github.com/braumsmilk/go-auth/testutils"
	"github.com/stretchr/testify/assert"
)

var cr containerv1.ContainerRepository

func TestMain(m *testing.M) {
	c := testutils.InitPostgres()
	defer func() {
		log.Printf("stopping container")
		timeout := time.Second * 30
		err := c.Stop(context.Background(), &timeout)
		if err != nil {
			panic(err)
		}
	}()
	cr = &containerv1.PostgresRepository{}

	m.Run()
}

func TestContainerRepository(t *testing.T) {
	ctx := context.Background()
	userr := userv1.PostgresRepository{}

	uid, err := userr.CreateUser(ctx, "user", "name", "pw")
	assert.Nilf(t, err, "should not error when creating user")

	cid, err := cr.CreateContainer(ctx, "test", containerv1.Channel)
	assert.Nilf(t, err, "should not error when creating container")

	err = cr.AddUserToContainer(ctx, uid, cid)
	assert.Nilf(t, err, "should not error when adding user to container")

	users, err := cr.GetUsersInContainer(ctx, cid, 0, 10)
	assert.Nilf(t, err, "should not error when getting users in container")
	assert.Lenf(t, users, 1, "should only have one user in container")

	containers, err := cr.GetJoinedContainers(ctx, uid)
	assert.Nilf(t, err, "should not fail to get user joined containers")
	assert.Lenf(t, containers, 1, "user has only joined one container")

	err = cr.RemoveUserFromContainer(ctx, uid, cid)
	assert.Nilf(t, err, "should not fail to remove user from container")

	users, err = cr.GetUsersInContainer(ctx, cid, 0, 10)
	assert.Nilf(t, err, "should not fail to get users in container")
	assert.Lenf(t, users, 0, "should not have any users in container")

	err = userr.DeleteUser(ctx, uid)
	assert.Nilf(t, err, "should not error when deleting user")

	users, err = cr.GetUsersInContainer(ctx, cid, 0, 10)
	assert.Nilf(t, err, "should not error when getting users in container")
	assert.Lenf(t, users, 0, "should not have any users in container")

	ctrs, err := cr.GetAllContainers(ctx)
	assert.Nilf(t, err, "should not error when getting all containers")
	assert.Lenf(t, ctrs, 1, "should only have one containers")

	err = cr.DeleteContainer(ctx, cid)
	assert.Nilf(t, err, "should not error when deleting container")

	ctrs, err = cr.GetAllContainers(ctx)
	assert.Nilf(t, err, "should not error when getting all containers")
	assert.Lenf(t, ctrs, 0, "should not have any containers")

	c, err := cr.GetContainer(ctx, cid)
	assert.Nilf(t, c, "should not have gotten a container")
	assert.Nilf(t, err, "should not get error when no containers are found")

	_, err = cr.CreateContainer(ctx, "new", containerv1.Channel)
	assert.Nilf(t, err, "should not fail to create 'new' container")

	qc, err := cr.QueryContainers(ctx, "ne")
	assert.Nilf(t, err, "should not fail to query containers")
	assert.Lenf(t, qc, 1, "should get one container when querying containers")
}

func BenchmarkContainerRepository_AddUserToContainer(t *testing.B) {
	usersN := 1000
	containerN := 1000
	ctx := context.Background()
	userr := userv1.PostgresRepository{}
	var err error
	t.Run(fmt.Sprintf("creating %d users", usersN), func(b *testing.B) {
		for i := 0; i < usersN; i++ {
			_, err = userr.CreateUser(ctx, fmt.Sprintf("user-%d", i), "name", "aaa")
			assert.Nilf(t, err, "should not error when creating user")
		}
	})

	var cid containerv1.Id
	t.Run(fmt.Sprintf("creating %d containers", containerN), func(b *testing.B) {
		for i := 0; i < containerN; i++ {
			cid, err = cr.CreateContainer(ctx, fmt.Sprintf("tester-container-%d", i), containerv1.Channel)
			assert.Nilf(t, err, "should not error when creating container")
		}
	})

	var users []userv1.User
	t.Run("getting all users", func(b *testing.B) {
		users, err = userr.GetAllUsers(ctx, 0, usersN)
		assert.Nilf(t, err, "should not error when getting all users")
		assert.Lenf(t, users, usersN, "should have gotten all users")
	})

	t.Run("adding all users to a container", func(b *testing.B) {
		for _, u := range users {
			t.Logf("adding user %d", u.UserId)
			err = cr.AddUserToContainer(ctx, u.UserId, cid)
			assert.Nilf(t, err, "should not error when adding user to container")
		}
	})

	t.Run("getting all users in container", func(b *testing.B) {
		usersInContainer, err := cr.GetUsersInContainer(ctx, cid, 0, usersN)
		assert.Nilf(t, err, "should not error when getting users in container")
		assert.Lenf(t, usersInContainer, usersN, "should have all the users in this container")
	})

	t.Run("deleting all users", func(b *testing.B) {
		for _, u := range users {
			err := userr.DeleteUser(ctx, u.UserId)
			assert.Nilf(t, err, "should not error when deleting user")
		}
	})

	var containers []containerv1.Container
	t.Run("getting all containers", func(b *testing.B) {
		containers, err = cr.GetAllContainers(ctx)
		assert.Nilf(t, err, "should not error when getting all containers")
	})

	t.Run("deleting all containers", func(b *testing.B) {
		for _, c := range containers {
			assert.Nilf(t, cr.DeleteContainer(ctx, c.Id), "should not error when deleting container")
		}
	})
}

func BenchmarkContainerRepository_createContainer(t *testing.B) {
	ctx := context.Background()
	containerN := 1000
	var ids []containerv1.Id = []containerv1.Id{}
	var err error
	t.Run(fmt.Sprintf("create %d container", containerN), func(b *testing.B) {
		for i := 0; i < containerN; i++ {
			id, err := cr.CreateContainer(ctx, "test", containerv1.Channel)
			assert.Nilf(b, err, "should not error when creating a container")
			ids = append(ids, id)
		}
	})

	var c *containerv1.Container
	t.Run(fmt.Sprintf("get %d containers", containerN), func(b *testing.B) {
		for _, id := range ids {
			c, err = cr.GetContainer(ctx, id)
			assert.Nilf(t, err, "should not error when getting container")
			assert.NotNilf(t, c, "should not have gotten a nil container")
			assert.Equalf(t, "test", c.Name, "should have gotten the right container")
		}
	})

	t.Run(fmt.Sprintf("delete %d containers", containerN), func(b *testing.B) {
		for _, id := range ids {
			err = cr.DeleteContainer(ctx, id)
			assert.Nilf(t, err, "should not error deleting a container")
		}
	})

}
