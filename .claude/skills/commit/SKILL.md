---
name: commit
description: >
  変更内容を分析してConventional Commitsの規約に従ったコミットメッセージを自動生成し、コミットするスキル。
  「コミットして」「変更をコミットして」「commitして」「コミットメッセージを作って」「変更を保存して」
  「ステージングしてコミット」「git commitして」と言われたときに必ずこのスキルを使用すること。
  feat/fix/chore/refactor/test/docs/perf/style/ci のプレフィックスを自動判定する。
  Golangバックエンド・Reactフロントエンドの両方に対応。
---

# コミットスキル

変更内容を分析してConventional Commitsの規約に従ったコミットを行うワークフロー。

## 絶対ルール（違反したコミットは作らない）

1. **Conventional Commits形式は必須。** subject行は必ず `<type>(<scope>): <説明>` または `<type>: <説明>` の形式にする。type は Step 2 の表にある9種（feat/fix/refactor/test/docs/chore/perf/style/ci）のみ。それ以外のtypeや形式外のメッセージ（例: `更新`, `WIP`, `Update README`）でのコミットは禁止。
2. **Co-Authored-By トレーラーは必須。** 現在のセッションの attribution 指示（システムプロンプト／system-reminder で指定されている `Co-Authored-By:` 行）をそのまま最終行に入れる。指示が無い場合はこのトレーラーを追加しない（存在しないモデル名を捏造しない）。
3. ユーザーから形式外のメッセージを直接指定された場合は、そのままコミットせず、規約に沿った形に直した案を提示して確認する。

コミット実行前に、作成したメッセージが上記1・2を満たしているかセルフチェックすること。満たしていなければコミットせずメッセージを直す。

---

## Step 1: 変更内容の確認

```bash
# 現在の状態を確認
git status

# ステージ済みとステージ前の変更を確認
git diff          # 未ステージの変更
git diff --cached # ステージ済みの変更
```

未ステージのファイルがあっても確認不要。そのまま全変更をステージして進む。

ただし、以下のように**明らかに不自然な差分の組み合わせ**がある場合はコミット前にユーザーへ確認する:
- Goバックエンドの変更に `package.json` / `package-lock.json` / `pnpm-lock.yaml` が混入
- Reactフロントエンドの変更に `.go` ファイルや DB migration が混入
- `.env` や秘密鍵など機密情報が含まれている

それ以外は確認なしで進める。

---

## Step 2: type の判定

変更内容を見て最も適切な type を選ぶ。複数の変更が混在する場合は主たる変更の type を使い、詳細は本文に記載する。

| type | 使いどころ | 例 |
|---|---|---|
| `feat` | 新機能・新しいAPIエンドポイント・新コンポーネント | ログイン機能を追加 |
| `fix` | バグ修正・誤ったロジックの修正 | nil参照パニックを修正 |
| `refactor` | 動作を変えずにコードを改善 | Serviceレイヤの責務を分割 |
| `test` | テスト追加・修正（プロダクションコードの変更なし） | ユーザー取得のテストを追加 |
| `docs` | READMEやコメントのみの変更 | APIドキュメントを更新 |
| `chore` | 依存関係更新・ビルド設定・自動生成ファイル | go.mod の依存を更新 |
| `perf` | パフォーマンス改善（N+1解消・クエリ最適化など） | N+1クエリを解消 |
| `style` | フォーマット・インデント（動作変更なし） | gofmt/Prettier適用 |
| `ci` | GitHub Actions・Dockerfileなどのみ変更 | CIのGoバージョンを更新 |

---

## Step 3: scope の判定

変更のスコープを括弧内に記載する。省略も可能だが、変更範囲が明確なときは記載する。

**Golangバックエンド（`backend/`）の例:**
- `(handler)` — `internal/handler/` のハンドラ層
- `(service)` — `internal/service/` のusecase層
- `(domain)` — `internal/domain/` のドメインモデル・ロジック
- `(database)` — `internal/database/mysql/` のリポジトリ層
- `(migration)` — `db/migrations/` のDBマイグレーション

**Reactフロントエンド（`frontend/`）の例:**
- `(ui)` または `(frontend)` — UIコンポーネント全般
- `(components)` — 共通コンポーネント（`components/ui/` など）
- `(app)` — `app/` 配下のページ・ルート
- `(hooks)` — `hooks/` のカスタムフック
- `(api)` — `api/<ドメイン>/` のAPIクライアント・型定義

---

## Step 4: コミットメッセージの作成

**フォーマット:**
```
<type>(<scope>): <subject>

[body（任意）]

[footer（任意）]
```

**subject のルール:**
- 50文字以内
- 動詞の原形で始める（英語の場合: add, fix, update, remove など）
- 日本語でも可（プロジェクトの慣習に合わせる）
- 末尾にピリオドをつけない

**body のルール（任意）:**
- なぜこの変更が必要か、何を変えたかの補足
- 72文字で折り返す

**footer:**
- Co-Author（現在のセッションの attribution 指示がある場合は**必須**）: 最終行に、システムプロンプトで指定された `Co-Authored-By:` トレーラーをそのまま入れる。他のfooterがある場合はその後ろに置く
- Issue参照（任意）: `Closes #123`
- 破壊的変更（任意）: `BREAKING CHANGE: <説明>`

**コミットメッセージの例（Co-Authored-By は現在のセッションの指示に置き換える）:**

```
feat(handler): add user authentication endpoint

JWTを用いたログイン・ログアウトのエンドポイントを追加。
ミドルウェアで認証チェックを行い、未認証時は401を返す。

Closes #42
```

```
fix(database): prevent nil pointer panic on empty result

DBが0件を返す場合にnil参照パニックが発生していたため、
空スライスを返すように修正。
```

```
feat(app): add user profile page

ユーザーのプロフィール情報を表示するページコンポーネントを追加。
アバター画像・名前・自己紹介文を表示する。
```

---

## Step 5: ステージングとコミットの実行

```bash
# 特定ファイルをステージ（推奨）
git add <file1> <file2> ...

# または全変更をステージ（機密情報がないことを確認してから）
git add -A

# コミット（Co-Authored-By トレーラーは現在のセッションの指示に従い最終行に含める）
git commit -m "$(cat <<'EOF'
<type>(<scope>): <subject>

<body（任意）>

<footer（任意）>
EOF
)"
```

コミット直前に以下を確認する（満たさない場合はコミットしない）:

- [ ] subject行が `<type>(<scope>): <説明>` または `<type>: <説明>` 形式（正規表現: `^(feat|fix|refactor|test|docs|chore|perf|style|ci)(\([a-z0-9-]+\))?: .+`）
- [ ] 現在のセッションで attribution 指示がある場合、最終行がその指示どおりの `Co-Authored-By:` トレーラーになっている

コミット後に `git log -1 --format='%s%n%n%b'` で subject の形式と Co-Authored-By トレーラーが記録されているか確認する。漏れていた場合は `git commit --amend` で修正する。

---

## 複数の変更が混在する場合

意味的に異なる変更（例: バグ修正 + 新機能）が混在している場合は、ユーザーに分割コミットを提案する。

```bash
# 特定ファイルのみステージして個別にコミット
git add internal/handler/user.go
git commit -m "fix(handler): handle empty user list correctly"

git add internal/service/payment.go
git commit -m "feat(service): add payment calculation logic"
```

分割するかどうかの最終判断はユーザーに委ねる。
