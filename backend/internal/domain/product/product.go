package product

// HasStock は在庫が指定数量を充足するかどうかを判定する
// (ドメイン間に新しい矢印を作らない方針。architecture.md参照)。
// Task 4(order)は注文確定時に、この判定結果を注文ドメインの純粋関数に渡す想定。
func (p *Product) HasStock(quantity int) bool {
	return p.Stock >= quantity
}
