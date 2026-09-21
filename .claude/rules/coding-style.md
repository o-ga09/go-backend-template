# コーディングスタイルルール

対象: `backend/`（Go）。他ルールが「何を」「どこに」書くかを定めるのに対し、本ルールは
Goコードの書き方そのもの（コメント量・クエリの組み立て方）を定める。

## 関数・メソッドのコメントは最小限にする 🔴

- 関数コメントに**仕様（何をするか）を書かない**。シグネチャ・実装を読めば分かることを
  コメントで繰り返さない
- コメントを書くのは、実装だけでは分からない **非自明な理由（なぜそう書いたか）** がある
  場合のみ（例: 特定のバグの回避策である、GORMの制約でこの順序が必須である、等）
- パッケージコメント・大きな設計判断の要約は許容するが、各メソッドに毎回docコメントを
  付ける必要はない

```go
// 誤: シグネチャから自明な内容をそのまま説明している
// FindByID はIDで商品を検索する。見つからない場合はerrors.ErrRecordNotFoundを返す。
func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
    ...
}

// 正: コメント無し(自明なため)。非自明な理由がある場合のみ残す
func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
    ...
}

// 正: 非自明な理由(GORMの制約)がある箇所にだけコメントを残す
// order_itemsは1件ずつCreateする(BaseModelPluginがスライスのreflectに未対応のため)。
for i := range o.Items {
    ...
}
```

## `pkg/context.GetDBFromCtx` は呼び出し側で `.WithContext(ctx)` し直さない 🔴

`GetDBFromCtx(ctx)` は内部で既に `db.WithContext(ctx)` を適用して返す
（`pkg/context/context.go`）。呼び出し側で再度 `.WithContext(ctx)` を呼ぶのは冗長なので
書かない。

```go
// 誤: GetDBFromCtxが既にWithContext済みのdbを返すのに、呼び出し側でも呼んでいる
db := Ctx.GetDBFromCtx(ctx)
db.WithContext(ctx).Where("id = ?", id).First(&p)

// 正
db := Ctx.GetDBFromCtx(ctx)
db.Where("id = ?", id).First(&p)
```

## 変数への代入は再利用する場合のみ 🔴

`Ctx.GetDBFromCtx(ctx)` の戻り値をサブクエリ等で再利用しないなら、変数に代入せず
そのままメソッドチェーンする。

```go
// 誤: dbを一度しか使わないのに変数に代入している
db := Ctx.GetDBFromCtx(ctx)
if err := db.Where("id = ?", id).First(&p).Error; err != nil {
    return nil, err
}

// 正: そのままチェーンする
if err := Ctx.GetDBFromCtx(ctx).Where("id = ?", id).First(&p).Error; err != nil {
    return nil, err
}

// 正: サブクエリ等で同じdbを複数回使う場合は変数に代入してよい
db := Ctx.GetDBFromCtx(ctx)
sub := db.Model(&Product{}).Select("id").Where("stock > ?", 0)
db.Where("id IN (?)", sub).Find(&ps)
```

## has-many の関連付けは GORM の `Preload` を使う 🔴

カート明細（`cart_items`）・注文明細（`order_items`）のような1対多の関連は、
`gorm:"-"` で関連付けを切って手動で2回クエリ（親→明細）してメモリ上で詰め替える
のではなく、GORMのアソシエーション（`gorm:"foreignKey:XxxID"`）+ `Preload("Items")`
を使う。

```go
// 誤: 手動で2回クエリしてメモリ上で詰め替える
type Cart struct {
    model.BaseModel
    UserID string     `gorm:"column:user_id"`
    Items  []CartItem `gorm:"-"`
}
// リポジトリ側
var c Cart
db.Where("user_id = ?", userID).First(&c)
var items []CartItem
db.Where("cart_id = ?", c.ID).Find(&items)
c.Items = items

// 正: GORMのアソシエーションとPreloadに任せる
type Cart struct {
    model.BaseModel
    UserID string     `gorm:"column:user_id"`
    Items  []CartItem `gorm:"foreignKey:CartID"`
}
// リポジトリ側
var c Cart
db.Preload("Items").Where("user_id = ?", userID).First(&c)
```

書き込み側（`Create`）で明細を1件ずつ`Create`している既存の実装
（`internal/database/mysql/order.go`）は、`BaseModelPlugin`がスライスのreflectに
未対応であることの回避策であり、これは維持する。GORMの関連付けを有効にすると
`db.Create(&order)`が明細もバッチINSERTしようとして同じ問題に当たるため、
親の`Create`では`.Omit(clause.Associations)`を指定し、明細への書き込みは
既存の1件ずつのループを使う。読み取り（`Preload`）と書き込みのバッチINSERT回避は
別問題として扱うこと。
