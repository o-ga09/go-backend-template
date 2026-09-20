// Package request はハンドラが受け取るHTTPリクエストボディの型を定義する。
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
	// UID はGoogleのsub(NextAuthのセッションから取得される一意な識別子)。
	UID string `json:"uid"`
	// DisplayName はユーザーの表示名。usersテーブルのnameカラムに保存する。
	DisplayName string `json:"displayName"`
	// ProfileImage はプロフィール画像URL。現時点では永続化するカラムが
	// 存在しないため保存しない(レスポンスには空文字で含める)。
	ProfileImage string `json:"profileImage"`
	// Email はメールアドレス(オプショナル)。省略時はUIDから生成する。
	Email string `json:"email,omitempty"`
}
