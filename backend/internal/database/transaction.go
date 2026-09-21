package database

import (
	"context"

	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"gorm.io/gorm"
)

//go:generate go run github.com/matryer/moq@latest -out mock/transaction_manager_mock.go -pkg mock . ITransactionManager

type ITransactionManager interface {
	// fnがエラーを返すとロールバックし、そのエラーをそのまま返す(呼び出し元が
	// 必要に応じてラップ・種別変換する)。
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct{}

func NewTransactionManager() ITransactionManager {
	return &transactionManager{}
}

func (t *transactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return Ctx.GetDBFromCtx(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(Ctx.SetDB(ctx, tx))
	})
}
