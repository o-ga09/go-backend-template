'use client'

import Link from 'next/link'
import { RequireAuth } from '@/components/require-auth'
import { useAuth } from '@/context/authContext'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { User as UserIcon } from 'lucide-react'

export default function MyPage() {
  return (
    <RequireAuth>
      <MyPageContent />
    </RequireAuth>
  )
}

function MyPageContent() {
  const { user } = useAuth()
  if (!user) {
    return null
  }

  return (
    <main className="container max-w-xl space-y-6 py-8">
      <h1 className="text-heading-2">マイページ</h1>

      <div className="flex items-center gap-4 rounded-lg border p-4">
        <Avatar className="h-12 w-12">
          <AvatarImage src={user.profileImage} alt={user.displayName} />
          <AvatarFallback>
            <UserIcon className="h-6 w-6" />
          </AvatarFallback>
        </Avatar>
        <p className="text-body-lg font-medium">{user.displayName}</p>
      </div>

      <Button asChild variant="outline">
        <Link href="/mypage/orders">注文履歴を見る</Link>
      </Button>
    </main>
  )
}
