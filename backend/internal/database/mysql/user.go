package mysql

import (
	"context"
	stderrors "errors"

	mysqldriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// mysqlErrDuplicateEntry はMySQLの一意制約違反(ER_DUP_ENTRY)のエラー番号。
const mysqlErrDuplicateEntry = 1062

// UserRepository はuserドメインのGORMリポジトリ。
// *gorm.DBをフィールドに保持せず、呼び出しごとにctxから取得する
// (transaction.mdのITransactionManagerと同じ理由: DB接続はSetDBミドルウェアが
// リクエストごとにcontextへ格納するため、リポジトリ自体はステートレスにする)。
// リポジトリはデータの読み書きのみを行い、ビジネスロジックを持たない。
type UserRepository struct{}

// NewUserRepository はUserRepositoryを生成する。
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

var _ user.IUserRepository = (*UserRepository)(nil)

// FindByID はIDでユーザーを検索する。
func (r *UserRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var u user.User
	if err := db.WithContext(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &u, nil
}

// FindByGoogleSub はGoogleのsubでユーザーを検索する。
func (r *UserRepository) FindByGoogleSub(ctx context.Context, googleSub string) (*user.User, error) {
	db := Ctx.GetDBFromCtx(ctx)

	var u user.User
	if err := db.WithContext(ctx).Where("google_sub = ?", googleSub).First(&u).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, pkgerrors.ErrRecordNotFound
		}
		return nil, err
	}
	return &u, nil
}

// Create はユーザーを新規作成する。ID/Version/タイムスタンプはBaseModelPluginが採番する。
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	db := Ctx.GetDBFromCtx(ctx)

	if err := db.WithContext(ctx).Create(u).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return pkgerrors.ErrUniqueConstraint
		}
		return err
	}
	return nil
}

// Update はユーザーを更新する。楽観ロック(WHERE version = ?)はBaseModelPluginが
// 自動的に付与する。競合時はerrors.ErrOptimisticLockConflictを返す。
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	db := Ctx.GetDBFromCtx(ctx)

	err := db.WithContext(ctx).Model(u).Updates(map[string]interface{}{
		"name":  u.Name,
		"email": u.Email,
	}).Error
	if err == nil {
		return nil
	}
	if isDuplicateEntryErr(err) {
		return pkgerrors.ErrUniqueConstraint
	}
	// errors.ErrOptimisticLockConflict(BaseModelPluginが競合時にセットする)は
	// そのまま返す。呼び出し元がerrors.Is(err, errors.ErrOptimisticLockConflict)で
	// 判別しerrors.MakeConflictErrorに変換する。
	return err
}

func isDuplicateEntryErr(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	if stderrors.As(err, &mysqlErr) {
		return mysqlErr.Number == mysqlErrDuplicateEntry
	}
	return false
}
