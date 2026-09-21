import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { API_BASE_URL } from '@/mocks/handlers'
import { server } from '@/mocks/server'
import { fetchProduct, fetchProducts, filterProductsByName } from './api'
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

describe('fetchProducts', () => {
  it('GET /api/productsのレスポンスをそのまま返す', async () => {
    server.use(http.get(`${API_BASE_URL}/api/products`, () => HttpResponse.json(products)))

    await expect(fetchProducts()).resolves.toEqual(products)
  })
})

describe('fetchProduct', () => {
  it('GET /api/products/:idで指定した商品を取得する', async () => {
    server.use(http.get(`${API_BASE_URL}/api/products/p1`, () => HttpResponse.json(products[0])))

    await expect(fetchProduct('p1')).resolves.toEqual(products[0])
  })
})
