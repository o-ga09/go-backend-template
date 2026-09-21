# 認証・認可設計 (Issue #3)

## 背景

フロントエンド(Next.js)はNextAuth(Google OAuth, JWTストラテジー)でセッションを持つ。
`frontend/context/authContext.tsx` はログイン確認後、バックエンドの以下3エンドポイントを
`credentials: 'include'`(Cookie送信)で呼び出す前提で実装されている。

- `GET /api/auth/user` — ログイン中ユーザー情報取得
- `POST /api/users` — 初回ログイン時のユーザー作成(既存ユーザーなら特定)
- `POST /api/auth/logout` — ログアウト

追加で、認可の実装例として `GET /api/users/:id` (本人のプロフィールのみ取得可能。
本人以外へのアクセスは403) を用意した。

## MVPでの認証方式の決定: バックエンド独自の署名付きセッションCookie

### 検討した選択肢

1. **NextAuthのCookieをバックエンドで検証する**
   NextAuthはセッションをJWE(暗号化JWT)としてCookieに保存する。バックエンド(Go)で
   これを検証するには、NextAuthのJWE鍵導出(HKDF等、バージョンによって仕様が変わる)を
   Go側で正確に再実装する必要があり、実装コストと将来のNextAuthバージョンアップ時の
   追従コストが高い。
2. **バックエンド独自のセッションCookieを発行する(採用)**
   フロントエンドがNextAuthでのログイン確認後に `POST /api/users` を呼んだ時点で、
   バックエンドが独自の署名付きCookieを発行し、以降のリクエストはこのCookieのみで
   ユーザーを識別する。NextAuthのトークン形式に依存しないため実装がシンプルで、
   NextAuthの内部仕様変更の影響を受けない。

MVPでは2を採用した。

### 実装

- **Cookie名**: `session_token` (`pkg/session.CookieName`)
- **値の形式**: `<userID>.<expiresAtUnix>.<signature>` (署名はHMAC-SHA256、
  `pkg/config.Config.SessionSecret` を鍵として使用。base64url未パディングエンコード)
- **TTL**: 7日間 (`internal/server.sessionTTL`)
- **属性**: `HttpOnly`, `Secure`, `SameSite=None`, `Path=/`
  フロントエンド(`localhost:3000`)とバックエンド(`localhost:8080`)が別オリジンであり、
  `credentials: 'include'` でCookieを送るため `SameSite=None` + `Secure` が必須
  (Fetch仕様上、`SameSite=None` は `Secure` 属性とセットでなければ多くのブラウザが
  Cookieを送信しない)。Chromeはlocalhostを信頼できるオリジンとして扱うため、開発環境の
  `http://localhost` でも `Secure` Cookieが送受信できる。
- **CORS**: `AllowOrigins` はワイルドカードではなく `pkg/config.Config.FrontendOrigin`
  (環境変数 `FRONTEND_ORIGIN`, デフォルト `http://localhost:3000`) の単一オリジンを許可し、
  `AllowCredentials: true` を設定する(`internal/server/middleware.go` の `CORS()`)。
  ワイルドカード + `AllowCredentials: true` の組み合わせはFetch仕様違反でブラウザに
  拒否されるため。
- **検証**: `internal/server/middleware.go` の `Authenticate` ミドルウェアが全リクエストで
  Cookieを検証し、有効であればユーザーIDのみを `pkg/context.SetUserID` でcontextに格納する
  (メールアドレス等のPIIはcontextに入れない。`context-propagation.md`)。Cookieが無い/
  不正な場合でもリクエスト自体は拒否せず、各ハンドラが `pkg/context.GetUserID` の結果
  (空文字なら未ログイン)を見て `errors.MakeAuthorizedError` (401) を返すかどうかを判断する。

### 既知の制限(MVPゆえの割り切り)

- **Googleのidトークンによる暗号学的な検証をバックエンドで行っていない。**
  `POST /api/users` はクライアントが送信した `uid`(Googleのsub)をそのまま信頼して
  ユーザーを作成/特定する。フロントエンドはNextAuthでのGoogleログインが成功した後に
  しかこのエンドポイントを呼ばないため実運用上のリスクは限定的だが、原理的には任意の
  `uid` を送ることで他人になりすませてしまう。本番運用に進める際は、GoogleのIDトークン
  (またはアクセストークン)をリクエストに含めてもらい、Google側のtokeninfoエンドポイント
  や公開鍵での署名検証を行う経路に強化することを推奨する。
- **`profileImage` は永続化していない。** `users` テーブルに保存先のカラムが無く、本Issueの
  範囲でのマイグレーション追加は対象外のため、レスポンスの `profileImage` は常に空文字を
  返す。将来 `avatar_url` 等のカラムを追加する形で解消する。
- **`email` はフロントエンドから送信されていない。** `frontend/context/authContext.tsx` の
  `POST /api/users` 呼び出しは `{ uid, displayName, profileImage }` のみを送信し、Emailを
  含めない(NextAuthのセッションにはEmailが含まれるが、フロントエンド側で転送されていない)。
  `users.email` はNOT NULL/UNIQUEのため、リクエストにEmailが無い場合は
  `<uid>@no-email.invalid` というプレースホルダーを生成して保存する
  (`internal/handler/user.go`)。`request.CreateUserRequest` にはオプショナルな `email`
  フィールドを用意してあるため、将来フロントエンドがNextAuthセッションのEmailを送信する
  よう修正されれば、そのまま実際のメールアドレスを保存できる。
- **セッション失効はTTL経過のみ。** Cookie発行後にユーザーを無効化・削除しても、TTLが
  切れるまで既存のセッションは有効なまま(サーバー側にセッションの取り消しリストが無い)。
  必要になれば `users` テーブルにセッションバージョン等を持たせて検証時に突き合わせる形に
  拡張できる。

## 認可: リソース所有者チェック

自分以外のユーザーのリソースにアクセスできないようにする判定は、`pkg/authz.IsOwner`
(リクエスト主体のIDとリソース所有者のIDを比較するだけの、ドメインに依存しない純粋関数)
として実装し、各ドメインエンティティがこれを呼ぶ薄いメソッドを持つ形にした
(`internal/domain/user/user.go` の `User.IsOwnedBy`)。

`architecture.md` は「ドメイン間に新しい矢印を作らない」ことを求めているため、この
所有者チェックのコアロジックは特定のドメイン(user)ではなく `pkg/` 配下の汎用ユーティリティ
として切り出した。将来のcart/order等のドメイン(#4)も、自身のエンティティに
`(o *Order) IsOwnedBy(requesterID string) bool { return authz.IsOwner(requesterID, o.UserID) }`
のような薄いメソッドを追加するだけで、同じ認可ルールを再利用できる。

`GET /api/users/:id` はこのロジックを使った実装例で、ログイン中ユーザー以外のIDが
指定された場合は `errors.MakeAuthorizationError` (403) を返す。

## CSRF対策 (Issue #4)

セッションCookieを `SameSite=None` で発行する設計(上記)は、クロスオリジンの
`credentials: 'include'` フェッチを成立させるために必須だが、副作用として
**クロスサイトのリクエスト(第三者サイトの自動送信フォーム等)でもCookieが自動送信される**
ため、状態変更エンドポイント自体にCSRF対策が無いと悪用されうる。Issue #3時点では
状態変更エンドポイントが `POST /api/users` (初回のみ・実害が限定的)しか無かったため
未対応だったが、Issue #4で `POST /api/orders` (ボディ無しで注文確定・在庫減算という
副作用を起こせる)等が追加されたことで顕在化したため、このIssueで対応した。

### 採用した方式: echo v5組み込みCSRFミドルウェア(`internal/server/middleware.go` の`CSRFProtection`)

優先されるのは **`Sec-Fetch-Site` ヘッダー(Fetch Metadata)による検証**。モダンブラウザの
fetch/XHRはこのヘッダーを自動付与し、クライアント側のJavaScriptから改ざんできないため、
`Origin` ヘッダーとの組み合わせで正規のクロスオリジン元(`pkg/config.Config.FrontendOrigin`
を`TrustedOrigins`に設定)からのリクエストかどうかを確実に検証できる。これにより
**フロントエンド側の実装変更は不要**(トークンの送出・保持が要らない)。

`Sec-Fetch-Site` を送らない環境(古いブラウザ等)向けには、ダブルサブミットCookie方式
(`X-CSRF-Token` ヘッダー)にフォールバックする。トークンは `GET /api/csrf`
(`internal/router/system.go`)で取得できる形にしてあるが、対象ブラウザが実質存在しない
現状ではこの経路がテスト以外で使われる想定は薄い(将来的なブラウザ後方互換性のための
保険)。

### 適用範囲

`CSRFProtection` は全リクエストに対して有効(GET/HEAD/OPTIONS/TRACEは検証対象外)。
特定のドメインだけを対象にするスキップ設定は行っていない(このプロジェクトの他の
ミドルウェア同様、`internal/server/server.go` の `Run()` でグローバルに適用する方針に
合わせた)。

## このIssueで合わせて修正した既存のバグ

実装を進める中で、認証・認可の受け入れ条件(401/403の判定)自体をブロックする以下の
既存バグを発見したため、本Issueの一部として修正した。

1. **`pkg/errors.MakeAuthorizationError`/`MakeAuthorizedError` のHTTPステータスコードが
   入れ替わっていた。** `.claude/rules/error-handling.md` の表では
   `MakeAuthorizedError`(認証失敗)→401、`MakeAuthorizationError`(認可失敗)→403と
   定義されているが、実装ではそれぞれ内部で付与する `ErrCode` が逆になっていた
   (`MakeAuthorizationError` が401用の`ErrCodeUnAuthorized`を、`MakeAuthorizedError` が
   403用の`ErrCodeUnAuthorization`を付与していた)。関数名・付与するメッセージ
   sentinelはそのままに、`ErrCode`の引数だけを入れ替えて修正した。
2. **`internal/server/middleware.go` の `ErrCodeToStatusAndMessage` に
   `errors.ErrCodeUnAuthorized`(401)のcaseが無かった。** 403用の
   `errors.ErrCodeUnAuthorization` のcaseはあったが401用が無く、認証失敗が
   すべてdefaultの500に落ちていた。401のcaseを追加した。
3. **`pkg/errors.Wrap(ctx, err)` が存在しなかった。** `.claude/rules/error-handling.md`や
   `transaction.md` は「詳細不明の外部エラーは `errors.Wrap(ctx, err)` でラップする」ことを
   前提に書かれているが、実装にこの関数が無かった。本Issueのハンドラ実装が前提とする
   ため追加した(既存のMake*Error関数と同じ形でスタックトレース付与・ログ出力を行う)。
4. **`ErrCodeToStatusAndMessage` の `ErrCodeInvalidArgument` が400を返していた。**
   `.claude/rules/error-handling.md` の表では `MakeBusinessError`(バリデーション違反)は
   422と定義されているため、422を返すよう修正した。

いずれも `internal/handler`, `internal/server` 以下が本Issue以前は空実装(`.gitkeep`のみ)
だったため、これまで実行されたことのないコードパスだった。
