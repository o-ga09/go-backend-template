---
name: code-review
description: >
  コードレビューを行うスキル。変更内容を🛑（必須修正）🟡（推奨修正）🟢（提案）の3レベルで評価し、
  問題がなければLGTMバッジを付与してGitHubでApproveする。
  「レビューして」「このPRをレビューして」「コードレビューをお願い」「変更を確認して」
  「PRをチェックして」「差分を見て」と言われたときに必ずこのスキルを使用すること。
  GoバックエンドとReactフロントエンドの両方のレビューに対応。
  GitHubのPR番号が与えられた場合はそのPRを、なければ現在のブランチの差分をレビューする。
---

# コードレビュースキル

変更内容を3段階のレベルで評価し、結果をレポート出力する。問題がなければLGTMとしてApproveする。

---

## レビューレベル定義

| レベル | 意味 | 対応方針 |
|---|---|---|
| 🛑 MUST | 必ず修正が必要。マージ不可。 | バグ・セキュリティ脆弱性・仕様違反・ビルド/テスト破壊・データ破損リスク・`.claude/rules/` 違反 |
| 🟡 SHOULD | 修正を強く推奨。マージは可能だが対応を期待。 | パフォーマンス問題・保守性低下・エラーハンドリング漏れ・ベストプラクティス違反 |
| 🟢 NIT | 提案・好みの問題。対応は任意。 | 命名改善・コメント追加・スタイル統一・リファクタリング提案 |

🛑 が1件でもある場合は LGTM/Approve しない。
🟡 のみの場合は、内容の重大度をユーザーに伝えた上で Approve するか確認する。

---

## Step 1: レビュー対象の取得

### PR番号が指定された場合

**GitHub MCP が使える場合（推奨）:**
```
mcp_github_get_pull_request(owner="<owner>", repo="<repo>", pull_number=<番号>)
mcp_github_list_pull_request_files(owner="<owner>", repo="<repo>", pull_number=<番号>)
```

**`gh` コマンドを使う場合:**
```bash
gh pr view <番号> --json title,body,files,commits
gh pr diff <番号>
```

### 現在のブランチの変更を対象にする場合

```bash
git diff main...HEAD --name-only   # 変更ファイル一覧
git diff main...HEAD               # 変更の詳細
git log main...HEAD --oneline      # コミット履歴
```

---

## Step 2: レビュー観点

### 全言語共通

- [ ] **セキュリティ**: SQLインジェクション・XSS・機密情報のハードコード・認証/認可漏れ
- [ ] **エラーハンドリング**: エラーが握りつぶされていないか・適切なエラーメッセージか
- [ ] **テスト**: 変更に対応したテストが存在するか・テストが意味のある検証をしているか
- [ ] **仕様準拠**: PRの説明・Issueの要件を満たしているか
- [ ] **破壊的変更**: API仕様の変更・DBスキーマ変更が適切にドキュメント化されているか

### `backend/`（Go）固有 — `.claude/rules/architecture.md` `error-handling.md` `context-propagation.md` `transaction.md` `testing.md` `package-structure.md` に準拠しているか

- [ ] **レイヤー構成**: シンプルな CRUD に不要な `internal/service/`（usecase層）を作っていないか。ドメインロジック（バリデーション・状態遷移判定）が domain 層の純粋関数になっているか
- [ ] **domain＝DBモデル**: 不要な変換専用構造体を作っていないか。`BaseModel` の各カラムをリポジトリで手動代入していないか（GORMプラグインに委ねる）
- [ ] **エラー処理**: `pkg/errors`（`ergo`ベース）経由になっているか。`fmt.Errorf` / 標準 `errors.New` を直接使っていないか。エラーメッセージに機密情報（パスワード・トークン・個人情報・決済情報）が含まれていないか
- [ ] **ログの重複**: `Make*Error`/`Wrap` 内で既にログ済みの失敗を呼び出し側で再度 `logger.Warn`/`Error` していないか
- [ ] **Context伝搬**: 第一引数が `ctx context.Context` になっているか。`context.Background()` を `main` 以外で作っていないか。機密情報・ビジネスパラメータを context に入れていないか
- [ ] **トランザクション**: 複数テーブル書き込みが `ITransactionManager.RunInTx` でラップされているか。外部API呼び出し・暗号化処理をトランザクション内に置いていないか。`gorm.DB.Transaction`/`Begin` を直接呼んでいないか
- [ ] **楽観ロック**: `version` チェックをアプリケーションコードで手書きしていないか（GORMプラグインに委ねる）
- [ ] **nil安全性 / goroutineリーク / リソース解放**: nilポインタデリファレンス、`context` キャンセル伝播漏れ、`defer f.Close()` 漏れ
- [ ] **依存注入**: コンストラクタで受け取った依存にメソッド内でnilガードを書いていないか
- [ ] **命名規則**: `internal/domain/<name>/` のパッケージ名がディレクトリ名と一致しているか。`utils`/`helpers`/`common` のような曖昧なパッケージを作っていないか
- [ ] **マイグレーション**: `AutoMigrate` を使わず `db/migrations/`（`sql-migrate`）で管理されているか
- [ ] **`go vet` / `golangci-lint`**: 静的解析で検出できる問題

### `frontend/`（Next.js）固有 — `.claude/rules/frontend.md` に準拠しているか

- [ ] **データフェッチ**: コンポーネントから直接 `fetch` していないか。TanStack Query 経由（`api/<ドメイン>/` のフック）になっているか
- [ ] **型安全性**: TypeScript の型定義が適切か・`any` の不用意な使用がないか
- [ ] **認証**: NextAuthのクライアントシークレット（`GOOGLE_CLIENT_SECRET`, `NEXTAUTH_SECRET`）をクライアントコンポーネントに露出させていないか
- [ ] **バリデーション**: フォームが Zod + React Hook Form になっているか
- [ ] **再レンダリング / useEffect依存配列**: 不必要な再レンダリング、依存配列の漏れによる無限ループのリスク
- [ ] **アクセシビリティ**: `alt` 属性・`aria-*` 属性・キーボード操作対応
- [ ] **XSS対策**: `dangerouslySetInnerHTML` の不用意な使用
- [ ] **コンポーネント設計**: 1ファイル1コンポーネント・shadcn/ui の活用
- [ ] **状態管理ライブラリの追加**: Redux等を新規導入していないか（TanStack Query + Context で足りる）

---

## Step 3: レビュー結果の出力

以下の形式でレビュー結果を出力する。

```markdown
## コードレビュー結果

**対象**: <PRタイトルまたはブランチ名>
**レビュー日時**: <日時>

---

### サマリー

| レベル | 件数 |
|---|---|
| 🛑 MUST | N件 |
| 🟡 SHOULD | N件 |
| 🟢 NIT | N件 |

---

### 指摘事項

#### 🛑 MUST（必須修正）

**[ファイル名:行番号]** 指摘内容

> ```go
> // 問題のあるコード
> ```

修正案: <具体的な修正方法>

---

#### 🟡 SHOULD（推奨修正）

**[ファイル名:行番号]** 指摘内容

修正案: <具体的な修正方法>

---

#### 🟢 NIT（提案）

**[ファイル名:行番号]** 提案内容

---

### 総評

<変更全体についての所感。良い点も積極的に記載する。>
```

---

## Step 4: 問題がない場合のLGTM

🛑 が0件の場合、以下のLGTMメッセージを出力する。

```markdown
---

![LGTM](https://img.shields.io/badge/review-LGTM%20%E2%9C%85-brightgreen)

コードレビューの結果、問題は見当たりません。マージ可能です。
```

🛑 が0件かつ🟡 が0件の場合は即座にApproveを実行する。
🟡 のみある場合はユーザーに確認を取ってからApproveを検討する。

---

## Step 5: GitHubでApprove

### GitHub MCP が使える場合（推奨）

```
mcp_github_create_pull_request_review(
  owner="<owner>",
  repo="<repo>",
  pull_number=<番号>,
  event="APPROVE",
  body="LGTM ✅\n\n変更内容を確認しました。問題ありません。"
)
```

### `gh` コマンドを使う場合

```bash
gh pr review <番号> --approve --body "LGTM ✅

変更内容を確認しました。問題ありません。"
```

---

## レビュー観点の補足

### 良い点も明記する

問題点だけでなく、優れた実装や工夫されている点も積極的に記載する。
これによりレビュイーがどの方向性が良かったかを学べる。

### 指摘は具体的に

「命名が良くない」ではなく「`data` より `userList` の方が意図が明確になります」のように具体的に書く。
修正案が明確でないと対応コストが上がるため、可能な限りコードスニペットを添える。

### コンテキストを考慮する

そのコードが存在する背景（技術的制約・期限・実験的な変更）を考慮する。
完璧主義よりも「このPRの目的を達成しているか」を優先する。
