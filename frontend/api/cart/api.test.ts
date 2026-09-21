import { describe, expect, it } from 'vitest'
import { cartItemCount } from './api'
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
