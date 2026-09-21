import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { API_BASE_URL } from '@/mocks/handlers'
import { server } from '@/mocks/server'
import { addCartItem, cartItemCount, fetchCart, removeCartItem, updateCartItem } from './api'
import type { Cart } from './types'

describe('cartItemCount', () => {
  it('カートが未定義の場合は0を返す', () => {
    expect(cartItemCount(undefined)).toBe(0)
  })

  it('カートが空の場合は0を返す', () => {
    const cart: Cart = { id: 'c1', items: [] }
    expect(cartItemCount(cart)).toBe(0)
  })

  it('明細の数量を合算して返す', () => {
    const cart: Cart = {
      id: 'c1',
      items: [
        { productId: 'p1', quantity: 2 },
        { productId: 'p2', quantity: 3 },
      ],
    }
    expect(cartItemCount(cart)).toBe(5)
  })
})

describe('fetchCart', () => {
  it('GET /api/cartのレスポンスをそのまま返す', async () => {
    const cart: Cart = { id: 'c1', items: [{ productId: 'p1', quantity: 1 }] }
    server.use(http.get(`${API_BASE_URL}/api/cart`, () => HttpResponse.json(cart)))

    await expect(fetchCart()).resolves.toEqual(cart)
  })
})

describe('addCartItem', () => {
  it('POST /api/cartにproductId・quantityを送信する', async () => {
    let capturedBody: unknown
    server.use(
      http.post(`${API_BASE_URL}/api/cart`, async ({ request }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ id: 'c1', items: [{ productId: 'p1', quantity: 2 }] })
      })
    )

    const result = await addCartItem({ productId: 'p1', quantity: 2 })

    expect(capturedBody).toEqual({ productId: 'p1', quantity: 2 })
    expect(result).toEqual({ id: 'c1', items: [{ productId: 'p1', quantity: 2 }] })
  })
})

describe('updateCartItem', () => {
  it('PUT /api/cartにproductId・quantityを送信する', async () => {
    let capturedBody: unknown
    server.use(
      http.put(`${API_BASE_URL}/api/cart`, async ({ request }) => {
        capturedBody = await request.json()
        return HttpResponse.json({ id: 'c1', items: [{ productId: 'p1', quantity: 3 }] })
      })
    )

    await updateCartItem({ productId: 'p1', quantity: 3 })

    expect(capturedBody).toEqual({ productId: 'p1', quantity: 3 })
  })
})

describe('removeCartItem', () => {
  it('DELETE /api/cartにproductIdをクエリパラメータとして送信する', async () => {
    let capturedURL: string | undefined
    server.use(
      http.delete(`${API_BASE_URL}/api/cart`, ({ request }) => {
        capturedURL = request.url
        return HttpResponse.json({ id: 'c1', items: [] })
      })
    )

    await removeCartItem('p1')

    expect(capturedURL).toContain('productId=p1')
  })
})
