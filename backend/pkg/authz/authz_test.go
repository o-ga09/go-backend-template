package authz_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/pkg/authz"
)

func TestIsOwner(t *testing.T) {
	tests := []struct {
		name        string
		requesterID string
		ownerID     string
		want        bool
	}{
		{"本人のリソースにはアクセスできる", "user-1", "user-1", true},
		{"他人のリソースにはアクセスできない", "user-1", "user-2", false},
		{"requesterIDが空の場合はアクセスできない", "", "user-1", false},
		{"ownerIDが空の場合はアクセスできない", "user-1", "", false},
		{"両方空の場合はアクセスできない", "", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := authz.IsOwner(tc.requesterID, tc.ownerID)
			if got != tc.want {
				t.Errorf("IsOwner(%q, %q) = %v, want %v", tc.requesterID, tc.ownerID, got, tc.want)
			}
		})
	}
}
