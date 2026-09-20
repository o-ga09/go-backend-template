# パッケージ分割ルール

対象: `backend/`（Go / Echo / GORM）。

## 命名規則

- パッケージ名はディレクトリ名と一致させる（`internal/domain/auth/` → `package auth`）
- パッケージ名は単数形・小文字・短く（`users` でなく `user`）
- alias は衝突回避にのみ使う

## ディレクトリと役割

```
backend/
├── cmd/
│   ├── api/               # エントリポイント（Echo 起動・依存注入。Cloud Run / ECS 両対応）
│   └── migrate/           # マイグレーション実行（sql-migrate）
├── db/
│   └── migrations/        # スキーマ定義の正（sql-migrate 形式のSQL）
├── internal/
│   ├── domain/<name>/     # ドメインエンティティ + インターフェース定義
│   │   ├── entity.go      # 型定義・インターフェース・go:generate コメント
│   │   ├── <name>.go      # ドメインメソッド
│   │   └── mock/          # moq 自動生成モック
│   ├── handler/           # Echo ハンドラ（1ファイル=1リソース）
│   │   ├── request/       # リクエスト型
│   │   └── response/      # レスポンス型
│   ├── router/            # ルーティング定義
│   ├── server/            # サーバー起動・ミドルウェア
│   ├── database/
│   │   ├── mysql/         # GORM リポジトリ（1ファイル=1ドメイン）+ BaseModel自動挿入・楽観ロックの GORM プラグイン
│   │   └── transaction/   # ITransactionManager（DBエンジン非依存）
│   ├── external/<name>/   # 外部サービスクライアントのラッパー（決済・通知など。必要になったら追加）
│   ├── service/           # usecase層（複雑なオーケストレーションが必要な場合のみ。architecture.md参照）
│   └── crypto/            # 暗号化/復号（KMS等。必要になったら追加）
└── pkg/
    ├── config/            # 環境変数（Config 構造体、caarlos0/env）
    ├── context/           # Context ヘルパー（Get/Set 系）
    ├── errors/            # エラー型・生成関数（ergo ベース）
    ├── logger/            # ロガー（Cloud Logging 連携）
    ├── constant/          # 定数
    ├── uuid/              # ID生成
    ├── model/             # domain 共通の埋め込み構造体（BaseModel。新規ドメイン追加時に作成。architecture.md参照）
    └── ...                # その他ユーティリティ
```

現在の実装例は `internal/domain/auth/`。新規ドメイン（EC商材の商品・カート・注文など）もこの構成に従う。
新規ドメインは domain 構造体に `pkg/model.BaseModel` を埋め込む（`architecture.md`「domain＝DBモデル・BaseModel・楽観ロック」）。暗号化・カラム名詰め替えが必要なドメインは対象外。

## ファイル分割の基準

- 1 ファイルが 300 行を超えたら分割を検討する
- handler は **1 リソース = 1 ファイル**（`auth.go`, `order.go` など）
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

## interface の置き場所

- インターフェースは **domain 層**に定義する（`entity.go` 内の `IXxxRepository`）
- `mock/` サブパッケージのモックは `moq` で自動生成する

```go
// internal/domain/order/entity.go
//go:generate moq -out mock/order_repository_mock.go -pkg moq . IOrderRepository
type IOrderRepository interface {
    Get(ctx context.Context, id string) (*Order, error)
    Save(ctx context.Context, o *Order) error
}
```

**外部サービスクライアントも必ずインターフェース越しに使う。** テストで実APIを叩かないため（`testing.md`）。
