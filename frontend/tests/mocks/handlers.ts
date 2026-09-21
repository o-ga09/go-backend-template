import { http, HttpResponse } from 'msw'

export const API_BASE_URL = 'http://localhost:8080'

/**
 * デフォルトのハンドラ。各テストで挙動を検証したい場合は server.use() で上書きする。
 */
export const handlers = [
  http.get(`${API_BASE_URL}/api/products`, () => HttpResponse.json([])),
  http.get(`${API_BASE_URL}/api/products/:id`, ({ params }) =>
    HttpResponse.json({
      id: params.id,
      name: '',
      description: '',
      priceYen: 0,
      stock: 0,
    })
  ),

  http.get(`${API_BASE_URL}/api/cart`, () => HttpResponse.json({ id: '', items: [] })),
  http.post(`${API_BASE_URL}/api/cart`, () => HttpResponse.json({ id: 'cart-1', items: [] })),
  http.put(`${API_BASE_URL}/api/cart`, () => HttpResponse.json({ id: 'cart-1', items: [] })),
  http.delete(`${API_BASE_URL}/api/cart`, () => HttpResponse.json({ id: 'cart-1', items: [] })),

  http.get(`${API_BASE_URL}/api/orders`, () => HttpResponse.json([])),
  http.get(`${API_BASE_URL}/api/orders/:id`, ({ params }) =>
    HttpResponse.json({ id: params.id, status: 'pending', totalPriceYen: 0, items: [] })
  ),
  http.post(`${API_BASE_URL}/api/orders`, () =>
    HttpResponse.json(
      { id: 'order-1', status: 'pending', totalPriceYen: 0, items: [] },
      { status: 201 }
    )
  ),
]
