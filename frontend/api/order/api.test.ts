import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { API_BASE_URL } from '@/mocks/handlers'
import { server } from '@/mocks/server'
import { createOrder, fetchOrder, fetchOrders } from './api'
import type { Order } from './types'

const order: Order = {
  id: 'o1',
  status: 'pending',
  totalPriceYen: 1000,
  items: [{ productId: 'p1', quantity: 1, unitPriceYen: 1000 }],
}

describe('fetchOrders', () => {
  it('GET /api/ordersのレスポンスをそのまま返す', async () => {
    server.use(http.get(`${API_BASE_URL}/api/orders`, () => HttpResponse.json([order])))

    await expect(fetchOrders()).resolves.toEqual([order])
  })
})

describe('fetchOrder', () => {
  it('GET /api/orders/:idで指定した注文を取得する', async () => {
    server.use(http.get(`${API_BASE_URL}/api/orders/o1`, () => HttpResponse.json(order)))

    await expect(fetchOrder('o1')).resolves.toEqual(order)
  })
})

describe('createOrder', () => {
  it('POST /api/ordersでボディ無しのリクエストを送信する', async () => {
    let capturedBody: string
    server.use(
      http.post(`${API_BASE_URL}/api/orders`, async ({ request }) => {
        capturedBody = await request.text()
        return HttpResponse.json(order, { status: 201 })
      })
    )

    const result = await createOrder()

    expect(capturedBody!).toBe('')
    expect(result).toEqual(order)
  })
})
