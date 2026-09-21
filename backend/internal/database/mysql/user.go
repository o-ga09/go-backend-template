package mysql

import (
	"context"
	stderrors "errors"

	mysqldriver "github.com/go-sql-driver/mysql"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	pkgerrors "github.com/o-ga09/go-backend-template/pkg/errors"
)

// mysqlErrDuplicateEntry はMySQLの一意制約違反(ER_DUP_ENTRY)のエラー番号。
const mysqlErrDuplicateEntry = 1062

// *gorm.DBをフィールドに保持せず、呼び出しごとにctxから取得する
// (transaction.mdのITransactionManagerと同じ理由: DB接続はSetDBミドルウェアが
// リクエストごとにcontextへ格納するため、リポジトリ自体はステートレスにする)。
type userRepository struct{}

func NewUserRepository() user.IUserRepository {
	return &userRepository{}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	if err := Ctx.GetDBFromCtx(ctx).Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByGoogleSub(ctx context.Context, googleSub string) (*user.User, error) {
	var u user.User
	if err := Ctx.GetDBFromCtx(ctx).Where("google_sub = ?", googleSub).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, u *user.User) error {
	if err := Ctx.GetDBFromCtx(ctx).Create(u).Error; err != nil {
		if isDuplicateEntryErr(err) {
			return pkgerrors.ErrUniqueConstraint
		}
		return err
	}
	return nil
}

// 楽観ロック競合(errors.ErrOptimisticLockConflict)はBaseModelPluginがセットし、
// そのまま呼び出し元に返る。
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	err := Ctx.GetDBFromCtx(ctx).Model(u).Updates(map[string]interface{}{
		"name":  u.Name,
		"email": u.Email,
	}).Error
	if err == nil {
		return nil
	}
	if isDuplicateEntryErr(err) {
		return pkgerrors.ErrUniqueConstraint
	}
	return err
}

func isDuplicateEntryErr(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	if stderrors.As(err, &mysqlErr) {
		return mysqlErr.Number == mysqlErrDuplicateEntry
	}
	return false
}
