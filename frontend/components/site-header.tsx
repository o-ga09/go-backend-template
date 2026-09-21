'use client'

import Link from 'next/link'
import { LogOut, ShoppingCart, User as UserIcon } from 'lucide-react'
import { useAuth } from '@/context/authContext'
import { useCart } from '@/api/cart/hooks'
import { cartItemCount } from '@/api/cart/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export function SiteHeader() {
  const { user, loading, login, logout } = useAuth()
  const { data: cart } = useCart(!!user)
  const itemCount = cartItemCount(cart)

  return (
    <header className="border-b border-border bg-background">
      <div className="container flex h-16 items-center justify-between">
        <Link href="/" className="text-heading-4 font-bold">
          Go Template Shop
        </Link>

        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" asChild className="relative">
            <Link href="/cart" aria-label="カート">
              <ShoppingCart className="h-5 w-5" />
              {itemCount > 0 && (
                <Badge className="absolute -right-1 -top-1 h-5 min-w-5 justify-center px-1">
                  {itemCount}
                </Badge>
              )}
            </Link>
          </Button>

          {loading ? null : user ? (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <Avatar className="h-8 w-8">
                    <AvatarImage src={user.profileImage} alt={user.displayName} />
                    <AvatarFallback>
                      <UserIcon className="h-4 w-4" />
                    </AvatarFallback>
                  </Avatar>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem asChild>
                  <Link href="/mypage">マイページ</Link>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                  <Link href="/mypage/orders">注文履歴</Link>
                </DropdownMenuItem>
                <DropdownMenuItem onSelect={() => void logout()}>
                  <LogOut className="mr-2 h-4 w-4" />
                  ログアウト
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          ) : (
            <Button onClick={() => void login()}>ログイン</Button>
          )}
        </div>
      </div>
    </header>
  )
}
