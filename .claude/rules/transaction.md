# トランザクションルール

対象: `backend/`（Go / GORM / TiDB Serverless）。

## 基本方針

- 複数テーブルへの書き込みを含む処理は必ず `ITransactionManager.RunInTx` でラップする
- 読み取り専用操作（SELECT）はトランザクション外で実行してよい
- バリデーションと入力チェックはトランザクション開始前に全件完了させる
- **LLM 呼び出しをトランザクション内に入れない**（下記）

## トランザクションを呼ぶ場所（新規コード）🔴

`architecture.md`「基本方針：レイヤードアーキテクチャ」と対応する。

- **シンプルな CRUD（usecase を作らない場合）は、ハンドラが `ITransactionManager.RunInTx` を直接呼ぶ。** usecase 層を作らない方針と整合させる
- **複雑なオーケストレーション（暗号化・LLM呼び出しを伴う既存の `internal/service/` など）は、従来通り service 層がトランザクションを保持してよい**（本方針以前からの例外）
- どちらの場合も、**LLM呼び出し・KMS呼び出しをトランザクション内に置かない**のは変わらず絶対（下記）

## LLM 呼び出しとトランザクションを重ねない 🔴

下書き生成・メモリ要約は数秒〜数十秒かかる（NFR-01-04：30秒以内）。
**この間トランザクションを開いたままにすると、TiDB の接続を長時間占有し、Cloud Run のスケール時に接続が枯渇する。**

```go
// 誤: LLM 呼び出しをトランザクションで囲む
return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    draft, err := s.llm.GenerateDraft(txCtx, materials)  // 30秒間トランザクションを保持
    ...
})

// 正: 生成 → その後に短いトランザクションで書き込む
draft, err := s.llm.GenerateDraft(ctx, materials)
if err != nil {
    return errors.Wrap(ctx, err)
}
return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    if err := s.draftRepo.Save(txCtx, draft); err != nil {
        return errors.Wrap(txCtx, err)
    }
    return s.memoryRepo.Update(txCtx, memory)
})
```

暗号化（`internal/crypto/`、Cloud KMS 呼び出し）も同様にトランザクション外で行う。

## インターフェースと実装の場所

- `ITransactionManager` は `internal/infra/database/transaction/transaction.go` で定義する
- `TransactionManager` はステートレスで、`*gorm.DB` は `ctx` から取得する（フィールドに持たない）
- モックは `moq` で自動生成し、`internal/infra/database/transaction/mock/` に置く
- 手書きスタブ禁止

## コンストラクタでの注入

```go
func NewSessionService(
    sessionRepo domainsession.ISessionRepository,
    answerRepo  domainanswer.IAnswerRepository,
    draftRepo   domaindraft.IDraftRepository,
    memoryRepo  domainmemory.IMemoryRepository,
    txManager   transaction.ITransactionManager,
) *SessionService { ... }
```

## RunInTx の実装パターン

```go
func (s *SessionService) Complete(ctx context.Context, sessionID string) error {
    // バリデーション（トランザクション外）
    sess, err := s.sessionRepo.FindByID(ctx, sessionID)
    if err != nil {
        return errors.Wrap(ctx, err)
    }
    if err := sess.CanComplete(); err != nil {
        return errors.MakeBusinessError(ctx, "session cannot be completed")
    }

    // DB書き込み（トランザクション内）
    return s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        if err := s.sessionRepo.Complete(txCtx, sess); err != nil {
            return errors.Wrap(txCtx, err)
        }
        return s.memoryRepo.Update(txCtx, memory)
    })
}
```

クロージャ外に結果を持ち出す場合は外側の変数を使う。

```go
var draftID string
err := s.txManager.RunInTx(ctx, func(txCtx context.Context) error {
    d := &draft.Draft{...}
    if err := s.draftRepo.Create(txCtx, d); err != nil {
        return errors.Wrap(txCtx, err)
    }
    draftID = d.ID
    return nil
})
```

### ハンドラから直接呼ぶ場合（シンプルな CRUD・usecase なし）

```go
func (h *xxxHandler) Update(c echo.Context) error {
    ctx := c.Request().Context()

    // バリデーション（トランザクション外）
    x, err := h.repo.FindByID(ctx, id)
    if err != nil {
        return errors.Wrap(ctx, err)
    }
    if err := x.CanUpdate(); err != nil {
        return errors.MakeBusinessError(ctx, "xxx cannot be updated")
    }

    // DB書き込み（トランザクション内）。楽観ロックは GORM プラグインが処理する
    if err := h.txManager.RunInTx(ctx, func(txCtx context.Context) error {
        return h.repo.Update(txCtx, x)
    }); err != nil {
        return errors.Wrap(ctx, err)
    }
    return c.JSON(http.StatusOK, response.FromXxx(x))
}
```

## TiDB 固有の注意

- TiDB は楽観的／悲観的トランザクションの挙動が MySQL と完全一致しない。**長時間トランザクション・大量行ロックを避ける**
- 同期の競合解決は last-write-wins でよい（`ADR-0007`：単一ユーザー・実質単一デバイス利用を想定）。悲観ロックで解決しようとしない
- Cloud Run（GCP）→ TiDB Cloud（AWS）は**クロスクラウド接続**でレイテンシが乗る。トランザクション内の往復回数を最小にする

## 禁止事項

- `gorm.DB.Transaction` / `gorm.DB.Begin` をハンドラ・サービス層で直接呼ばない。`ITransactionManager` 経由のみ
- `pkg/errors` 以外のエラーパッケージを使わない
- 手書きスタブでのモック（`moq` 生成を使う）
- バリデーションをトランザクション内に混在させない
- **LLM 呼び出し・KMS 呼び出しをトランザクション内に置かない**
- **GORM の `AutoMigrate` を使わない。** スキーマの正は `db/migrations/`（`sql-migrate`）
- **楽観ロック（`version`チェック）をアプリケーションコードで手書きしない。** GORM プラグインに委ねる（`architecture.md`「domain＝DBモデル・BaseModel・楽観ロック」）

## テストでのモック利用

```go
txMock := &transactionmock.ITransactionManagerMock{
    RunInTxFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
        return fn(ctx)  // テストでは fn をそのまま呼ぶ（トランザクションなし）
    },
}
svc := service.NewSessionService(sessionRepo, answerRepo, draftRepo, memoryRepo, txMock)
```

## トランザクションが不要なケース

- 単一テーブルへの書き込みのみ（**回答の1問ずつの保存はこれに当たる**。NFR-02-02：設問ごとに確定時点で永続化する）
- 読み取りのみ
- 外部 API 呼び出し（Anthropic / Cloud KMS / 運用通知）
