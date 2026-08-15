# テストルール

`docs/requirements.md` NFR-10 に対応。**TDD（失敗するテスト → 実装 → リファクタ）で進める。**

## 原則に紐づくテストは必ず書く 🔴

このプロダクトの価値は原則10箇条にあり、**原則はテストでしか守られない。**
原則を実装するロジックは UI・DB から独立した純粋関数として実装し、テストに**根拠を書く**（NFR-10-02）。

```go
// 原則4（「特にない」は掘るサイン）/ docs/research/2026-07-22-interview-01.md:26
// 人力テストで2回とも聞き手が掘り返しを忘れた。機械的に発火させる。
func TestDetectDigDeeper(t *testing.T) { ... }
```

必須のテストケース（要件の受け入れ条件そのもの）:

| 対象 | 根拠 |
|---|---|
| 検知語の**部分一致**。実発話「心当たりはとくにない」「関係ないような気がするけど」「わからない」で発火すること | FR-04-01a。**完全一致実装ではこの3件が素通りする** |
| 掘り下げが同一設問で1回しか発火しないこと | FR-04-03 / 原則4 |
| 基本設問が5ステップを超えないこと（掘り下げはカウントしない） | FR-03-03 / 原則7 |
| 質問文定数が `docs/interview-design.md` と一致すること | FR-03-04 |
| 生成プロンプトが `docs/generation-prompt.md` と一致すること | FR-06-02 |
| メモリに送信フラグ相当のフィールドが存在しないこと | FR-09-90 / 原則8 |
| ログ・エラーに本文がマスキングされること | NFR-03-05 |
| 推しをまたいでメモリが混ざらないこと | FR-02-91 |

## Go（`backend/`）

### ファイル配置

- テストファイルは実装ファイルと同一パッケージ・同一ディレクトリに置く（`foo.go` → `foo_test.go`）
- ブラックボックステストが必要な場合のみ `package xxx_test` を使う

### テストの書き方

- テーブル駆動テスト（`[]struct{ name, input, want }`）を基本形とする
- サブテストは `t.Run` で命名する。テスト名は日本語可
- HTTP handler のテストは `net/http/httptest` + `echo.New()` で実際のルーターを立てる

```go
func TestSessionServer_Answer(t *testing.T) {
    tests := []struct {
        name       string
        input      string
        wantDig    bool
        wantStatus int
    }{
        {"検知語そのまま", "特にない", true, http.StatusOK},
        {"前に語が付く（実発話）", "心当たりはとくにない", true, http.StatusOK},
        {"語尾が違う（実発話）", "関係ないような気がするけど", true, http.StatusOK},
        {"通常回答", "武道館のライブです", false, http.StatusOK},
    }
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            e := echo.New()
            req := httptest.NewRequest(http.MethodPost, "/sessions/1/answers", body(tc.input))
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            // ...
            assert.Equal(t, tc.wantStatus, rec.Code)
        })
    }
}
```

### モック方針

- **DB は実 DB を使う**（`docker compose up -d` で起動した MySQL。TiDB は MySQL 互換なのでローカルは MySQL でよい）。GORM のモック禁止
- **LLM（Anthropic API）は必ずインターフェースモックを使う。実 API を絶対に叩かない**
  - 課金が発生する／出力が非決定的でテストにならない／CI が外部依存で壊れる
- **Cloud KMS もモックする。** 暗号化ラッパーのラウンドトリップはローカルの鍵で検証する
- モックは `internal/domain/<ドメイン>/mock/` に置く（`moq` で生成）
- **テスト内にモック・スタブをローカル実装しない**

```go
// 正: moq 生成のモックを使う
llmMock := &moq.IDraftGeneratorMock{
    GenerateFunc: func(ctx context.Context, m *draft.Materials) (*draft.Draft, error) {
        return testDraft, nil
    },
}
```

### 生成品質は自動テストで判定しない

LLM 出力は非決定的で、**自動テストでは「壊れていないこと」までしか保証できない**（NFR-06）。

- 自動テストで見るのは「プロンプトが正しく組み立てられたか」「保存経路が動くか」まで
- **品質の判定は `docs/research/reference-output.md` との人力比較**。結果は都度 `docs/research/` に記録する
- プロンプトを変えたら比較評価をやり直す（NFR-06-02）

### 禁止事項

- `time.Sleep` をテスト内で使わない
- テストで外部ネットワークにアクセスしない（**特に Anthropic API**）
- `os.Exit` を呼ぶコードはテストできないため、`main` 以外では使わない
- **テストデータに実在の推しの名前・`reference-output.md` の内容を使わない**（個人的な内容を含むため外部公開・デモに使用しない）

### カバレッジ

- `go test ./... -coverprofile=coverage.out` で計測する
- `domain/` のビジネスロジックは 80% 以上を目標とする
- **原則を実装するロジック（FR-03 / FR-04 / FR-06）は 100% を目標とする**

## クライアント（React Native / Expo）

### ロジック（Vitest）

- テストフレームワーク: **Vitest**（`pnpm test` / `pnpm test:watch`）
- 対象は `src/**/*.{test,spec}.ts`。**RN コンポーネントは対象外、純粋ロジックのみ**（`package.json` の設定どおり）
- 既存の `src/interview/interview.ts` + `interview.test.ts` が参考パターン。**原則番号をコメントで紐付ける形を踏襲する**

```ts
// 原則7: 質問は5問まで
// docs/principles.md / docs/interview-design.md
describe('shouldEndInterview', () => {
  it('5問目の回答後に終了する', () => { ... })
})
```

### 画面（jest-expo VRT）

- テストフレームワーク: **jest-expo** + `@testing-library/react-native`（`pnpm test:vrt` / `pnpm test:vrt:watch`）
- 対象は `src/**/*.test.tsx` / `App.test.tsx`（`jest.config.js`）。vitest とは拡張子（`.ts` / `.tsx`）で棲み分ける
- スナップショット（`toMatchSnapshot()`）で画面の見た目の意図しない変化を検知する。**ロジックの正しさの検証ではない**（それは上記 Vitest の役割）
- スナップショット差分が出たら、意図した変更かをレビューで確認してから `pnpm test:vrt:update` で更新する。**差分を見ずに機械的に更新しない**
- 実機での動作確認（タップ操作・アニメーション・実際のレイアウト崩れ）は引き続き必須。VRT は実機確認を代替しない

- コミット前に `pnpm typecheck` / `pnpm lint` / `pnpm test` / `pnpm test:vrt` が通ること
