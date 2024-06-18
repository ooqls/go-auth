package usercachev1

type MockUserCache struct {
	m map[int]string
}

func(m *MockUserCache) AddUser(id int, name string) {
	m.m[id] = name
}

func(m *MockUserCache) GetUser(id int) string {
	return m.m[id]
}

