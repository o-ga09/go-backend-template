package product_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/product"
)

func TestProduct_HasStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		quantity int
		want     bool
	}{
		{"在庫が数量以上ある場合は充足する", 10, 3, true},
		{"在庫が数量とちょうど一致する場合は充足する", 3, 3, true},
		{"在庫が数量未満の場合は充足しない", 2, 3, false},
		{"在庫が0の場合は充足しない", 0, 1, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &product.Product{Stock: tc.stock}
			got := p.HasStock(tc.quantity)
			if got != tc.want {
				t.Errorf("HasStock(%d) = %v, want %v", tc.quantity, got, tc.want)
			}
		})
	}
}
