---
name: create-pr
description: >
  PRを作成するスキル。ブランチの変更内容を分析し、既存のPRテンプレートに沿ったPull Requestを作成する。
  「PRを作って」「プルリクエストを作成して」「PR出して」「変更をPRにして」「マージ依頼を出して」
  と言われたときに必ずこのスキルを使用すること。
  Golangバックエンド・Reactフロントエンドの両方に対応。
  GitHub MCPが使える場合はMCPを優先する。
---

# PR作成スキル

変更内容を分析して、PRテンプレートに沿ったPull Requestを作成するワークフロー。

---

## Step 1: 変更内容の把握

現在のブランチと差分を確認する。

```bash
# ベースブランチとの差分ファイル一覧
git diff main...HEAD --name-only

# コミット履歴（ベースブランチからの変更）
git log main...HEAD --oneline

# 変更の詳細（必要に応じて）
git diff main...HEAD
```

確認ポイント:
- 変更ファイルの種類（`backend/`（Go）/ `frontend/`（React）/ 両方 / `.github/`（CI・インフラ）/ `.claude/`（ルール・スキル）/ 設定）
- 変更の目的（新機能 / バグ修正 / リファクタリング / ドキュメント）
- 破壊的変更の有無（API仕様変更、DB migrationなど）

---

## Step 2: PRテンプレートの取得

`.github/PULL_REQUEST_TEMPLATE.md` を読み込む。テンプレートが複数ある場合は対象の変更に最も近いものを選ぶ。

```bash
# テンプレートの確認
cat .github/PULL_REQUEST_TEMPLATE.md
# または
ls .github/PULL_REQUEST_TEMPLATE/
```

テンプレートが存在しない場合は以下のデフォルト構成で作成する:

```markdown
## 変更内容

## 変更理由・背景

## テスト方法

## スクリーンショット（フロントエンド変更時）

## チェックリスト
- [ ] テストを追加・更新した
- [ ] ドキュメントを更新した
- [ ] 破壊的変更がある場合は記載した
```

---

## Step 3: PRタイトルの決定

Conventional Commits の type をベースにタイトルを構成する。

| type | 使いどころ |
|---|---|
| `feat` | 新機能追加 |
| `fix` | バグ修正 |
| `refactor` | 機能変更を伴わないコード改善 |
| `test` | テスト追加・修正 |
| `docs` | ドキュメントのみの変更 |
| `chore` | ビルド・依存関係・ツールの変更 |
| `perf` | パフォーマンス改善 |
| `ci` | CI/CD設定の変更 |
| `style` | コードスタイルのみの変更（動作変更なし） |

**フォーマット**: `[type] 変更内容を端的に説明する（日本語可）`

例:
- `[feat] ユーザー認証エンドポイントを追加`
- `[fix] 合計金額の計算ロジックを修正`
- `[refactor] Webhookハンドラを責務ごとに分割`

---

## Step 4: PR本文の作成

**PR本文（概要・変更内容・変更理由・背景等の説明文）は日本語で記述する。** コード例・コマンド・識別子（関数名・パッケージ名・エンドポイントパス等）は原文のまま英語でよい。

テンプレートの各セクションを埋める。

### 変更内容
- 何を変更したかを箇条書きで記載
- `backend/` の変更: パッケージ名・関数名・エンドポイントを明記
- `frontend/` の変更: コンポーネント名・画面名を明記

### 変更理由・背景
- なぜこの変更が必要か
- 関連 Issue があれば `Closes #<issue_number>` で紐付ける

### テスト方法
- 実行したテストコマンドと結果を記載

```bash
# backend/ の場合
cd backend
go test ./...
go vet ./...
golangci-lint run --config=./.golangci.yml ./...

# frontend/ の場合
cd frontend
pnpm test
pnpm type-check
pnpm lint
pnpm format
```

### スクリーンショット
`frontend/` に変更がある場合は Before/After のスクリーンショットを添付する。
`ui-screenshot` スキルが利用可能なら使用する。

### チェックリスト
全項目を確認し、該当する場合は `[x]` にする。

---

## Step 5: Assignee とラベルの決定

**Assignee は常に自分（`@me`）にする。** PRを作成する＝作業者本人が担当者なので、明示的な指示がなくても必ず設定する。

ラベルはリポジトリの既存ラベル一覧から、変更のスコープに合うものを選ぶ。存在しないラベルを新規作成しない。

```bash
gh label list --limit 100
```

選定方針:
- Step 3 で判定した Conventional Commits の type をラベル名にマッピングする（例: `feat`→新機能系ラベル、`fix`→不具合系ラベル、`docs`→ドキュメント系ラベルなど、リポジトリのラベル名に合わせる）
- ブランチに複数種別のコミットが混在する場合（例: `feat` と `fix` の両方）は該当するラベルを複数付与する
- 対象領域（`backend`／`frontend`／`.github`／CI等）に対応するラベルがあれば合わせて付与する
- 該当するラベルが無ければ無理に付けない

---

## Step 6: PRの作成

### GitHub MCP が使える場合（推奨）

```
mcp_github_create_pull_request(
  owner="<owner>",
  repo="<repo>",
  title="<PRタイトル>",
  body="<テンプレートに沿ったPR本文>",
  head="<現在のブランチ名>",
  base="main",
  draft=false
)
```

MCPの `create_pull_request` は assignee / label を受け付けないことが多い。作成後に `mcp_github_update_pull_request`（または後述の `gh pr edit`）で Assignee とラベルを設定する。

### `gh` コマンドを使う場合

まずブランチをプッシュしてからPRを作成する。Assignee とラベルは作成時に指定する。

```bash
git push origin <branch-name>

gh pr create \
  --title "<PRタイトル>" \
  --body "$(cat <<'EOF'
<テンプレートに沿ったPR本文>
EOF
)" \
  --base main \
  --assignee "@me" \
  --label "<ラベル1>" --label "<ラベル2>"
```

### 既に開いているPRを更新する場合

対象ブランチのPRが既に存在する場合は新規作成せず、Assignee とラベルのみ追記する。

```bash
gh pr edit <PR番号> --add-assignee "@me" --add-label "<ラベル1>"
```

---

## Step 7: 作成確認

PR URLとAssignee・ラベルを取得して設定完了を確認する。

```bash
gh pr view --json url,title,assignees,labels
```

作成後にユーザーへ報告する内容:
- PR URL
- タイトル
- ベースブランチ / ヘッドブランチ
- Assignee（`@me` が設定されていること）
- 付与したラベル

---

## 補足: ブランチの命名規則

PRを作成する前にブランチ名がプロジェクトの規約に沿っているか確認する。

| 変更の種類 | ブランチ名の例 |
|---|---|
| 新機能 | `feature/<issue-number>-<description>` |
| バグ修正 | `fix/<issue-number>-<description>` |
| リファクタリング | `refactor/<description>` |
| ドキュメント | `docs/<description>` |
| CI/ツール | `chore/<description>` |
| リリース | `release/<version>` |

ブランチ名にスペースは使わず、ハイフン区切りの英小文字を使う。
