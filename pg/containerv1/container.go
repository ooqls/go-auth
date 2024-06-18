package containerv1

type Id int64

type ContainerType int32

var (
	Group   ContainerType = 0
	Direct  ContainerType = 1
	Channel ContainerType = 2
)

type Container struct {
	Id            Id            `db:"id"`
	Name          string        `db:"name"`
	ContainerType ContainerType `db:"container_type"`
}
