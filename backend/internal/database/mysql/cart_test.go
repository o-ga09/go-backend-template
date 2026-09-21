package mysql_test

import (
	"context"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/domain/user"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// insertTestUser はusersテーブルへユーザーを新規作成する
// (carts.user_id/cart_items.product_idのFOREIGN KEY制約を満たすため)。
func insertTestUser(t *testing.T, ctx context.Context) *user.User {
	t.Helper()

	sub := "google-sub-" + uuid.GenerateID()
	u := &user.User{
		GoogleSub: sub,
		Email:     sub + "@example.com",
		Name:      "cart-test-user",
	}
	if err := mysql.NewUserRepository().Create(ctx, u); err != nil {
		t.Fatalf("failed to insert test user: %v", err)
	}
	return u
}

func TestCartRepository_CreateAndFindByUserID(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewCartRepository()
	u := insertTestUser(t, ctx)

	t.Run("存在しないユーザーのカートはErrRecordNotFound", func(t *testing.T) {
		_, err := repo.FindByUserID(ctx, u.ID)
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})

	created, err := repo.Create(ctx, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" {
		t.Fatal("BaseModelPluginによってIDが採番されているはず")
	}
	if created.UserID != u.ID {
		t.Errorf("UserID = %q, want %q", created.UserID, u.ID)
	}

	t.Run("FindByUserIDで取得できる(明細は空)", func(t *testing.T) {
		got, err := repo.FindByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("ID = %q, want %q", got.ID, created.ID)
		}
		if len(got.Items) != 0 {
			t.Errorf("Items length = %d, want 0", len(got.Items))
		}
	})

	t.Run("同じユーザーの二重作成はErrUniqueConstraint", func(t *testing.T) {
		_, err := repo.Create(ctx, u.ID)
		if !pkgerrors.Is(err, pkgerrors.ErrUniqueConstraint) {
			t.Errorf("error = %v, want ErrUniqueConstraint", err)
		}
	})
}

func TestCartRepository_AddItem(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewCartRepository()
	u := insertTestUser(t, ctx)
	c, err := repo.Create(ctx, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	p1 := insertTestProduct(t, ctx, "mug", "", 1000, 10)
	p2 := insertTestProduct(t, ctx, "cup", "", 500, 10)

	if err := repo.AddItem(ctx, c.ID, p1.ID, 2); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if err := repo.AddItem(ctx, c.ID, p2.ID, 1); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	t.Run("追加した商品がFindByUserIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}
		if len(got.Items) != 2 {
			t.Fatalf("Items length = %d, want 2", len(got.Items))
		}
	})

	t.Run("同じ商品を再度追加すると数量が加算される", func(t *testing.T) {
		if err := repo.AddItem(ctx, c.ID, p1.ID, 3); err != nil {
			t.Fatalf("AddItem() error = %v", err)
		}
		got, err := repo.FindByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}
		var found bool
		for _, item := range got.Items {
			if item.ProductID == p1.ID {
				found = true
				if item.Quantity != 5 {
					t.Errorf("Quantity = %d, want 5", item.Quantity)
				}
			}
		}
		if !found {
			t.Fatal("追加した商品明細が見つからない")
		}
	})
}

func TestCartRepository_UpdateItemQuantity(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewCartRepository()
	u := insertTestUser(t, ctx)
	c, err := repo.Create(ctx, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	p := insertTestProduct(t, ctx, "mug", "", 1000, 10)
	if err := repo.AddItem(ctx, c.ID, p.ID, 2); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	t.Run("数量を指定した値に変更できる", func(t *testing.T) {
		if err := repo.UpdateItemQuantity(ctx, c.ID, p.ID, 9); err != nil {
			t.Fatalf("UpdateItemQuantity() error = %v", err)
		}
		got, err := repo.FindByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}
		if len(got.Items) != 1 || got.Items[0].Quantity != 9 {
			t.Fatalf("Items = %+v, want single item with Quantity=9", got.Items)
		}
	})

	t.Run("既存の値と同じ数量を指定しても成功する(ErrRecordNotFoundにならない)", func(t *testing.T) {
		// clientFoundRows=trueが効いていない場合、値が変化しないUPDATEは
		// RowsAffected=0になり誤ってErrRecordNotFoundを返してしまう
		// (connect.go withClientFoundRowsの回帰テスト)。
		if err := repo.UpdateItemQuantity(ctx, c.ID, p.ID, 9); err != nil {
			t.Fatalf("UpdateItemQuantity() error = %v", err)
		}
	})

	t.Run("存在しない明細の変更はErrRecordNotFound", func(t *testing.T) {
		err := repo.UpdateItemQuantity(ctx, c.ID, "non-existent-product", 1)
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})
}

func TestCartRepository_RemoveItem(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewCartRepository()
	u := insertTestUser(t, ctx)
	c, err := repo.Create(ctx, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	p := insertTestProduct(t, ctx, "mug", "", 1000, 10)
	if err := repo.AddItem(ctx, c.ID, p.ID, 2); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	t.Run("商品を削除できる", func(t *testing.T) {
		if err := repo.RemoveItem(ctx, c.ID, p.ID); err != nil {
			t.Fatalf("RemoveItem() error = %v", err)
		}
		got, err := repo.FindByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByUserID() error = %v", err)
		}
		if len(got.Items) != 0 {
			t.Errorf("Items length = %d, want 0", len(got.Items))
		}
	})

	t.Run("存在しない明細の削除はErrRecordNotFound", func(t *testing.T) {
		err := repo.RemoveItem(ctx, c.ID, p.ID)
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})
}

func TestCartRepository_Clear(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewCartRepository()
	u := insertTestUser(t, ctx)
	c, err := repo.Create(ctx, u.ID)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	p1 := insertTestProduct(t, ctx, "mug", "", 1000, 10)
	p2 := insertTestProduct(t, ctx, "cup", "", 500, 10)
	if err := repo.AddItem(ctx, c.ID, p1.ID, 1); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}
	if err := repo.AddItem(ctx, c.ID, p2.ID, 1); err != nil {
		t.Fatalf("AddItem() error = %v", err)
	}

	if err := repo.Clear(ctx, c.ID); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	got, err := repo.FindByUserID(ctx, u.ID)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if len(got.Items) != 0 {
		t.Errorf("Items length = %d, want 0 after Clear()", len(got.Items))
	}
	if got.ID != c.ID {
		t.Errorf("Clear()後もカート自体は残っているはず: ID = %q, want %q", got.ID, c.ID)
	}
}
