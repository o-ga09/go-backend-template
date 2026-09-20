package mysql_test

import (
	"context"
	"testing"
	"time"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/domain/product"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// insertTestProduct はproductsテーブルへ直接SQLでテストデータをinsertする
// (productドメインには本Issue(#4)スコープ外のCreateが無いため。
// backend/tmp/task-2-brief.md参照)。
func insertTestProduct(t *testing.T, ctx context.Context, name, description string, priceYen, stock int) *product.Product {
	t.Helper()
	return insertTestProductAt(t, ctx, name, description, priceYen, stock, time.Now())
}

// insertTestProductAt はinsertTestProductのcreated_at/updated_atを指定できる版。
// List()のORDER BY(created_at DESC)を検証するために作成日時を制御したい場合に使う。
func insertTestProductAt(t *testing.T, ctx context.Context, name, description string, priceYen, stock int, createdAt time.Time) *product.Product {
	t.Helper()

	db := Ctx.GetDBFromCtx(ctx)
	p := &product.Product{
		Name:        name,
		Description: description,
		PriceYen:    priceYen,
		Stock:       stock,
	}
	p.ID = uuid.GenerateID()
	p.Version = 1
	p.CreatedAt = createdAt
	p.UpdatedAt = createdAt

	if err := db.WithContext(ctx).Exec(
		"INSERT INTO products (id, version, name, description, price_yen, stock, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		p.ID, p.Version, p.Name, p.Description, p.PriceYen, p.Stock, p.CreatedAt, p.UpdatedAt,
	).Error; err != nil {
		t.Fatalf("failed to insert test product: %v", err)
	}
	return p
}

func TestProductRepository_FindByIDAndList(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewProductRepository()
	db := insertTestProduct(t, ctx, "irodori-mug", "彩り マグカップ", 1500, 10)

	t.Run("FindByIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, db.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if got.Name != "irodori-mug" {
			t.Errorf("Name = %q, want %q", got.Name, "irodori-mug")
		}
		if got.PriceYen != 1500 {
			t.Errorf("PriceYen = %d, want %d", got.PriceYen, 1500)
		}
		if got.Stock != 10 {
			t.Errorf("Stock = %d, want %d", got.Stock, 10)
		}
	})

	t.Run("存在しないIDはErrRecordNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-id")
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})

	t.Run("Listで作成した商品が含まれる", func(t *testing.T) {
		got, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		found := false
		for _, p := range got {
			if p.ID == db.ID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() result does not contain created product(id=%s)", db.ID)
		}
	})

	t.Run("Listは作成日時の降順で返す", func(t *testing.T) {
		older := insertTestProductAt(t, ctx, "older-product", "", 100, 1, time.Now().Add(-2*time.Hour))
		newer := insertTestProductAt(t, ctx, "newer-product", "", 200, 1, time.Now().Add(-1*time.Hour))

		got, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		var newerIdx, olderIdx = -1, -1
		for i, p := range got {
			if p.ID == newer.ID {
				newerIdx = i
			}
			if p.ID == older.ID {
				olderIdx = i
			}
		}
		if newerIdx == -1 || olderIdx == -1 {
			t.Fatalf("List() result does not contain both test products (newerIdx=%d, olderIdx=%d)", newerIdx, olderIdx)
		}
		if newerIdx >= olderIdx {
			t.Errorf("newer product index = %d, older product index = %d; want newer before older", newerIdx, olderIdx)
		}
	})
}
