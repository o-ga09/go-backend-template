import { describe, expect, it } from 'vitest'
import { filterProductsByName } from './api'
import type { Product } from './types'

const products: Product[] = [
  { id: 'p1', name: 'コーヒー豆', description: '', priceYen: 1000, stock: 5 },
  { id: 'p2', name: '紅茶', description: '', priceYen: 800, stock: 0 },
  { id: 'p3', name: 'Tea Cup', description: '', priceYen: 1500, stock: 3 },
]

describe('filterProductsByName', () => {
  it('クエリが空の場合は全件を返す', () => {
    expect(filterProductsByName(products, '')).toEqual(products)
  })

  it('部分一致する商品のみを返す', () => {
    expect(filterProductsByName(products, 'コーヒー')).toEqual([products[0]])
  })

  it('大文字・小文字を区別しない', () => {
    expect(filterProductsByName(products, 'tea cup')).toEqual([products[2]])
  })

  it('一致する商品が無い場合は空配列を返す', () => {
    expect(filterProductsByName(products, '存在しない商品')).toEqual([])
  })
})
