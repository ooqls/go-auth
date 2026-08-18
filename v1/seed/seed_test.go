package seed

import (
	"fmt"
	"testing"

	v1 "github.com/ooqls/go-auth/v1"
)

func TestSeed(t *testing.T) {

	var err error = v1.ErrAlreadyExists(fmt.Errorf("test"), v1.M{})
	if _, ok := err.(*v1.AlreadyExistsError); !ok {
		t.Fatalf("expected error to be of type AlreadyExistsError, got %T", err)
	}
}
