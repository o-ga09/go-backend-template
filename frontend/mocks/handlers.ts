import { http, HttpResponse } from 'msw'
import { mockCart, mockOrders, mockProducts } from './fixtures'

export const API_BASE_URL = 'http://localhost:8080'

/**
 * デフォルトのハンドラ。各テストで挙動を検証したい場合は server.use() で上書きする。
 * ブラウザでのモック起動(NEXT_PUBLIC_API_MOCKING=enabled)時は、これらの値が
 * そのまま画面に表示される。
 */
export const handlers = [
  http.get(`${API_BASE_URL}/api/products`, () => HttpResponse.json(mockProducts)),
  http.get(`${API_BASE_URL}/api/products/:id`, ({ params }) => {
    const product = mockProducts.find(p => p.id === params.id)
    if (!product) {
      return HttpResponse.json({ error: 'product not found' }, { status: 404 })
    }
    return HttpResponse.json(product)
  }),

  http.get(`${API_BASE_URL}/api/cart`, () => HttpResponse.json(mockCart)),
  http.post(`${API_BASE_URL}/api/cart`, () => HttpResponse.json(mockCart)),
  http.put(`${API_BASE_URL}/api/cart`, () => HttpResponse.json(mockCart)),
  http.delete(`${API_BASE_URL}/api/cart`, () => HttpResponse.json(mockCart)),

  http.get(`${API_BASE_URL}/api/orders`, () => HttpResponse.json(mockOrders)),
  http.get(`${API_BASE_URL}/api/orders/:id`, ({ params }) => {
    const order = mockOrders.find(o => o.id === params.id)
    if (!order) {
      return HttpResponse.json({ error: 'order not found' }, { status: 404 })
    }
    return HttpResponse.json(order)
  }),
  http.post(`${API_BASE_URL}/api/orders`, () => HttpResponse.json(mockOrders[0], { status: 201 })),
]
