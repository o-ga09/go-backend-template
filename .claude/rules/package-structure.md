# パッケージ分割ルール

対象: `backend/`（Go / Echo / GORM）。

## 命名規則

- パッケージ名はディレクトリ名と一致させる（`internal/domain/oshi/` → `package oshi`）
- パッケージ名は単数形・小文字・短く（`oshis` でなく `oshi`）
- alias は衝突回避にのみ使う

## ディレクトリと役割

```
backend/
├── cmd/
│   ├── api/               # Cloud Run エントリポイント（Echo 起動・依存注入）
│   └── migration/         # マイグレーション実行（sql-migrate。既存）
├── db/
│   └── migrations/        # スキーマ定義の正（sql-migrate 形式のSQL）
├── internal/
│   ├── domain/<name>/     # ドメインエンティティ + インターフェース定義
│   │   ├── entity.go      # 型定義・インターフェース・go:generate コメント
│   │   ├── <name>.go      # ドメインメソッド
│   │   └── mock/          # moq 自動生成モック
│   ├── server/            # Echo ハンドラ（1ファイル=1リソース）
│   │   ├── request/       # リクエスト型
│   │   ├── response/      # レスポンス型
│   │   ├── server.go      # サーバー起動・依存注入
│   │   └── route.go       # ルーティング定義
│   ├── database/
│   │   └── mysql/         # GORM リポジトリ（1ファイル=1ドメイン）+ BaseModel自動挿入・楽観ロックの GORM プラグイン
│   ├── external/
│   │   └── anthropic/     # Anthropic 公式 Go SDK のラッパー
│   ├── crypto/            # Cloud KMS envelope encryption
│   └── infra/database/transaction/  # ITransactionManager
└── pkg/
    ├── config/            # 環境変数（Config 構造体、caarlos0/env）
    ├── context/           # Context ヘルパー（Get/Set 系）
    ├── errors/            # エラー型・生成関数（ergo ベース）
    ├── logger/            # ロガー（Cloud Logging / マスキング）
    ├── model/              # domain 共通の埋め込み構造体（BaseModel。architecture.md 参照）
    ├── trace/             # trace_id抽出・生成 + OpenTelemetry TracerProvider構築（Cloud Traceエクスポーター。ADR-0012）
    ├── testutil/          # テスト用ユーティリティ
    └── ...                # その他ユーティリティ（ulid, ptr など）
```

ドメインは `oshi` / `session` / `answer` / `draft` / `memory` から始める。
新規ドメインは domain 構造体に `pkg/model.BaseModel` を埋め込む（`architecture.md`「domain＝DBモデル・BaseModel・楽観ロック」）。暗号化・カラム名詰め替えが必要な既存ドメイン（`oshi` 等）は対象外。

## `internal/external/anthropic/` の分割

LLM 呼び出しは **2箇所のみ**（`ADR-0006`：エージェントフレームワークは不採用、単発呼び出しのみ）。

| ファイル | 役割 |
|---|---|
| `client.go` | Anthropic Go SDK のラッパー。インターフェースを定義し、SDK 型を外へ漏らさない |
| `dig.go` | 「特にない」の意味的判定（FR-04-02）。掘る文は**静的テンプレート**でLLM生成しない |
| `draft.go` | 下書き生成。`docs/generation-prompt.md` の埋め込み |
| `memory.go` | メモリ要約生成 |
| `prompt/` | プロンプト定数。**`docs/` の文言と1:1で対応させ、コード側で改変しない**（FR-03-04 / FR-06-02） |

**プロンプトを変えたくなったら、先に `docs/` を改訂してから同期する。** 逆順は禁止。

## ファイル分割の基準

- 1 ファイルが 300 行を超えたら分割を検討する
- server ハンドラは **1 リソース = 1 ファイル**（`oshi.go`, `session.go`, `draft.go`, `memory.go`）
- mysql リポジトリは **1 ドメイン = 1 ファイル**

## 新しいパッケージを作るとき

以下の順序で判断する:

1. 既存パッケージに追加できないか検討する
2. ドメインロジックなら `internal/domain/<name>/` に新ドメインを作る
3. 外部サービスクライアントなら `internal/external/<name>/` に追加し、インターフェースを domain 層に置く
4. 汎用ユーティリティなら `pkg/<name>/` に置く
5. `cmd/` 配下は main パッケージのみ。ロジックを書かない

## 禁止パターン

- `utils/`, `helpers/`, `common/` のような曖昧なパッケージ名は作らない
- 循環 import 禁止（ビルドエラーになる）
- グローバル変数でインスタンスを共有しない。コンストラクタで依存注入する
- `init()` 関数の使用禁止
- **`internal/external/anthropic/` を domain 層から import しない。** 生成はユースケース側の関心事
- **プロンプト文字列をハンドラやリポジトリに直書きしない。** `internal/external/anthropic/prompt/` に集約する

## interface の置き場所

- インターフェースは **domain 層**に定義する（`entity.go` 内の `IXxxRepository`）
- `mock/` サブパッケージのモックは `moq` で自動生成する

```go
// internal/domain/draft/entity.go
//go:generate moq -out mock/draft_repository_mock.go -pkg moq . IDraftRepository
type IDraftRepository interface {
    Get(ctx context.Context, sessionID string) (*Draft, error)
    Save(ctx context.Context, d *Draft) error
}

//go:generate moq -out mock/draft_generator_mock.go -pkg moq . IDraftGenerator
type IDraftGenerator interface {
    Generate(ctx context.Context, m *Materials) (*Draft, error)
}
```

**LLM クライアントも必ずインターフェース越しに使う。** テストで実APIを叩かないため（`testing.md`）。
