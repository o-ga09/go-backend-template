// Package session はバックエンドが自ら発行・検証する軽量セッショントークンを
// 提供する。
//
// MVPでの認証方式の決定事項(詳細はbackend/docs/auth.md参照):
// フロントエンド(Next.js)はNextAuthのJWE暗号化Cookieでセッションを持つが、
// そのCookieをGo側で検証(復号)するのは、NextAuthのJWE鍵導出をGoで安全に
// 再実装するコストとリスクが高いため見送る。代わりに、フロントエンドが
// (NextAuthでのログイン確認後に)POST /api/users を呼んだ時点でバックエンドが
// 独自の署名付きセッションCookieを発行し、以降のリクエストはこのCookieのみで
// 識別する。
package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"github.com/o-ga09/go-backend-template/pkg/errors"
)

// sessionTTL はバックエンド発行セッションCookieの有効期間(backend/docs/auth.md参照)。
const SessionTTL = 7 * 24 * time.Hour

// CookieName はセッションCookieの名前。
const CookieName = "session_token"

// Manager はセッショントークンの発行・検証を行う。
// 秘密鍵とTTLのみを保持するステートレスな構造体。
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager はsecretとttlを指定してManagerを生成する。
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// Issue はuserIDに対する署名付きトークンと有効期限を発行する。
// トークンの形式は "<userID>.<expiresAtUnix>.<signature>" (base64url)。
func (m *Manager) Issue(userID string) (token string, expiresAt time.Time) {
	expiresAt = time.Now().Add(m.ttl)
	payload := userID + "." + strconv.FormatInt(expiresAt.Unix(), 10)
	return payload + "." + m.sign(payload), expiresAt
}

// Verify はトークンの署名と有効期限を検証し、有効であればuserIDを返す。
// 署名が不正な場合はerrors.ErrInvalidSession、期限切れの場合は
// errors.ErrSessionExpiredを返す(いずれもpkg/errorsで一元管理する生のsentinel。
// ログ出力・HTTPステータスへの変換は呼び出し元(handler)がerrors.MakeAuthorizedError
// 等に変換する際に行う)。
func (m *Manager) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.ErrInvalidSession
	}

	userID, expStr, sig := parts[0], parts[1], parts[2]
	if userID == "" {
		return "", errors.ErrInvalidSession
	}

	payload := userID + "." + expStr
	if !hmac.Equal([]byte(sig), []byte(m.sign(payload))) {
		return "", errors.ErrInvalidSession
	}

	expUnix, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return "", errors.ErrInvalidSession
	}
	if time.Now().Unix() > expUnix {
		return "", errors.ErrSessionExpired
	}

	return userID, nil
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
