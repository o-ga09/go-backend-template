package mysql_test

import (
	"testing"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/domain/order"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

func TestOrderRepository_CreateFindByIDAndListByUserID(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewOrderRepository()
	u := insertTestUser(t, ctx)
	p1 := insertTestProduct(t, ctx, "mug", "", 1000, 10)
	p2 := insertTestProduct(t, ctx, "cup", "", 500, 10)

	o := &order.Order{
		UserID:        u.ID,
		Status:        order.StatusPending,
		TotalPriceYen: 1000*2 + 500,
		Items: []order.OrderItem{
			{ProductID: p1.ID, Quantity: 2, UnitPriceYen: 1000},
			{ProductID: p2.ID, Quantity: 1, UnitPriceYen: 500},
		},
	}

	t.Run("Createで注文と明細が作成される", func(t *testing.T) {
		if err := repo.Create(ctx, o); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if o.ID == "" {
			t.Fatal("BaseModelPluginによってIDが採番されているはず")
		}
		for _, item := range o.Items {
			if item.ID == "" {
				t.Errorf("order item ID is empty: %+v", item)
			}
			if item.OrderID != o.ID {
				t.Errorf("item.OrderID = %q, want %q", item.OrderID, o.ID)
			}
		}
	})

	t.Run("FindByIDで注文と明細が取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, o.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if got.UserID != u.ID {
			t.Errorf("UserID = %q, want %q", got.UserID, u.ID)
		}
		if got.TotalPriceYen != o.TotalPriceYen {
			t.Errorf("TotalPriceYen = %d, want %d", got.TotalPriceYen, o.TotalPriceYen)
		}
		if len(got.Items) != 2 {
			t.Fatalf("Items length = %d, want 2", len(got.Items))
		}
	})

	t.Run("存在しないIDはErrRecordNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-id")
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})

	t.Run("ListByUserIDで作成した注文が含まれる", func(t *testing.T) {
		got, err := repo.ListByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("ListByUserID() error = %v", err)
		}
		found := false
		for _, order := range got {
			if order.ID == o.ID {
				found = true
				if len(order.Items) != 2 {
					t.Errorf("Items length = %d, want 2", len(order.Items))
				}
			}
		}
		if !found {
			t.Errorf("ListByUserID() result does not contain created order(id=%s)", o.ID)
		}
	})

	t.Run("他ユーザーのListByUserIDには含まれない", func(t *testing.T) {
		other := insertTestUser(t, ctx)
		got, err := repo.ListByUserID(ctx, other.ID)
		if err != nil {
			t.Fatalf("ListByUserID() error = %v", err)
		}
		for _, order := range got {
			if order.ID == o.ID {
				t.Errorf("ListByUserID(other user) unexpectedly contains order(id=%s)", o.ID)
			}
		}
	})

	t.Run("複数注文がある場合、明細がまとめて取得されても注文ごとに正しく振り分けられる", func(t *testing.T) {
		o2 := &order.Order{
			UserID:        u.ID,
			Status:        order.StatusPending,
			TotalPriceYen: 500 * 3,
			Items: []order.OrderItem{
				{ProductID: p2.ID, Quantity: 3, UnitPriceYen: 500},
			},
		}
		if err := repo.Create(ctx, o2); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.ListByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("ListByUserID() error = %v", err)
		}

		var gotO1, gotO2 *order.Order
		for _, ord := range got {
			switch ord.ID {
			case o.ID:
				gotO1 = ord
			case o2.ID:
				gotO2 = ord
			}
		}
		if gotO1 == nil || gotO2 == nil {
			t.Fatalf("ListByUserID() result missing orders (o1found=%v, o2found=%v)", gotO1 != nil, gotO2 != nil)
		}
		if len(gotO1.Items) != 2 {
			t.Errorf("order1 Items length = %d, want 2", len(gotO1.Items))
		}
		if len(gotO2.Items) != 1 || gotO2.Items[0].Quantity != 3 {
			t.Errorf("order2 Items = %+v, want single item with quantity=3", gotO2.Items)
		}
	})
}
