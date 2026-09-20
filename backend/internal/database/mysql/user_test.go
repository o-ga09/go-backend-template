package mysql_test

import (
	"context"
	"os"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/domain/user"
	"github.com/o-ga09/go-backend-template/pkg/config"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// testing.md「DBは実DBを使う」に従い、実際のMySQL(docker composeで起動したもの)に
// 接続してテストする。DATABASE_URLが未設定/接続不可の場合はスキップする。
func setupCtx(t *testing.T) context.Context {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping mysql integration test")
	}

	ctx := t.Context()
	ctx = context.WithValue(ctx, config.ConfigKey, &config.Config{Database_url: dsn})
	ctx = Ctx.SetRequestID(ctx, uuid.GenerateID())

	db, err := mysql.Connect(ctx)
	if err != nil {
		t.Skipf("failed to connect to mysql; skipping: %v", err)
	}
	return Ctx.SetDB(ctx, db)
}

func TestUserRepository_CreateAndFind(t *testing.T) {
	ctx := setupCtx(t)
	repo := mysql.NewUserRepository()

	sub := "google-sub-" + uuid.GenerateID()
	u := &user.User{
		GoogleSub: sub,
		Email:     sub + "@example.com",
		Name:      "taro",
	}

	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if u.ID == "" {
		t.Fatal("BaseModelPluginによってIDが採番されているはず")
	}
	if u.Version != 1 {
		t.Fatalf("Version = %d, want 1(BaseModelPluginによる初期化)", u.Version)
	}
	if u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Fatal("CreatedAt/UpdatedAtが設定されているはず")
	}

	t.Run("FindByIDで取得できる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if got.GoogleSub != sub {
			t.Errorf("GoogleSub = %q, want %q", got.GoogleSub, sub)
		}
	})

	t.Run("FindByGoogleSubで取得できる", func(t *testing.T) {
		got, err := repo.FindByGoogleSub(ctx, sub)
		if err != nil {
			t.Fatalf("FindByGoogleSub() error = %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("ID = %q, want %q", got.ID, u.ID)
		}
	})

	t.Run("存在しないIDはErrRecordNotFound", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "non-existent-id")
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound", err)
		}
	})

	t.Run("重複するgoogle_subの作成はErrUniqueConstraint", func(t *testing.T) {
		dup := &user.User{GoogleSub: sub, Email: "dup-" + sub + "@example.com", Name: "dup"}
		err := repo.Create(ctx, dup)
		if !pkgerrors.Is(err, pkgerrors.ErrUniqueConstraint) {
			t.Errorf("error = %v, want ErrUniqueConstraint", err)
		}
	})

	t.Run("正常な更新でVersionがインクリメントされる", func(t *testing.T) {
		got, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		beforeVersion := got.Version
		got.Name = "jiro"
		if err := repo.Update(ctx, got); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		after, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if after.Name != "jiro" {
			t.Errorf("Name = %q, want %q", after.Name, "jiro")
		}
		if after.Version != beforeVersion+1 {
			t.Errorf("Version = %d, want %d", after.Version, beforeVersion+1)
		}
	})

	t.Run("古いVersionでの更新はErrOptimisticLockConflict", func(t *testing.T) {
		stale, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}

		// 他リクエストが先に更新したことを模擬する
		latest, err := repo.FindByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		latest.Name = "saburo"
		if err := repo.Update(ctx, latest); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		stale.Name = "should-conflict"
		err = repo.Update(ctx, stale)
		if !pkgerrors.Is(err, pkgerrors.ErrOptimisticLockConflict) {
			t.Errorf("error = %v, want ErrOptimisticLockConflict", err)
		}
	})
}
