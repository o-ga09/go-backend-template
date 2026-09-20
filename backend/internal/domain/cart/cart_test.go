package cart_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
)

func TestCart_IsOwnedBy(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		requesterID string
		want        bool
	}{
		{"本人の場合はtrue", "user-1", "user-1", true},
		{"他人の場合はfalse", "user-1", "user-2", false},
		{"requesterIDが空の場合はfalse", "user-1", "", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &cart.Cart{UserID: tc.userID}
			if got := c.IsOwnedBy(tc.requesterID); got != tc.want {
				t.Errorf("IsOwnedBy(%q) = %v, want %v", tc.requesterID, got, tc.want)
			}
		})
	}
}

func TestCart_CanCheckout(t *testing.T) {
	tests := []struct {
		name    string
		items   []cart.CartItem
		wantErr bool
	}{
		{"商品が入っている場合は注文できる", []cart.CartItem{{ProductID: "p1", Quantity: 1}}, false},
		{"カートが空の場合は注文できない", nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := &cart.Cart{Items: tc.items}
			err := c.CanCheckout()
			if tc.wantErr {
				if err == nil {
					t.Fatal("CanCheckout() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("CanCheckout() error = %v, want nil", err)
			}
		})
	}
}
