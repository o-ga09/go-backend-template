# エラーハンドリングルール

対象: `backend/`（Go / Echo / GORM）。

## 基本原則

エラーは `pkg/errors` パッケージを経由して扱う。`fmt.Errorf` や標準 `errors.New` を直接使わない。

このパッケージは **`github.com/newmo-oss/ergo`** をベースに構築されている。`ergo` はスタックトレース付きエラー・sentinel エラー・エラーコードを統合した Go エラーライブラリ。

```go
// 誤
return fmt.Errorf("session not found: %w", err)

// 正
return errors.MakeNotFoundError(ctx, "session not found")
```

## ergo の使い方方針

`ergo` を直接 import するのは `pkg/errors/errors.go` のみ。それ以外のパッケージは `pkg/errors` 経由で使う。

### sentinel エラーの定義

新しいドメインエラーは `ergo.NewSentinel` で定義し、`pkg/errors/errors.go` に追加する（`ergo.New` ではない。sentinel はパッケージ変数として１回しか作られないため、`ergo.New` を使うとパッケージ初期化時点の無意味なスタックトレースが固定で付与されてしまう。`ergo.NewSentinel` はスタックトレースを持たない設計になっている）。

```go
// errors.go に追加
var ErrDigDeeperLimit = ergo.NewSentinel("dig deeper limit reached")   // 原則4
var ErrFreePlanLimit  = ergo.NewSentinel("free plan limit reached")   // FR-08-02
var ErrLLMUnavailable = ergo.NewSentinel("llm unavailable")
```

### エラーのラップ（ergo.Wrap）

下位レイヤーのエラーに文脈を付けるとき。スタックトレースが自動付与される。

```go
// ergo.Wrap は内部で使う。外部からは errors.Wrap(ctx, err) を呼ぶ
err = ergo.Wrap(ErrRecordNotFound, "session lookup failed")
```

### エラーコードの付与（ergo.WithCode）

HTTP ステータスに対応するコードを付与するとき。`Make*Error` 関数が内部で自動付与するため、通常は直接呼ばない。

```go
err = ergo.WithCode(err, ErrCodeNotFound) // Make*Error 内部で行う
```

### スタックトレースの取得（ergo.StackTraceOf）

`pkg/errors` 内部の `errorLogAttrs`（`logError` / `logByStatus` から呼ばれる）が `Make*Error` / `Wrap` のたびに自動で `stacktrace` 属性として出力する。呼び出し側で直接呼ぶ必要はない。

`ergo.New` / `ergo.Wrap` はその場で `runtime.Callers` によるスタックトレースを記録する（`ergo.NewSentinel` で作った sentinel には無い）。これは「そのエラーが作られた行」だけでなく、**その時点のゴルーチンの呼び出し履歴（呼び出し元の呼び出し元…）を丸ごと含む**ため、`stacktrace` 属性だけで発生元とそこに至った呼び出し経路の両方を特定できる。ログに `error: "...: llm unavailable"` のような短い文字列しか出ていない場合、多くは `Wrap`/`Make*Error` に渡した `err` が sentinel（`ErrXxx = ergo.NewSentinel(...)`）そのもので、下位レイヤーの実エラーを経由していないことが原因。**握りつぶさず、実際に失敗した `err` を渡すこと**（次節）。

## 最優先の制約：エラーに平文を載せない 🚫

このプロダクトが扱うのは推しへの本心。**エラーメッセージはログ・監視・運用通知チャネルへそのまま流れる**（`NFR-03-05` / `NFR-03-10` / `FR-16-92`）。
`ergo` はスタックトレースを自動付与するため、**エラーに載せた文字列は確実に外へ出る**前提で書く。

```go
// 誤: 本文・回答がエラーメッセージに入る
return errors.MakeBusinessError(ctx, fmt.Sprintf("invalid answer: %s", answerText))

// 誤: 推しの名前が入る
return errors.MakeNotFoundError(ctx, "oshi not found: "+oshi.Name)

// 正: 識別子と種別だけ
return errors.MakeBusinessError(ctx, "answer is empty")
return errors.MakeNotFoundError(ctx, "oshi not found")
```

**載せてよいもの**：ID、質問キー（`q1`〜`q5`）、件数、ステータス、エラー種別。
**載せてはいけないもの**：手紙本文、インタビュー回答、メモリの中身、推しの名前、LLM のリクエスト／レスポンス本体。

LLM 呼び出しの失敗をラップするときは、**プロンプトとレスポンスを含めない**。ただし「含めない」は「エラー情報ごと握りつぶす」ことではない。SDK の `err.Error()` を直接ログに出さないことと、失敗の原因（タイムアウトか・レート制限か・認証エラーか）を診断可能な形で残すことは両立できる。**握りつぶすと、本番障害の一次切り分けにアプリログが使えなくなる**（実際に Vertex AI クォータ枯渇による障害で、アプリログには `kind` しか残っておらず、Vertex API に直接 curl するまで原因不明だったインシデントがあった）。

```go
// 誤①: SDK のエラーをそのままラップする（err.Error() にAPIレスポンスの生JSONが
// 含まれうる。例: anthropic-sdk-go の apierror.Error.Error() はレスポンスボディを含む）
if err != nil {
    return errors.Wrap(ctx, err)
}

// 誤②: 実際に発生したエラー(err)を最後まで一度も使わず、無関係な sentinel だけを
// 返す。err.Error() は避けられているが、ステータスコードやエラー種別も一緒に
// 消えてしまい、ログからは「LLM呼び出しが失敗した」以上の情報が一切取れない
// （error-handling.md 旧版が推奨していたパターン。stacktrace も sentinel には
// 無いため呼び出し元まで辿れなくなる）
if err != nil {
    logger.Warn(ctx, "llm call failed", slog.String("kind", "draft"))
    return errors.MakeUnavailableError(ctx, errors.ErrLLMUnavailable)
}

// 正: err自体は本文を含む可能性があるためログに出さないが、本文を含まない
// 安全なフィールド（HTTPステータスコード・SDKのエラー種別・リクエストIDなど）
// だけを抽出し、attrs として Make*Error に渡す。err自体は sentinel を返すため
// クライアントへは相変わらず種別化されたメッセージしか返らない。
// ログ出力はMakeUnavailableError内部で一度だけ行われる（呼び出し側で
// logger.Warn を別途呼ばない。「ログ出力のタイミング」参照）。
if err != nil {
    var apiErr *anthropicsdk.Error
    attrs := []slog.Attr{slog.String("kind", "draft")}
    if errors.As(err, &apiErr) {
        attrs = append(attrs,
            slog.Int("llm_status_code", apiErr.StatusCode),
            slog.String("llm_error_type", string(apiErr.Type())),
            slog.String("llm_request_id", apiErr.RequestID),
        )
    }
    return errors.MakeUnavailableError(ctx, errors.ErrLLMUnavailable, attrs...)
}
```

同じ考え方は LLM 以外の外部 SDK（DB ドライバ・KMS・ストア連携 API 等）にも適用する。**「本文が混入しうるので raw error を出さない」と「失敗の原因を一切残さない」はイコールではない。** 安全に取り出せる分類情報（ステータスコード・エラーコード・型名など）まで一緒に捨てないこと。

## エラー種別と生成関数

| 状況 | 使う関数 | HTTP ステータス相当 |
|---|---|---|
| 認証失敗（トークン無効など） | `errors.MakeAuthorizedError(ctx, msg)` | 401 |
| 認可失敗（権限不足・有料機能への無料アクセス） | `errors.MakeAuthorizationError(ctx, msg)` | 403 |
| リソース未発見 | `errors.MakeNotFoundError(ctx, msg)` | 404 |
| 競合（重複登録など） | `errors.MakeConflictError(ctx, msg)` | 409 |
| バリデーション違反 | `errors.MakeBusinessError(ctx, msg)` | 422 |
| 予期しない内部エラー | `errors.MakeSystemError(ctx, err)` | 500 |
| 頻度制限（生成要求の再試行制限等） | `errors.MakeRateLimitError(ctx, msg)` | 429 |
| 外部サービス（LLM等）の一時的な障害 | `errors.MakeUnavailableError(ctx, err, attrs...)` | 503 |

MVP は認証を持たない（`ADR-0006`：RevenueCat の匿名 App User ID のみ）ため、**401 は当面使わない**。
権利（Entitlement）による分岐は 403（`MakeAuthorizationError`）で表現する。

## エラーのラップとログ

- **ラップのみ**: 詳細不明の外部エラーは `errors.Wrap(ctx, err)` でラップする（ログも自動出力）
- **新規生成**: ドメインルール違反は `errors.Make*Error` で意味のあるエラーを生成する
- **伝搬**: 一度ラップしたエラーは再ラップしない。`errors.IsWrapped(err)` で確認してから `Wrap` を呼ぶ

```go
// 外部エラーをラップして伝搬（ログはWrap内で出力済み）
result, err := h.repo.Session.FindByID(ctx, id)
if err != nil {
    return errors.Wrap(ctx, err)
}

// ドメインルール違反は意味のあるエラーで返す
if !owner.HasEntitlement() {
    return errors.MakeAuthorizationError(ctx, "この機能は有料プランで利用できます")
}
```

## ログ出力のタイミング

- エラーのログは `errors.Make*Error` と `errors.Wrap` の内部で出力される。**呼び出し側で重複ログを出さない**
- 同じ失敗イベントについて `logger.Warn`/`logger.Error` を呼んでから `Make*Error`/`Wrap` を呼ぶ（＝同じ失敗が2行ログに出る）のは重複ログであり禁止。**分類情報（ステータスコード等）を残したい場合は `logger.Warn` を別に呼ぶのではなく、`Make*Error` の可変長 `attrs` 引数（例: `MakeUnavailableError(ctx, err, attrs...)`）で渡す。** これは `ergo.Wrap` の attrs としてエラーに埋め込まれ、`Make*Error`/`Wrap` 内部のログ出力に自動的に乗る
- 追加のコンテキスト情報（session ID, 質問キーなど、そのエラー固有ではなく呼び出し元の状況を示す情報）が必要な場合は `logger.Warn` で補足してよい。**このとき本文系フィールドを渡さない**

```go
// 分類情報はMake*Errorのattrsで渡す（同一失敗イベントについて別行のログを出さない）
if err != nil {
    return errors.MakeUnavailableError(ctx, errors.ErrLLMUnavailable, slog.String("kind", "draft"))
}

// 呼び出し元固有の補足情報は logger.Warn を別途呼んでよい（重複ログではない：
// 「draft save failed」というイベント自体は一度しかログされていない）
if err := h.repo.Draft.Save(ctx, draft); err != nil {
    logger.Warn(ctx, "draft save failed", slog.String("session_id", draft.SessionID))
    return errors.Wrap(ctx, err)
}
```

`pkg/logger` は本文系フィールド名（`answer_text`, `content`, `memory`, `oshi_name` など）を**マスキング対象として持つ**。
マスキングの回帰テストを最初に書く（`ADR-0007` 帰結欄・`NFR-03-05`）。後回しにしない。

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
- 変換は `internal/server/` の `echo.HTTPErrorHandler` に集約し、各ハンドラで個別に書かない
- **クライアントへ返すメッセージにも平文を含めない**

```go
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := errors.GetCode(err)
    _ = c.JSON(code, response.Error{Message: errors.PublicMessage(err)})
}
```

## ユーザー体験としてのエラー

要件上、エラー時に**入力を失わないことが機能要件**（FR-06-11 / NFR-02-03）。

- 生成失敗は「やり直せる」状態で返す。回答は永続化済みであること
- 生成基盤の障害はリトライ可能であることをクライアントに伝える（NFR-02-06）
- **エラーを課金の動機に使わない**（原則10 / FR-10-91）。「失敗しました。有料プランなら…」のような文言を返さない

## 意図的に失敗を伝搬させない場合（graceful degradation）

NFR-01-03a のように「この処理が失敗してもユーザーを待たせず、機能を諦めて先に進む」設計判断自体は正しい（例: `internal/external/anthropic/dig.go` の意味的判定は、失敗時に `(false, nil)` を返してユーザー体験を止めない）。

ただし、**戻り値の型に `error` を持つ関数が「実際には決して non-nil を返さない」設計にする場合、「握りつぶした」で終わらせず、必ず失敗を観測できる形で残す。**

- ログは呼び出し元で `logger.Warn` を直書きするのではなく、下位の呼び出し（`errors.Wrap` / `Make*Error` 経由）で既に出力済みならそれに任せ、**呼び出し元で重複ログを出さない**
- 下位の呼び出しがまだログしていない失敗（JSONパース失敗など、外部SDK呼び出しではないローカルな失敗）は、`_ = errors.Wrap(ctx, err)` のように戻り値を握りつぶしていることを明示しつつ、ログだけは通す
- 関数のドキュメントコメントに「この関数は常に nil を返す設計である」ことと、**その理由（NFR/原則の参照）** を明記する。理由なく `error` 型を持ちながら握りつぶす実装は、シグネチャと実装が矛盾しているとみなしレビューで指摘する

```go
// 正: 下位呼び出しで既にログ済みの失敗は再ログしない
raw, err := j.client.completeTool(ctx, "dig", system, user, tool)
if err != nil {
    // completeTool内でMakeUnavailableError（分類情報・スタックトレース付き）済み。
    // NFR-01-03a：判定失敗はユーザーを待たせないため「掘り下げなし」に倒す。
    return false, nil
}

// 正: ここでしか失敗が分からない場合はerrors.Wrapでログだけ通す
if _, err := parseNeedsDigDeeper(raw); err != nil {
    _ = errors.Wrap(ctx, err) // ログのみ。呼び出し元へは伝搬させない設計（NFR-01-03a）
    return false, nil
}
```

## nil チェック

- 関数冒頭で nil チェックし、早期 return する（ネストを深めない）
- `errors.MakeSystemError(ctx, nil)` は nil を返す。nil チェックなしで渡して良い
