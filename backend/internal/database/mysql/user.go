package mysql

import (
	"context"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
)

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
	return Ctx.GetDBFromCtx(ctx).Create(u).Error
}

// 楽観ロック競合(errors.ErrOptimisticLockConflict)はBaseModelPluginがセットし、
// そのまま呼び出し元に返る。
func (r *userRepository) Update(ctx context.Context, u *user.User) error {
	return Ctx.GetDBFromCtx(ctx).Model(u).Updates(map[string]interface{}{
		"name":  u.Name,
		"email": u.Email,
	}).Error
}
