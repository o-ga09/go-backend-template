# DESIGN.md

`frontend/`（Next.js）の UI デザインをまとめたドキュメント。デザインシステム（トークン・コンポーネント方針）と、EC商材サンプルの画面仕様の2部構成。バックエンドのアーキテクチャ設計判断は [CLAUDE.md](CLAUDE.md)「アーキテクチャの設計判断」を参照。

## デザインシステム

`frontend/tailwind.config.ts` / `frontend/app/globals.css` に定義済みのトークンを正とする。ここでは一覧化のみ行い、値の変更は実装側（Tailwind設定）で行う。

### カラー

shadcn/ui のデフォルトパレット（HSLカスタムプロパティ、`tailwind.config.ts` の `theme.extend.colors` から `hsl(var(--xxx))` として参照）をベースにする。ライト/ダーク両方を `app/globals.css` の `:root` / `.dark` で定義済み（`darkMode: 'class'`）。

| トークン | 用途 |
|---|---|
| `background` / `foreground` | ページ全体の背景・文字色 |
| `primary` / `primary-foreground` | 主要アクション（購入する・注文するボタンなど） |
| `secondary` / `secondary-foreground` | 補助的なアクション |
| `muted` / `muted-foreground` | 補足情報・非活性テキスト |
| `accent` / `accent-foreground` | ホバー・選択状態のハイライト |
| `destructive` / `destructive-foreground` | 削除・キャンセルなど破壊的操作 |
| `card` / `card-foreground` | カード型コンテナ（商品カードなど） |
| `border` / `input` / `ring` | 枠線・フォーム入力・フォーカスリング |

新しい色を追加する場合は個別コンポーネントに直書きせず、`globals.css` にトークンを追加してから Tailwind の `colors` に登録する。

### タイポグラフィ

フォントは Geist Sans（本文）/ Geist Mono（コード・数値表示）を `next/font/google` で読み込み（`lib/font.ts`）、CSS変数（`--font-geist-sans` / `--font-geist-mono`）として `layout.tsx` の `<body>` に適用済み。

`tailwind.config.ts` の `fontSize` にタイポグラフィスケールを定義済み。

| クラス | サイズ / 行間 | 用途 |
|---|---|---|
| `text-display-1` / `text-display-2` | 72px / 60px, line-height 1.1 | ランディング等の大見出し |
| `text-heading-1`〜`text-heading-4` | 48px〜24px, line-height 1.2 | 画面タイトル・セクション見出し |
| `text-body-lg` / `text-body-base` / `text-body-sm` / `text-body-xs` | 18px〜12px, line-height 1.5 | 本文・補足・キャプション |

新しいテキストスタイルが必要な場合は `tailwind.config.ts` の `fontSize` にスケールとして追加し、コンポーネント側でアドホックな `text-[Npx]` を書かない。

### スペーシング・角丸

- スペーシング: `tailwind.config.ts` の `spacing`（4px刻み。`1`=4px 〜 `24`=96px）を使う。アドホックな `p-[Npx]` を避ける
- 角丸: `--radius`（`globals.css`、デフォルト `0.5rem`）から `rounded-lg` / `rounded-md` / `rounded-sm` を導出。個別に `rounded-[Npx]` を指定しない

### コンポーネント方針

- UI コンポーネントは shadcn/ui（Radix UI + Tailwind CSS）を使う。新規追加は `npx shadcn@latest add <component>` で `components/ui/` に生成する（`.claude/rules/frontend.md`）
- ダイアログ系コンポーネントで z-index の衝突が起きた実績があるため（`globals.css` の `.dialog-overlay` / `.dialog-content` 上書き）、新しいオーバーレイ系コンポーネント（Dialog/Drawer/Popover等）を追加する際は既存の z-index 上書きと衝突しないか確認する
- ダークモードは `class` 戦略。新規コンポーネントはライト/ダーク両方のトークンで見た目を確認する

---

## 画面仕様（EC商材サンプル）

対応する Issue: 「Frontend: EC画面の実装(商品一覧・詳細・カート・注文・マイページ)」。**現時点でこれらの画面は未実装**であり、以下は実装時に従う設計仕様。実装後は本セクションを実態に合わせて更新する。

### 画面一覧

| パス | 画面名 | 概要 | 認証 |
|---|---|---|---|
| `/` | 商品一覧 | 商品をカード（`card` トークン）のグリッドで表示。検索・カテゴリ絞り込み | 不要 |
| `/products/[id]` | 商品詳細 | 商品画像・説明・価格・在庫状況、カート追加ボタン（`primary`） | 不要 |
| `/cart` | カート | カート内商品の一覧・数量変更・削除、合計金額、レジに進むボタン | 必要 |
| `/checkout` | 注文（購入手続き） | 配送先・支払い方法の確認、注文確定ボタン（`primary`、二重送信防止） | 必要 |
| `/mypage` | マイページ | ユーザー情報・注文履歴へのリンク | 必要 |
| `/mypage/orders` | 注文履歴 | 過去の注文一覧・ステータス表示 | 必要 |

「認証が必要」な画面は、未ログイン時は NextAuth のログイン導線へリダイレクトする（`context/authContext.tsx` の認証状態を参照。`.claude/rules/frontend.md`）。

### 共通レイアウト

- ヘッダー: ロゴ／サービス名、カートアイコン（点数バッジ表示）、ログイン状態に応じたユーザーメニュー
- 状態表示: 全画面でローディング（TanStack Query の `isLoading`）・エラー・空状態（例: カートが空）をコンポーネントレベルで明示的にハンドリングする（`.claude/rules/frontend.md`「API呼び出し」）
- トースト通知: `components/ui/sonner.tsx`（導入済み）をカート追加・注文確定などの完了通知に使う

### 商品一覧画面

- 商品カード: 画像・商品名（`heading-4`）・価格（`body-lg`）・カート追加ボタン
- 検索・絞り込みはURLクエリパラメータで状態を持つ（画面リロードでも条件を保持）
- データ取得は `api/product/` のフック経由（TanStack Query）。コンポーネントから直接 `fetch` しない

### 商品詳細画面

- 在庫切れ時はカート追加ボタンを非活性化し、理由を明示する
- 数量選択は在庫数を上限にする

### カート画面

- 数量変更・削除は楽観的更新（optimistic update）を検討するが、失敗時は必ず表示状態を元に戻す
- 空カート時は商品一覧への導線を表示する

### 注文（チェックアウト）画面

- 注文確定ボタンは二重送信を防止する（連打・多重タブ対策）
- 送信中はボタンを非活性化しローディング状態を表示する
- 失敗時は入力内容（配送先等）を保持したまま再試行できるようにする（`.claude/rules/error-handling.md`「ユーザー体験としてのエラー」）

### マイページ・注文履歴画面

- 注文ステータス（例: 処理中・発送済み・完了・キャンセル）はバッジで視覚的に区別する
- 個人情報（氏名・住所等）の表示は必要最小限にする

---

## 更新方針

新しい画面・コンポーネントを追加する際は、まず `design-feature` スキルで機能設計を行い、UIに関わる部分をこのファイルに追記してから `implement-component` スキルで実装する。
