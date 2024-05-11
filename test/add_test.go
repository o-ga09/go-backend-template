package test

import (
	"testing"

	"github.com/o-ga09/langchain-go/internal/add"
)

func TestAdd(t *testing.T) {
	if add.Add(1, 2) != 3 {
		t.Fail()
	}
}
