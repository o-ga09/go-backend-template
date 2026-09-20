// Package request はハンドラが受け取るHTTPリクエストのパラメータ型を定義する。
package request

// CreateUserRequest はユーザー作成(初回ログイン時)リクエストのボディ。
//
// フロントエンド(frontend/context/authContext.tsx)は現状
// { uid, displayName, profileImage } のみを送信し、Emailは含めない
// (NextAuthのセッションにはEmailが含まれるが、フロントエンド側でまだ
// 転送されていない)。バックエンドのusersテーブルはemailをNOT NULL/UNIQUEと
// しているため、Emailが省略された場合はUID(Googleのsub)から一意な
// プレースホルダーを生成する(internal/handler/user.go参照。
// backend/docs/auth.md「既知の制限」)。Emailは将来フロントエンドが
// NextAuthセッションのメールアドレスを送るようになった際に使われる、
// 現時点ではオプショナルなフィールドとして定義しておく。
type CreateUserRequest struct {
	UID          string `json:"uid" validate:"required"`
	DisplayName  string `json:"displayName" validate:"required"`
	ProfileImage string `json:"profileImage"`
	Email        string `json:"email,omitempty" validate:"omitempty,email"` // 省略時はUIDから生成する
}

// GetUserRequest はユーザー取得リクエストのパスパラメータ。
type GetUserRequest struct {
	ID string `param:"id" validate:"required"`
}
