import { geistMono, geistSans } from '@/lib/font'
import './globals.css'
import { AuthProvider } from '@/context/authContext'
import { ApiProvider } from '@/providers/apiProvider'
import NextTopLoader from 'nextjs-toploader'
import { topLoaderConfig } from '@/lib/loaderConfig'
import { Toaster } from '@/components/ui/sonner'
import { SessionProvider } from '@/providers/sessionProvider'
import { SiteHeader } from '@/components/site-header'
import { MSWProvider } from '@/components/msw-provider'

export const viewport = 'width=device-width, initial-scale=1'

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="ja">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <MSWProvider>
          <SessionProvider>
            <AuthProvider>
              <ApiProvider>
                <NextTopLoader {...topLoaderConfig} />
                <SiteHeader />
                {children}
                <Toaster />
              </ApiProvider>
            </AuthProvider>
          </SessionProvider>
        </MSWProvider>
      </body>
    </html>
  )
}
