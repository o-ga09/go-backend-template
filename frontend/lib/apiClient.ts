const baseURL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE'
  body?: unknown
  query?: Record<string, string | undefined>
}

function buildURL(path: string, query?: Record<string, string | undefined>): string {
  const url = new URL(path, baseURL)
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value !== undefined) {
        url.searchParams.set(key, value)
      }
    }
  }
  return url.toString()
}

/**
 * バックエンドAPIを呼び出す共通クライアント。credentials: 'include'でセッション
 * Cookieを送信する。エラー時はErrorHandler(backend)が返す{error: string}を
 * ApiErrorに変換する。
 */
export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(buildURL(path, options.query), {
    method: options.method ?? 'GET',
    credentials: 'include',
    headers: options.body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  })

  if (!response.ok) {
    const message = await extractErrorMessage(response)
    throw new ApiError(response.status, message)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}

async function extractErrorMessage(response: Response): Promise<string> {
  try {
    const data = (await response.json()) as { error?: string }
    if (data.error) {
      return data.error
    }
  } catch {
    // JSONでないレスポンスはデフォルトメッセージにフォールバックする
  }
  return `リクエストに失敗しました (${response.status})`
}
