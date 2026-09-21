import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { API_BASE_URL } from '@/tests/mocks/handlers'
import { server } from '@/tests/mocks/server'
import { apiFetch, ApiError } from './apiClient'

describe('apiFetch', () => {
  it('レスポンスをJSONとしてパースして返す', async () => {
    server.use(http.get(`${API_BASE_URL}/api/products/p1`, () => HttpResponse.json({ id: 'p1' })))

    const result = await apiFetch<{ id: string }>('/api/products/p1')

    expect(result).toEqual({ id: 'p1' })
  })

  it('credentials: includeでリクエストする', async () => {
    let capturedCredentials: RequestCredentials | undefined
    server.use(
      http.get(`${API_BASE_URL}/api/products/p1`, ({ request }) => {
        capturedCredentials = request.credentials
        return HttpResponse.json({ id: 'p1' })
      })
    )

    await apiFetch('/api/products/p1')

    expect(capturedCredentials).toBe('include')
  })

  it('クエリパラメータをURLに付与する', async () => {
    let capturedURL: string | undefined
    server.use(
      http.delete(`${API_BASE_URL}/api/cart`, ({ request }) => {
        capturedURL = request.url
        return HttpResponse.json({})
      })
    )

    await apiFetch('/api/cart', { method: 'DELETE', query: { productId: 'p1' } })

    expect(capturedURL).toContain('productId=p1')
  })

  it('レスポンスが失敗した場合はサーバーのerrorメッセージでApiErrorを投げる', async () => {
    server.use(
      http.get(`${API_BASE_URL}/api/cart`, () =>
        HttpResponse.json({ error: 'insufficient stock' }, { status: 422 })
      )
    )

    await expect(apiFetch('/api/cart')).rejects.toMatchObject(
      new ApiError(422, 'insufficient stock')
    )
  })

  it('JSONでないエラーレスポンスの場合はデフォルトメッセージにフォールバックする', async () => {
    server.use(
      http.get(
        `${API_BASE_URL}/api/cart`,
        () => new HttpResponse('internal error', { status: 500 })
      )
    )

    await expect(apiFetch('/api/cart')).rejects.toMatchObject(
      new ApiError(500, 'リクエストに失敗しました (500)')
    )
  })

  it('204レスポンスはundefinedを返す', async () => {
    server.use(http.get(`${API_BASE_URL}/api/cart`, () => new HttpResponse(null, { status: 204 })))

    await expect(apiFetch('/api/cart')).resolves.toBeUndefined()
  })
})
