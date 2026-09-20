# エラーハンドリングルール

対象: `backend/`（Go / Echo / GORM）。

## 基本原則

エラーは `pkg/errors` パッケージを経由して扱う。`fmt.Errorf` や標準 `errors.New` を直接使わない。

このパッケージは **`github.com/newmo-oss/ergo`** をベースに構築されている。`ergo` はスタックトレース付きエラー・sentinel エラー・エラーコードを統合した Go エラーライブラリ。

```go
// 誤
return fmt.Errorf("order not found: %w", err)

// 正
return errors.MakeNotFoundError(ctx, "order not found")
```

## ergo の使い方方針

`ergo` を直接 import するのは `pkg/errors/errors.go` のみ。それ以外のパッケージは `pkg/errors` 経由で使う。

### sentinel エラーの定義

新しいドメインエラーは `ergo.NewSentinel` で定義し、`pkg/errors/errors.go` に追加する（`ergo.New` ではない。sentinel はパッケージ変数として１回しか作られないため、`ergo.New` を使うとパッケージ初期化時点の無意味なスタックトレースが固定で付与されてしまう。`ergo.NewSentinel` はスタックトレースを持たない設計になっている）。

```go
// errors.go に追加
var ErrOrderLimitExceeded  = ergo.NewSentinel("order limit exceeded")
var ErrExternalUnavailable = ergo.NewSentinel("external service unavailable")
```

### エラーのラップ（ergo.Wrap）

下位レイヤーのエラーに文脈を付けるとき。スタックトレースが自動付与される。

```go
// ergo.Wrap は内部で使う。外部からは errors.Wrap(ctx, err) を呼ぶ
err = ergo.Wrap(ErrRecordNotFound, "order lookup failed")
```

### エラーコードの付与（ergo.WithCode）

HTTP ステータスに対応するコードを付与するとき。`Make*Error` 関数が内部で自動付与するため、通常は直接呼ばない。

```go
err = ergo.WithCode(err, ErrCodeNotFound) // Make*Error 内部で行う
```

### スタックトレースの取得（ergo.StackTraceOf）

`Make*Error` / `Wrap` のたびに自動で `stacktrace` 属性としてログに出力する。呼び出し側で直接呼ぶ必要はない。

`ergo.New` / `ergo.Wrap` はその場で `runtime.Callers` によるスタックトレースを記録する（`ergo.NewSentinel` で作った sentinel には無い）。これは「そのエラーが作られた行」だけでなく、**その時点のゴルーチンの呼び出し履歴（呼び出し元の呼び出し元…）を丸ごと含む**ため、`stacktrace` 属性だけで発生元とそこに至った呼び出し経路の両方を特定できる。ログに短い文字列しか出ていない場合、多くは `Wrap`/`Make*Error` に渡した `err` が sentinel（`ErrXxx = ergo.NewSentinel(...)`）そのもので、下位レイヤーの実エラーを経由していないことが原因。**握りつぶさず、実際に失敗した `err` を渡すこと**（次節）。

## 最優先の制約：エラーに機密情報を載せない 🚫

**エラーメッセージはログ・監視チャネルへそのまま流れる。** `ergo` はスタックトレースを自動付与するため、**エラーに載せた文字列は確実に外へ出る**前提で書く。

```go
// 誤: パスワードがエラーメッセージに入る
return errors.MakeBusinessError(ctx, fmt.Sprintf("invalid password: %s", password))

// 誤: メールアドレスが入る
return errors.MakeNotFoundError(ctx, "user not found: "+user.Email)

// 正: 識別子と種別だけ
return errors.MakeBusinessError(ctx, "password is invalid")
return errors.MakeNotFoundError(ctx, "user not found")
```

**載せてよいもの**：ID、件数、ステータス、エラー種別。
**載せてはいけないもの**：パスワード・トークンなどの認証情報、個人情報（メールアドレス・氏名・住所等）、決済情報、外部APIのリクエスト／レスポンス本体。

外部API呼び出しの失敗をラップするときは、**リクエスト・レスポンス本体を含めない**。ただし「含めない」は「エラー情報ごと握りつぶす」ことではない。SDK の `err.Error()` を直接ログに出さないことと、失敗の原因（タイムアウトか・レート制限か・認証エラーか）を診断可能な形で残すことは両立できる。**握りつぶすと、本番障害の一次切り分けにアプリログが使えなくなる。**

```go
// 誤①: SDK のエラーをそのままラップする（err.Error() にAPIレスポンスの生JSONや
// 決済情報が含まれうる）
if err != nil {
    return errors.Wrap(ctx, err)
}

// 誤②: 実際に発生したエラー(err)を最後まで一度も使わず、無関係な sentinel だけを
// 返す。err.Error() は避けられているが、ステータスコードやエラー種別も一緒に
// 消えてしまい、ログからは「外部API呼び出しが失敗した」以上の情報が一切取れない
if err != nil {
    logger.Warn(ctx, "payment api call failed", slog.String("kind", "payment"))
    return errors.MakeUnavailableError(ctx, errors.ErrExternalUnavailable)
}

// 正: err自体は機密情報を含む可能性があるためログに出さないが、機密情報を含まない
// 安全なフィールド（HTTPステータスコード・SDKのエラー種別・リクエストIDなど）
// だけを抽出し、attrs として Make*Error に渡す。err自体は sentinel を返すため
// クライアントへは相変わらず種別化されたメッセージしか返らない。
// ログ出力はMakeUnavailableError内部で一度だけ行われる（呼び出し側で
// logger.Warn を別途呼ばない。「ログ出力のタイミング」参照）。
if err != nil {
    var apiErr *paymentsdk.Error
    attrs := []slog.Attr{slog.String("kind", "payment")}
    if errors.As(err, &apiErr) {
        attrs = append(attrs,
            slog.Int("status_code", apiErr.StatusCode),
            slog.String("error_type", string(apiErr.Type())),
            slog.String("request_id", apiErr.RequestID),
        )
    }
    return errors.MakeUnavailableError(ctx, errors.ErrExternalUnavailable, attrs...)
}
```

同じ考え方は外部 SDK 全般（DB ドライバ・KMS・決済/通知API等）に適用する。**「本文が混入しうるので raw error を出さない」と「失敗の原因を一切残さない」はイコールではない。** 安全に取り出せる分類情報（ステータスコード・エラーコード・型名など）まで一緒に捨てないこと。

## エラー種別と生成関数

| 状況 | 使う関数 | HTTP ステータス相当 |
|---|---|---|
| 認証失敗（トークン無効など） | `errors.MakeAuthorizedError(ctx, msg)` | 401 |
| 認可失敗（権限不足） | `errors.MakeAuthorizationError(ctx, msg)` | 403 |
| リソース未発見 | `errors.MakeNotFoundError(ctx, msg)` | 404 |
| 競合（重複登録など） | `errors.MakeConflictError(ctx, msg)` | 409 |
| バリデーション違反 | `errors.MakeBusinessError(ctx, msg)` | 422 |
| 予期しない内部エラー | `errors.MakeSystemError(ctx, err)` | 500 |
| 頻度制限（リトライ制限等。必要になったら追加） | `errors.MakeRateLimitError(ctx, msg)` | 429 |
| 外部サービスの一時的な障害（必要になったら追加） | `errors.MakeUnavailableError(ctx, err, attrs...)` | 503 |

`MakeRateLimitError` / `MakeUnavailableError` は現時点の `pkg/errors` にはまだ無い。外部API連携など必要になった時点で `MakeNotFoundError` 等と同じ形（`ergo.Wrap` + `ergo.WithCode` + ログ出力）で追加する。

## ctxを取れない箇所でのコード付与（`WithInvalidArgumentCode`）

`Make*Error` は必ず `ctx` を取り、その場でログ出力する。しかし `echo.Validator`/`echo.Binder` インターフェース（`Validate(i any) error` / `Bind(c *echo.Context, target any) error`）のように、**呼び出し元の型が決まっていて `ctx` を追加できない**箇所がある（`pkg/validator` が実装例。`request-validation.md`）。

このような箇所では `errors.WithInvalidArgumentCode(err)` で **422 のコードだけ**を事前に付与する（ログは出さない）。呼び出し元がその後 `ctx` を持つ地点（ハンドラ）で `errors.Wrap(ctx, err)` を呼べば、そこで初めてログが出力され、かつ事前に付与しておいたコードは `ergo.CodeOf` の Unwrap チェーン探索で保持されたまま引き継がれる。**`ErrTypeBussiness` ではラップしない**（`IsWrapped` が `true` になり `Wrap` がログを出さずに素通ししてしまうため。「ログ出力のタイミング」参照）。

```go
// pkg/validator/validator.go: ctxを持たないValidate(i any) errorの中
return errors.WithInvalidArgumentCode(translate(i, verrs))

// ハンドラ側: ctxがある地点でWrapするだけでログ出力・422変換が完結する
if err := c.Validate(&req); err != nil {
    return errors.Wrap(ctx, err)
}
```

他の状況（`ctx` を渡せる通常のハンドラ・service・repository）では、これまで通り `Make*Error` を使う。`WithInvalidArgumentCode` は「`ctx` が取れない」という制約がある場合だけの例外。

## エラーのラップとログ

- **ラップのみ**: 詳細不明の外部エラーは `errors.Wrap(ctx, err)` でラップする（ログも自動出力）
- **新規生成**: ドメインルール違反は `errors.Make*Error` で意味のあるエラーを生成する
- **伝搬**: 一度ラップしたエラーは再ラップしない。`errors.IsWrapped(err)` で確認してから `Wrap` を呼ぶ

```go
// 外部エラーをラップして伝搬（ログはWrap内で出力済み）
result, err := h.repo.Order.FindByID(ctx, id)
if err != nil {
    return errors.Wrap(ctx, err)
}

// ドメインルール違反は意味のあるエラーで返す
if !order.CanCancel() {
    return errors.MakeBusinessError(ctx, "order cannot be canceled")
}
```

## ログ出力のタイミング

- エラーのログは `errors.Make*Error` と `errors.Wrap` の内部で出力される。**呼び出し側で重複ログを出さない**
- 同じ失敗イベントについて `logger.Warn`/`logger.Error` を呼んでから `Make*Error`/`Wrap` を呼ぶ（＝同じ失敗が2行ログに出る）のは重複ログであり禁止。**分類情報（ステータスコード等）を残したい場合は `logger.Warn` を別に呼ぶのではなく、`Make*Error` の可変長 `attrs` 引数（例: `MakeUnavailableError(ctx, err, attrs...)`）で渡す。** これは `ergo.Wrap` の attrs としてエラーに埋め込まれ、`Make*Error`/`Wrap` 内部のログ出力に自動的に乗る
- 追加のコンテキスト情報（注文IDなど、そのエラー固有ではなく呼び出し元の状況を示す情報）が必要な場合は `logger.Warn` で補足してよい。**このとき機密情報フィールドを渡さない**

```go
// 分類情報はMake*Errorのattrsで渡す（同一失敗イベントについて別行のログを出さない）
if err != nil {
    return errors.MakeUnavailableError(ctx, errors.ErrExternalUnavailable, slog.String("kind", "payment"))
}

// 呼び出し元固有の補足情報は logger.Warn を別途呼んでよい（重複ログではない：
// 「order save failed」というイベント自体は一度しかログされていない）
if err := h.repo.Order.Save(ctx, order); err != nil {
    logger.Warn(ctx, "order save failed", slog.String("order_id", order.ID))
    return errors.Wrap(ctx, err)
}
```

機密情報・個人情報を含みうるフィールド（パスワード、トークン、メールアドレス、決済情報など）は `pkg/logger` でマスキングする（または最初からログに渡さない）。マスキングを実装する場合は、回帰テストを最初に書く。後回しにしない。

## sentinel エラーの判定

- エラー種別の判定は `errors.Is(err, target)` を使う（`errors` パッケージの再エクスポート版）
- `ErrRecordNotFound` など定義済み sentinel は `pkg/errors/errors.go` に追加する

```go
if errors.Is(err, errors.ErrRecordNotFound) {
    // not found として扱う
}
```

## handler でのエラー変換

- Echo handler では `errors.GetCode(err)` で HTTP ステータスを取得してレスポンスに変換する
- 変換は `internal/server/` の `ErrorHandler` ミドルウェア（または `echo.HTTPErrorHandler`）に集約し、各ハンドラで個別に書かない
- **クライアントへ返すメッセージにも機密情報を含めない**

```go
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := errors.GetCode(err)
    _ = c.JSON(code, response.Error{Message: errors.PublicMessage(err)})
}
```

## ユーザー体験としてのエラー

- 処理失敗時はユーザーの入力・操作を失わない設計にする（例: フォーム送信失敗時に入力内容を保持する）。可能な限り「やり直せる」状態で返す
- リトライ可能なエラー（外部サービスの一時障害等）はその旨をクライアントに伝える
- **エラーを課金・アップセルの動機に使わない。** 「失敗しました。有料プランなら…」のような文言を返さない

## 意図的に失敗を伝搬させない場合（graceful degradation）

「補助的な処理が失敗してもユーザーを待たせず、機能を諦めて先に進む」設計判断自体は正しい場合がある（例: おすすめ商品の取得に失敗しても商品一覧は表示する）。

ただし、**戻り値の型に `error` を持つ関数が「実際には決して non-nil を返さない」設計にする場合、「握りつぶした」で終わらせず、必ず失敗を観測できる形で残す。**

- ログは呼び出し元で `logger.Warn` を直書きするのではなく、下位の呼び出し（`errors.Wrap` / `Make*Error` 経由）で既に出力済みならそれに任せ、**呼び出し元で重複ログを出さない**
- 下位の呼び出しがまだログしていない失敗（JSONパース失敗など、外部SDK呼び出しではないローカルな失敗）は、`_ = errors.Wrap(ctx, err)` のように戻り値を握りつぶしていることを明示しつつ、ログだけは通す
- 関数のドキュメントコメントに「この関数は常に nil を返す設計である」ことと、**その理由** を明記する。理由なく `error` 型を持ちながら握りつぶす実装は、シグネチャと実装が矛盾しているとみなしレビューで指摘する

```go
// 正: 下位呼び出しで既にログ済みの失敗は再ログしない
raw, err := s.client.FetchRecommendations(ctx, userID)
if err != nil {
    // FetchRecommendations内でMakeUnavailableError（分類情報・スタックトレース付き）済み。
    // おすすめ取得の失敗でユーザーを待たせないため、空リストで先に進む。
    return nil, nil
}

// 正: ここでしか失敗が分からない場合はerrors.Wrapでログだけ通す
if _, err := parseRecommendations(raw); err != nil {
    _ = errors.Wrap(ctx, err) // ログのみ。呼び出し元へは伝搬させない設計
    return nil, nil
}
```

## nil チェック

- 関数冒頭で nil チェックし、早期 return する（ネストを深めない）
- `errors.MakeSystemError(ctx, nil)` は nil を返す。nil チェックなしで渡して良い
