'use client'

import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { useSession, signIn, signOut as nextAuthSignOut } from 'next-auth/react'
import { useRouter } from 'next/navigation'

const baseURL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080'

type AuthContextType = {
  user: User | null
  loading: boolean
  login: () => Promise<void>
  logout: () => Promise<void>
  refetchUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const { data: session, status } = useSession()
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  // ユーザー情報を取得
  const fetchUser = async () => {
    try {
      const response = await fetch(`${baseURL}/api/auth/user`, {
        credentials: 'include',
      })

      if (response.ok) {
        const data = await response.json()
        setUser(data)
      } else {
        setUser(null)
      }
    } catch (error) {
      console.error('Failed to fetch user:', error)
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  // セッションが変更されたときにユーザー情報を取得
  useEffect(() => {
    if (status === 'loading') {
      setLoading(true)
      return
    }

    if (status === 'authenticated' && session) {
      fetchUser()
    } else {
      setUser(null)
      setLoading(false)
    }
  }, [session, status])

  const login = async () => {
    try {
      // NextAuthのGoogle認証を実行
      const result = await signIn('google', {
        redirect: false,
      })

      if (!result?.ok) {
        throw new Error('Sign in failed')
      }

      // ユーザー情報を取得
      let userRes = await fetch(`${baseURL}/api/auth/user`, {
        credentials: 'include',
      })

      console.log('User fetch response status:', userRes.status)

      // ユーザーが存在しない場合は新規作成
      if (!userRes.ok) {
        // セッション情報から必要なデータを取得
        const sessionRes = await fetch('/api/auth/session')
        const sessionData = await sessionRes.json()

        userRes = await fetch(`${baseURL}/api/users`, {
          method: 'POST',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            uid: sessionData.userId,
            displayName: sessionData.user.name,
            profileImage: sessionData.user.image,
          }),
        })
      }

      const userData: User = await userRes.json()
      console.log('Logged in user data:', userData)
      setUser(userData)
      router.push(`/profile/${userData.name}`)
    } catch (error) {
      setUser(null)
      throw error
    }
  }

  const logout = async () => {
    try {
      await fetch(`${baseURL}/api/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      })

      // NextAuthのセッションをクリア
      await nextAuthSignOut({ redirect: false })
      setUser(null)
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  const refetchUser = async () => {
    setLoading(true)
    await fetchUser()
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refetchUser }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
