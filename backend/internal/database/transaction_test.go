package database_test

import (
	"context"
	stderrors "errors"
	"os"
	"testing"

	"github.com/o-ga09/go-backend-template/internal/database"
	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/domain/user"
	"github.com/o-ga09/go-backend-template/pkg/config"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/uuid"
)

// testing.md「DBは実DBを使う」に従い、実際のMySQL(docker composeで起動したもの)に
// 接続してテストする。DATABASE_URLが未設定の場合のみスキップする。
// DATABASE_URLが設定されているのに接続に失敗した場合(認証ミス・DB停止・
// マイグレーション未適用等)はt.Fatalfで落とす。ここをt.Skipにすると、CI環境で
// 接続設定が壊れた際にトランザクション基盤の回帰検知が無言のまま無効化されて
// しまうため。
func setupCtx(t *testing.T) context.Context {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL is not set; skipping database integration test")
	}

	ctx := t.Context()
	ctx = context.WithValue(ctx, config.ConfigKey, &config.Config{Database_url: dsn})
	ctx = Ctx.SetRequestID(ctx, uuid.GenerateID())

	db, err := mysql.Connect(ctx)
	if err != nil {
		t.Fatalf("DATABASE_URL is set but failed to connect to mysql: %v", err)
	}
	return Ctx.SetDB(ctx, db)
}

func TestTransactionManager_RunInTx(t *testing.T) {
	ctx := setupCtx(t)
	txManager := database.NewTransactionManager()
	userRepo := mysql.NewUserRepository()

	t.Run("エラーなしで完了した場合はコミットされる", func(t *testing.T) {
		sub := "tx-commit-" + uuid.GenerateID()
		var createdID string

		err := txManager.RunInTx(ctx, func(txCtx context.Context) error {
			u := &user.User{GoogleSub: sub, Email: sub + "@example.com", Name: "commit-user"}
			if err := userRepo.Create(txCtx, u); err != nil {
				return err
			}
			createdID = u.ID
			return nil
		})
		if err != nil {
			t.Fatalf("RunInTx() error = %v", err)
		}
		// コミットされ行が残るため、共有DBに残骸を蓄積させないよう明示的に削除する。
		t.Cleanup(func() { deleteUser(t, ctx, createdID) })

		got, err := userRepo.FindByID(ctx, createdID)
		if err != nil {
			t.Fatalf("FindByID() error = %v, want コミットされているためnil", err)
		}
		if got.GoogleSub != sub {
			t.Errorf("GoogleSub = %q, want %q", got.GoogleSub, sub)
		}
	})

	t.Run("エラーを返した場合はロールバックされる", func(t *testing.T) {
		sub := "tx-rollback-" + uuid.GenerateID()
		var createdID string
		wantErr := stderrors.New("intentional rollback")

		err := txManager.RunInTx(ctx, func(txCtx context.Context) error {
			u := &user.User{GoogleSub: sub, Email: sub + "@example.com", Name: "rollback-user"}
			if err := userRepo.Create(txCtx, u); err != nil {
				return err
			}
			createdID = u.ID
			return wantErr
		})
		if !stderrors.Is(err, wantErr) {
			t.Fatalf("RunInTx() error = %v, want %v", err, wantErr)
		}

		_, err = userRepo.FindByID(ctx, createdID)
		if !pkgerrors.Is(err, pkgerrors.ErrRecordNotFound) {
			t.Errorf("error = %v, want ErrRecordNotFound(ロールバックされているはず)", err)
		}
	})
}

// deleteUser はテストで作成したusers行を後始末する。IUserRepositoryにDeleteが
// 無いため、テスト専用にctxのDBを直接操作する(共有DBに残骸を蓄積させないため)。
func deleteUser(t *testing.T, ctx context.Context, id string) {
	t.Helper()
	if id == "" {
		return
	}
	db := Ctx.GetDBFromCtx(ctx)
	if err := db.WithContext(ctx).Exec("DELETE FROM users WHERE id = ?", id).Error; err != nil {
		t.Logf("failed to clean up test user (id=%s): %v", id, err)
	}
}
