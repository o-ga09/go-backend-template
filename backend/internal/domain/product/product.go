package product

func (p *Product) HasStock(quantity int) bool {
	return p.Stock >= quantity
}
