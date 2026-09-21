package order_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/domain/cart"
	"github.com/o-ga09/go-backend-template/internal/domain/order"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

func TestOrder_IsOwnedBy(t *testing.T) {
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
			o := &order.Order{UserID: tc.userID}
			if got := o.IsOwnedBy(tc.requesterID); got != tc.want {
				t.Errorf("IsOwnedBy() = %v, want %v", got, tc.want)
			}
		})
	}
}

func newProduct(id string, priceYen, stock int) *product.Product {
	p := &product.Product{PriceYen: priceYen, Stock: stock}
	p.ID = id
	return p
}

func TestNewFromCart(t *testing.T) {
	t.Run("カート内の全商品の在庫が足りる場合は注文を組み立てられる", func(t *testing.T) {
		c := &cart.Cart{
			UserID: "user-1",
			Items: []cart.CartItem{
				{ProductID: "p1", Quantity: 2},
				{ProductID: "p2", Quantity: 1},
			},
		}
		products := map[string]*product.Product{
			"p1": newProduct("p1", 1000, 10),
			"p2": newProduct("p2", 500, 5),
		}

		got, err := order.NewFromCart(c, products)
		if err != nil {
			t.Fatalf("NewFromCart() error = %v", err)
		}
		if got.UserID != "user-1" {
			t.Errorf("UserID = %q, want %q", got.UserID, "user-1")
		}
		if got.Status != order.StatusPending {
			t.Errorf("Status = %q, want %q", got.Status, order.StatusPending)
		}
		wantTotal := 1000*2 + 500*1
		if got.TotalPriceYen != wantTotal {
			t.Errorf("TotalPriceYen = %d, want %d", got.TotalPriceYen, wantTotal)
		}
		if len(got.Items) != 2 {
			t.Fatalf("Items length = %d, want 2", len(got.Items))
		}
		for _, item := range got.Items {
			switch item.ProductID {
			case "p1":
				if item.Quantity != 2 || item.UnitPriceYen != 1000 {
					t.Errorf("p1 item = %+v, want Quantity=2 UnitPriceYen=1000", item)
				}
			case "p2":
				if item.Quantity != 1 || item.UnitPriceYen != 500 {
					t.Errorf("p2 item = %+v, want Quantity=1 UnitPriceYen=500", item)
				}
			default:
				t.Errorf("unexpected item = %+v", item)
			}
		}
	})

	t.Run("在庫が不足している商品がある場合はErrInsufficientStock", func(t *testing.T) {
		c := &cart.Cart{
			UserID: "user-1",
			Items: []cart.CartItem{
				{ProductID: "p1", Quantity: 100},
			},
		}
		products := map[string]*product.Product{
			"p1": newProduct("p1", 1000, 1),
		}

		_, err := order.NewFromCart(c, products)
		if !errors.Is(err, errors.ErrInsufficientStock) {
			t.Errorf("error = %v, want ErrInsufficientStock", err)
		}
	})

	t.Run("カート内商品に対応する商品情報が渡されない場合はErrRecordNotFound", func(t *testing.T) {
		c := &cart.Cart{
			UserID: "user-1",
			Items: []cart.CartItem{
				{ProductID: "p-missing", Quantity: 1},
			},
		}
		products := map[string]*product.Product{}

		_, err := order.NewFromCart(c, products)
		if !errors.Is(err, errors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})
}
