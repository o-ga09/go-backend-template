# アーキテクチャルール

対象: `backend/`（Go / Echo / GORM）。クライアント（Next.js）は `frontend.md`。

技術選定の根拠は `docs/decisions/ADR-0006`（技術スタック）・`ADR-0007`（保存方式・インフラ）。
**本ルールは実装の作法を決めるものであり、ADR の決定を上書きしない。**

## 全体構成

```
[RN/Expo クライアント]  ローカルSQLite（オフライン閲覧・無料範囲の最新1件）
        │ HTTPS
[Echo / Cloud Run (asia-northeast1)]
        │ GORM
[TiDB Serverless (ap-northeast-1) / MySQL互換]
```

LLM 呼び出しは**必ずバックエンド側**で行う。APIキー・プロンプトテンプレートをクライアントに含めない（`ADR-0006`）。

## 基本方針：レイヤードアーキテクチャ 🔴

**このセクションは今後の新規コードに適用する。** `internal/service/`（`oshi`/`session`/`answer`/`draft`/`memory`）は本方針の策定以前に実装済みで、暗号化・LLM呼び出しを伴う複雑なオーケストレーションを行うため、旧方針（service層必須・トランザクションを service 層が保持）のまま許容する既存の例外として扱う。無理に本方針へ合わせて作り直さない。新規ドメイン・新規エンドポイントから以下を適用する。

基本形は2層（`server` → `domain`）。usecase 層（`internal/service/` 相当）を挟むのは、複数リポジトリ・暗号化・LLM呼び出しなど**複雑なオーケストレーションが必要な場合のみ**の例外。

```
[シンプルな CRUD（デフォルト）]
server (handler) → domain（repository interface + ドメインロジック）

[複雑なオーケストレーション（暗号化・LLM呼び出し・複数ドメイン合成など。例外）]
server (handler) → service (usecase) → domain
```

- **シンプルな CRUD に usecase を作らない。** ハンドラが直接 domain のリポジトリを呼ぶ
- **ドメインロジック（バリデーション・状態遷移の可否判定）は必ず domain 層の純粋関数に置く。** ハンドラ・usecase はそれを呼び出すだけにする
- **domain のモデルは DB のモデルを兼ねる。** GORM 用の変換専用構造体を作らない。ただし暗号化やカラム名の詰め替えが必要なドメイン（例: 既存の `oshi` の `secret.Sealed` / `Since`⇄`history`）は、既存の分離設計を維持してよい（本方針以前からの例外。詳細は下記「domain＝DBモデル・BaseModel」）
- リポジトリ interface は domain 層に置く（`entity.go` の `go:generate moq`）。usecase を挟む場合も、リポジトリ interface の置き場所はドメイン層のまま変わらない

## レイヤー構成と依存方向

依存は必ず内側に向ける。外側のレイヤーが内側を知る。内側は外側を知らない。

```
[シンプルな CRUD]
server            → domain, database/mysql, infra/database/transaction （配線と直接呼び出し）
database/mysql    → domain   (ドメイン型を使うが、ドメインロジックを持たない)

[複雑なオーケストレーション（既存 internal/service/ など）]
server            → service
service           → domain
service           → crypto, infra/database/transaction
external/anthropic → domain  (プロンプト生成の入出力にドメイン型を使う)
crypto            → 依存なし  (暗号化/復号のみ。ドメインを知らない)
```

- `domain/` は `server/`, `service/`, `database/`, `external/` を import しない
- **`internal/service/` は複雑なオーケストレーションが必要な場合のみのユースケース層。** ハンドラ（`server/`）から呼ばれ、複数のリポジトリ・`crypto.ISealer`・`ITransactionManager`・LLMインターフェースを組み合わせる。ドメインルール自体（バリデーション・状態遷移の可否判定）はドメイン層の純粋関数に置き、service はそれを呼び出すだけにする
- `database/mysql/` はビジネスロジックを持たない（データの読み書きのみ）
- `external/anthropic/` は Anthropic Go SDK をラップする。**SDK 型を domain や server に漏らさない**
- **暗号化・復号はトランザクションの外で行う。** `database/mysql/` のリポジトリ内で `crypto.ISealer` を呼ばない（`transaction.md`「LLM呼び出しとトランザクションを重ねない」と同じ理由で、KMS呼び出しもトランザクション内に置かない）。ドメインエンティティが保持するのは暗号文（`pkg/secret.Sealed`）で、平文（`pkg/secret.Text`）は呼び出し元（ハンドラ、または既存の service 層）のローカル変数としてのみ生存する
- oshi ドメインが権利（entitlement）を参照する必要がある場面（FR-02-03の2件目登録拒否等）は、**ドメイン間に新しい矢印を作らない。** 呼び出し側（ハンドラ、または既存の service 層）が `entitlement.IEntitlementChecker` を呼び、判定結果（bool）だけを対象ドメインの純粋関数（例: `oshi.CanRegisterMore(count, entitled)`）に渡す

## ドメイン層のルール

- エンティティは `internal/domain/<ドメイン名>/entity.go` に置く
- バリデーションロジックはドメイン層のメソッドとして実装する
- ドメイン間の依存は最小限にする。循環 import 禁止

```go
// 正: ドメイン型のメソッドとしてバリデーション
func (s *Session) CanAdvance() error { ... }

// 誤: handler 層でバリデーション
if len(s.Answers) >= 5 { ... }
```

主要ドメイン: `oshi`（推し）, `session`（インタビューセッション）, `answer`（回答）, `draft`（下書き）, `memory`（メモリ）。

## domain＝DBモデル・BaseModel・楽観ロック 🔴

**このセクションは今後の新規ドメインに適用する。** 暗号化やカラム名の詰め替えが必要な既存ドメイン（`oshi` 等）は対象外とし、従来通りリポジトリで手動マッピングする。

- `ID` / `Version` / `CreatedAt` / `UpdatedAt` など必須カラムは、共通の `BaseModel` を domain 構造体に埋め込んで持つ

  ```go
  // pkg/model/base.go
  type BaseModel struct {
      ID        string
      Version   int
      CreatedAt time.Time
      UpdatedAt time.Time
  }

  // internal/domain/xxx/entity.go
  type Xxx struct {
      model.BaseModel
      // ドメイン固有フィールド
  }
  ```

- 監査項目（作成者など）が必要なドメインは `BaseModel` に `CreatedUserID` 等を追加する。個別ドメインに都度同じフィールドを生やさない
- **`BaseModel` のカラム（ID採番・`CreatedAt`/`UpdatedAt`・`Version`のインクリメント）は GORM プラグイン（`Callbacks`）で自動挿入する。** リポジトリや変換関数の中で代入しない
- **楽観ロック（`Version`チェック）は GORM のプラグイン機構で実装する。** `WHERE version = ?` をリポジトリに手書きしたり、アプリケーションコードで比較・分岐を書かない。競合時は GORM が返すエラーをリポジトリ層で判別し、`errors.MakeConflictError` に変換する

これらは domain がそのまま GORM モデルとして永続化される前提（domain＝DBモデル）で成り立つ。GORM プラグイン本体は `internal/database/mysql/` に置く。

## データベース層のルール

- DB アクセスは `internal/database/mysql/` 配下の GORM リポジトリ経由のみ
- 各ドメインの CRUD は個別ファイルに分離する（`oshi.go`, `session.go`, `draft.go` など）
- 外部サービスクライアントはインターフェース越しに呼ぶ
  - Anthropic API → `internal/external/anthropic/`
  - Cloud KMS → `internal/crypto/`
- **マイグレーションは GORM の AutoMigrate を使わない。** `db/migrations/` の SQL と `cmd/migration`（`sql-migrate`）で管理する。スキーマの正はマイグレーションファイル（`BaseModel` の各カラムも例外なくマイグレーションSQLで定義する）

## このプロダクト固有の禁止事項 🚫

要件定義（`docs/requirements.md`）とADRから来る制約。**アーキテクチャの段階で守る。**

- **送信フラグ・送信日時をモデルに持たせない**（原則8 / FR-09-90）。カラムもフィールドも作らない。後から消すのではなく、最初から存在させない
- **手紙本文・インタビュー回答・メモリをログ／トレース／エラーメッセージに出さない**（NFR-03-05）。これらを含む構造体に `String()` や `LogValue()` を素朴に実装しない
- **平文の保存禁止。** 本文系カラムは `internal/crypto/` で暗号化してから永続化する。復号は LLM 呼び出しの直前のみ、処理後は再暗号化する（`ADR-0007`）
- **推しをまたいだデータの混在を型で防ぐ。** メモリ・セッションの取得は必ず `oshi_id` で絞る（FR-02-91）

## 設定・初期化のルール

- 環境変数は `pkg/config/config.go` の `Config` 構造体で一元管理する（`caarlos0/env`）
- `Config` は起動時に context に格納し、`config.GetCtxEnv(ctx)` で取得する
- DB クライアント・LLM クライアントはコンストラクタで直接注入する
- **Cloud Run はスケール0から起動するため、DB接続が急増しうる。** `gorm.DB` から `*sql.DB` を取り出して `SetMaxOpenConns` / `SetMaxIdleConns` / `SetConnMaxLifetime` を起動時に必ず設定する（`ADR-0007`）

## 依存注入のルール

コンストラクタで受け取った依存はメソッド内で nil チェックしない。

```go
// 誤: メソッド内に nil ガードを書く
func (h *draftHandler) Generate(c echo.Context) error {
    if h.llm == nil {
        return nil
    }
    ...
}

// 正: コンストラクタで依存を受け取り、メソッドはそのまま使う
func (h *draftHandler) Generate(c echo.Context) error {
    res, err := h.llm.GenerateDraft(ctx, materials)
    ...
}
```

- `if h.repo != nil`, `if h.client != nil` のような防御的チェックを実装メソッドに書かない
- 必須の依存はコンストラクタ引数で明示し、nil を渡せないようにする
- テスト用モックはインターフェースから `moq` で自動生成したものを使う
