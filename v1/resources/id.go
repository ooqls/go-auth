package resources

import "github.com/google/uuid"

type resourceIdentifier struct {
	Name  string
	Group string
	Kind  string
	ID    uuid.UUID
}

func (r resourceIdentifier) IsByName() bool {
	return r.Name != "" && r.Group != "" && r.Kind != ""
}

func (r resourceIdentifier) IsById() bool {
	return r.ID != uuid.Nil
}

func ByName(group, kind, name string) resourceIdentifier {
	return resourceIdentifier{
		Name:  name,
		Group: group,
		Kind:  kind,
	}
}

func ById(id uuid.UUID) resourceIdentifier {
	return resourceIdentifier{
		ID: id,
	}
}
