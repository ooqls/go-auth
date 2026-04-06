package datagen

import "fmt"

type Permission struct {
	Group   string
	Kind    string
	Name    string
	Actions []string
}

func (p *Permission) ToString() []string {
	retStrings := []string{}
	for _, action := range p.Actions {
		retStrings = append(retStrings, fmt.Sprintf("%s:%s:%s:%s", p.Group, p.Kind, p.Name, action))
	}
	return retStrings
}

type RolePermissions = []Permission
