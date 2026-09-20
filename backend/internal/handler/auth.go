package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/o-ga09/go-backend-template/internal/domain/user"
	"github.com/o-ga09/go-backend-template/internal/handler/response"
	Ctx "github.com/o-ga09/go-backend-template/pkg/context"
	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// AuthHandler はログイン中セッション(/api/auth)に関するエンドポイントを扱う。
type AuthHandler struct {
	repo user.IUserRepository
}

// NewAuthHandler はAuthHandlerを生成する。
func NewAuthHandler(repo user.IUserRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

// CurrentUser はログイン中ユーザー情報を返す。セッションが無い/無効な場合は401を返す
// (フロントエンドはレスポンスが非2xxであれば未ログイン扱いにする)。
// GET /api/auth/user
func (h *AuthHandler) CurrentUser(c *echo.Context) error {
	ctx := c.Request().Context()

	userID := Ctx.GetUserID(ctx)
	if userID == "" {
		return errors.MakeAuthorizedError(ctx, "no active session")
	}

	u, err := h.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, errors.ErrRecordNotFound) {
			// セッションが指すユーザーが既に存在しない(削除済み等)。
			// 無効なセッションとして扱う。
			return errors.MakeAuthorizedError(ctx, "no active session")
		}
		return errors.Wrap(ctx, err)
	}

	return c.JSON(http.StatusOK, response.FromUser(u))
}

// Logout はセッションCookieを失効させる。DBアクセスは不要で常に成功する。
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *echo.Context) error {
	clearSessionCookie(c)
	return c.NoContent(http.StatusOK)
}
