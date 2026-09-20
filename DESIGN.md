# DESIGN.md

このリポジトリのアーキテクチャ全体像と、その設計判断の理由をまとめたドキュメント。実装の詳細な作法（禁止事項・書き方）は [`.claude/rules/`](.claude/rules/) に、セットアップ手順は [README.md](README.md) に譲る。

## 目的

このリポジトリは、実運用に耐える構成を最初から示す Web アプリケーションテンプレート。認証・認可付きの EC 商材サンプル（商品・カート・注文）を題材に、レイヤー構成・エラーハンドリング・トランザクション管理・テスト方針の「型」を提供し、新規プロジェクト立ち上げ時にサンプル実装を骨組みへ置き換えるだけで使える状態を目指す。

## 全体構成

```
[Next.js クライアント]
        │ HTTPS
[Echo（Cloud Run または ECS）]
        │ GORM
[MySQL]
```

- クライアントは Next.js（App Router）1つのみ。TanStack Query でサーバー状態を管理し、認証は NextAuth（Google OAuth）
- サーバーは Go（Echo）。Cloud Run（スケール0対応）と ECS のどちらでも動作する前提で `pkg/config` に環境変数を一元化
- DB は MySQL。スキーマの正は `db/migrations/`（`sql-migrate`）。GORM の `AutoMigrate` は使わない

## レイヤー構成の設計判断

基本形は2層（`server` → `domain`）。usecase 層（`internal/service/`）は例外であり、以下のいずれかに該当する場合のみ導入する。

- 複数リポジトリをまたぐオーケストレーションが必要
- 外部API呼び出し（決済・通知など）を伴う
- 暗号化/復号（KMS呼び出し）を伴う

**なぜ2層をデフォルトにするか**: シンプルな CRUD に usecase 層を強制すると、ハンドラ→usecase→domain の3層すべてが「呼ぶだけ」の薄いパススルーになりやすく、変更のたびに3ファイルを触る割に得られる恩恵が小さい。オーケストレーションが必要になった時点で初めて usecase 層を挟むことで、複雑さに見合った層構成を保つ。

ドメインロジック（バリデーション・状態遷移の可否判定）は必ず domain 層の純粋関数に置く。ハンドラ・usecase はそれを呼び出すだけにする。詳細は [`.claude/rules/architecture.md`](.claude/rules/architecture.md) を参照。

## domain＝DBモデルという前提

domain の構造体がそのまま GORM モデルとして永続化される（変換専用構造体を作らない）。これにより：

- ドメイン層とデータアクセス層の間でマッピングコードを書く必要がなくなる
- `BaseModel`（ID・Version・CreatedAt・UpdatedAt）を埋め込むだけで、採番・タイムスタンプ・楽観ロックを GORM プラグインに委譲できる

暗号化やカラム名の詰め替えが必要なドメインだけ、この前提から外れて手動マッピングする。詳細は [`.claude/rules/architecture.md`](.claude/rules/architecture.md)「domain＝DBモデル・BaseModel・楽観ロック」を参照。

## トランザクションと外部API呼び出しの分離

外部API呼び出し（決済・通知・KMSでの暗号化）とDBトランザクションは重ねない。理由は、応答時間が読めない外部呼び出しの間 DB 接続を占有すると、Cloud Run のスケール0起動時などに接続が枯渇しやすいため。呼び出し順序は「外部API呼び出し（トランザクション外）→ 短いトランザクションでの書き込み」に固定する。詳細は [`.claude/rules/transaction.md`](.claude/rules/transaction.md) を参照。

## エラーハンドリングの設計判断

`pkg/errors`（`ergo` ベース）に統一し、`fmt.Errorf` / 標準 `errors.New` を直接使わない。狙いは：

- エラーコード（`ergo.WithCode`）とスタックトレースを常に一貫した形で持たせ、`errors.GetCode(err)` から HTTP ステータスへ機械的に変換できるようにする
- エラーメッセージに機密情報（パスワード・トークン・個人情報・決済情報）を載せない運用を、レビューで検出しやすい形（`Make*Error` 関数の使用箇所を見るだけで判断できる）にする

詳細は [`.claude/rules/error-handling.md`](.claude/rules/error-handling.md) を参照。

## セキュリティ

- シークレット検出（gitleaks / trufflehog）と SAST（Semgrep）、依存関係の脆弱性チェック（govulncheck / pnpm audit）を CI（`.github/workflows/security.yml`）で強制する
- 機密情報（認証情報・個人情報・決済情報）は `context.Context` に入れない。生存スコープを最小にし、引数で渡す（[`.claude/rules/context-propagation.md`](.claude/rules/context-propagation.md)）
- 暗号化が必要なドメインは `internal/crypto/`（`crypto.ISealer`）経由で行い、平文はローカル変数としてのみ生存させる

## 今後の拡張について

新しいドメイン（EC商材の商品・カート・注文など）を追加する場合も、上記の方針（2層デフォルト・domain＝DBモデル・パッケージ構成）に従う。個別機能の設計は `.claude/skills/design-feature` スキルを使って整理し、必要であればこの DESIGN.md に設計判断を追記する。
