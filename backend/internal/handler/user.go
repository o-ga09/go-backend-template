package handler

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	"github.com/o-ga09/go-backend-template/internal/handler/request"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
	"github.com/o-ga09/go-backend-template/pkg/session"
)

// UserHandler はユーザーリソース(/api/users)に関するエンドポイントを扱う。
// シンプルなCRUDのためusecase層を挟まず、domainのリポジトリを直接呼び出す
// (architecture.md「基本方針：レイヤードアーキテクチャ」)。
type UserHandler struct {
	repo       user.IUserRepository
	sessionMgr *session.Manager
}

// NewUserHandler はUserHandlerを生成する。
func NewUserHandler(repo user.IUserRepository, sessionMgr *session.Manager) *UserHandler {
	return &UserHandler{repo: repo, sessionMgr: sessionMgr}
}

// Create は初回ログイン時のユーザー作成(または既存ユーザーの特定)を行い、
// セッションCookieを発行する。
// POST /api/users
func (h *UserHandler) Create(c echo.Context) error {
	ctx := c.Request().Context()

	var req request.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.MakeBusinessError(ctx, "invalid request body")
	}

	if err := user.CanCreate(req.UID, req.DisplayName); err != nil {
		return errors.MakeBusinessError(ctx, "user create request is invalid")
	}

	existing, err := h.repo.FindByGoogleSub(ctx, req.UID)
	if err != nil && !errors.Is(err, errors.ErrRecordNotFound) {
		return errors.Wrap(ctx, err)
	}

	u := existing
	status := http.StatusOK
	if u == nil {
		email := req.Email
		if email == "" {
			// フロントエンドはEmailを送信しないため(known limitation。
			// backend/docs/auth.md参照)、UIDから一意なプレースホルダーを生成する。
			email = req.UID + "@no-email.invalid"
		}
		u = &user.User{
			GoogleSub: req.UID,
			Email:     email,
			Name:      req.DisplayName,
		}
		if err := h.repo.Create(ctx, u); err != nil {
			if errors.Is(err, errors.ErrUniqueConstraint) {
				return errors.MakeConflictError(ctx, "user already exists")
			}
			return errors.Wrap(ctx, err)
		}
		status = http.StatusCreated
	}

	setSessionCookie(c, h.sessionMgr, u.ID)
	return c.JSON(status, response.FromUser(u))
}

// GetByID はユーザーのプロフィールを取得する。ログインしていない場合は401、
// 本人以外のリソースへのアクセスは403で拒否する(認可はdomain.User.IsOwnedByに
// 委譲。architecture.md「ドメイン層のルール」)。
// GET /api/users/:id
func (h *UserHandler) GetByID(c echo.Context) error {
	ctx := c.Request().Context()

	requesterID := Ctx.GetUserID(ctx)
	if requesterID == "" {
		return errors.MakeAuthorizedError(ctx, "authentication required")
	}

	id := c.Param("id")
	u, err := h.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			return errors.MakeNotFoundError(ctx, "user not found")
		}
		return errors.Wrap(ctx, err)
	}

	if !u.IsOwnedBy(requesterID) {
		return errors.MakeAuthorizationError(ctx, "cannot access other user's resource")
	}

	return c.JSON(http.StatusOK, response.FromUser(u))
}
