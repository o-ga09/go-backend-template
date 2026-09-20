package database

import (
	"context"

	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"gorm.io/gorm"
)

//go:generate go run github.com/matryer/moq@latest -out mock/transaction_manager_mock.go -pkg mock . ITransactionManager

// ITransactionManager は複数テーブルへの書き込みをトランザクションでラップする
// ためのインターフェース(transaction.md参照)。DBエンジン非依存の共通実装として
// internal/database(internal/database/mysqlと同じ階層)に置く。
type ITransactionManager interface {
	// RunInTx はfnをトランザクション内で実行する。fnがエラーを返すとロールバック
	// し、そのエラーをそのまま返す。fnがnilを返すとコミットする。
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// transactionManager はITransactionManagerのGORM実装。ステートレスであり、
// *gorm.DBをフィールドに持たず、呼び出しごとにctxから取得する
// (internal/database/mysqlのリポジトリと同じ理由)。
type transactionManager struct{}

// NewTransactionManager はTransactionManagerを生成する。
func NewTransactionManager() ITransactionManager {
	return &transactionManager{}
}

// RunInTx はctxから取得した*gorm.DBでトランザクションを開始し、txCtx(トランザクション
// 用の*gorm.DBを積んだcontext)でfnを実行する。fn内のリポジトリ呼び出しは
// Ctx.GetDBFromCtx(txCtx)でトランザクション用DBを取得するため、呼び出し側は
// リポジトリの引数をtxCtxに差し替えるだけでよい。
func (t *transactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db := Ctx.GetDBFromCtx(ctx)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := Ctx.SetDB(ctx, tx)
		return fn(txCtx)
	})
}
