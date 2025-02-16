package test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/add"
)

func TestAdd(t *testing.T) {
	if add.Add(1, 2) != 3 {
		t.Fail()
	}
}
