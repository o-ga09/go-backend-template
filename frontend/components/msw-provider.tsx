'use client'

import { useEffect, useState } from 'react'

const mockingEnabled = process.env.NEXT_PUBLIC_API_MOCKING === 'enabled'

/**
 * NEXT_PUBLIC_API_MOCKING=enabled のときのみmocks/browser.ts(msw/browser)を
 * 動的importして起動する。バックエンドを起動せずにpnpm devで画面を確認したい
 * 場合のための開発用途。Service Worker起動完了までchildrenの描画を遅らせ、
 * 初回フェッチがモック未適用のまま実backendへ飛ぶ競合を防ぐ。
 */
export function MSWProvider({ children }: { children: React.ReactNode }) {
  const [ready, setReady] = useState(!mockingEnabled)

  useEffect(() => {
    if (!mockingEnabled) {
      return
    }
    void (async () => {
      const { worker } = await import('@/mocks/browser')
      await worker.start({ onUnhandledRequest: 'bypass' })
      setReady(true)
    })()
  }, [])

  if (!ready) {
    return null
  }

  return <>{children}</>
}
