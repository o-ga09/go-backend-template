import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiFetch, ApiError } from './apiClient'

describe('apiFetch', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('レスポンスをJSONとしてパースして返す', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify({ id: 'p1' }), { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await apiFetch<{ id: string }>('/api/products/p1')

    expect(result).toEqual({ id: 'p1' })
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/products/p1'),
      expect.objectContaining({ credentials: 'include', method: 'GET' })
    )
  })

  it('クエリパラメータをURLに付与する', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({}), { status: 200 }))
    vi.stubGlobal('fetch', fetchMock)

    await apiFetch('/api/cart', { method: 'DELETE', query: { productId: 'p1' } })

    const calledURL = fetchMock.mock.calls[0][0] as string
    expect(calledURL).toContain('productId=p1')
  })

  it('レスポンスが失敗した場合はサーバーのerrorメッセージでApiErrorを投げる', async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValue(
        new Response(JSON.stringify({ error: 'insufficient stock' }), { status: 422 })
      )
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiFetch('/api/cart')).rejects.toMatchObject(
      new ApiError(422, 'insufficient stock')
    )
  })

  it('JSONでないエラーレスポンスの場合はデフォルトメッセージにフォールバックする', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response('internal error', { status: 500 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiFetch('/api/cart')).rejects.toMatchObject(
      new ApiError(500, 'リクエストに失敗しました (500)')
    )
  })

  it('204レスポンスはundefinedを返す', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(apiFetch('/api/cart')).resolves.toBeUndefined()
  })
})
