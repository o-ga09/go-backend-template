'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { useAuth } from '@/context/authContext'
import { Skeleton } from '@/components/ui/skeleton'

/**
 * 認証必須画面をラップする。未ログイン(ロード完了後にuserがnull)の場合は
 * 商品一覧へリダイレクトする(ログイン導線はヘッダーのログインボタン)。
 */
export function RequireAuth({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!loading && !user) {
      toast.error('ログインが必要です')
      router.replace('/')
    }
  }, [loading, user, router])

  if (loading || !user) {
    return (
      <div className="container space-y-4 py-8">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-32 w-full" />
      </div>
    )
  }

  return <>{children}</>
}
