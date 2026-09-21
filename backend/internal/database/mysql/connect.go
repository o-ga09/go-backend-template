package mysql

import (
	"context"
	"fmt"
	"log"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/o-ga09/go-backend-template/internal/database"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
)

const (
	maxRetries      = 5
	retryInterval   = 2 * time.Second
	maxIdleConns    = 10
	maxOpenConns    = 100
	connMaxLifetime = 1 * time.Hour
)

func Connect(ctx context.Context) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	env := Ctx.GetCfgFromCtx(ctx)
	dsn, err := withClientFoundRows(env.Database_url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DATABASE_URL: %w", err)
	}
	logger := database.NewSentryLogger()
	// リトライ処理
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: false,
			},
			Logger: logger,
			// MySQLドライバのTranslate(error_translator.go)がERROR 1062/1451/1452を
			// gorm.ErrDuplicatedKey/ErrForeignKeyViolatedへ変換する。手動でのエラー
			// 番号判定(isDuplicateEntryErr等)が不要になる(pkg/errors参照)。
			TranslateError: true,
		})
		if err != nil {
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to open database after %d retries: %w", maxRetries, err)
			}
			time.Sleep(retryInterval)
			continue
		}

		// 接続成功
		break
	}

	// BaseModelを埋め込んだドメインエンティティのID採番・楽観ロックを扱うプラグイン
	if err := db.Use(NewBaseModelPlugin()); err != nil {
		return nil, fmt.Errorf("failed to register base model plugin: %w", err)
	}

	// SQLDBインスタンスを取得
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}

	// コネクションプール設定
	sqlDB.SetMaxIdleConns(10)           // アイドル状態の最大接続数
	sqlDB.SetMaxOpenConns(100)          // 最大接続数
	sqlDB.SetConnMaxLifetime(time.Hour) // 接続の最大生存期間
	return db, nil
}

// withClientFoundRows はDSNにclientFoundRows=trueを強制する。これが無いと
// UPDATE/DELETEのRowsAffectedは「実際に値が変わった行数」になり、
// 楽観ロック判定(base_model_plugin.go)や条件付きUPDATE(DecreaseStock等)の
// 「対象行が存在したか」の判定が、値が変化しない更新(例: 同じ数量への更新)で
// 誤ってconflict/not foundと判定されてしまう。clientFoundRows=trueにすると
// 「WHERE句に一致した行数」を返すようになり、この誤判定を防げる。
func withClientFoundRows(dsn string) (string, error) {
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return "", err
	}
	cfg.ClientFoundRows = true
	return cfg.FormatDSN(), nil
}
