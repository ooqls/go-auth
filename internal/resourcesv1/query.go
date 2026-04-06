package resourcesv1

type Query struct {
	Page     int32
	PageSize int32
	Name     []string
	Group    []string
	Kind     []string
}
