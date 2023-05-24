package dean

import (
	"testing"
)

func TestEmptyId(t *testing.T) {
	defer func() { _ = recover() }()
	// should panic with empty Id
	NewThing("", "foo", "bar")
	t.Errorf("did not panic")
}

func TestEmptyModel(t *testing.T) {
	defer func() { _ = recover() }()
	// should panic with empty Id
	NewThing("foo", "", "bar")
	t.Errorf("did not panic")
}

func TestEmptyName(t *testing.T) {
	defer func() { _ = recover() }()
	// should panic with empty Id
	NewThing("foo", "bar", "")
	t.Errorf("did not panic")
}
