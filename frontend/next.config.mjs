import path from 'path'

/** @type {import('next').NextConfig} */
const nextConfig = {
  compress: true,
  reactStrictMode: true,
  output: 'standalone',

  // 静的アセットのベースURLを環境変数で制御
  // manifest.jsonなどの特定ファイルは同一オリジンから配信する必要があるため、
  // assetPrefixは_next/staticのみに適用される
  assetPrefix:
    process.env.NODE_ENV === 'production' ? process.env.NEXT_PUBLIC_FRONT_STATIC_URL : '',

  // 画像設定
  images: {
    unoptimized: true, // R2で配信するため最適化を無効化
    dangerouslyAllowSVG: true,
    formats: ['image/webp'],
    remotePatterns: [
      {
        protocol: 'https',
        hostname: '**.r2.cloudflarestorage.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: '**.r2.dev',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'lh3.googleusercontent.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'firebasestorage.googleapis.com',
        port: '',
        pathname: '/**',
      },
      {
        protocol: 'https',
        hostname: 'example.com',
        port: '',
        pathname: '/**',
      },
    ],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },

  // TypeScript設定
  typescript: {
    ignoreBuildErrors: true, // Base64画像に関する型エラーを無視
  },

  // 実験的機能
  experimental: {
    optimizePackageImports: ['framer-motion'],
    serverActions: {
      bodySizeLimit: '2mb',
    },
    largePageDataBytes: 128 * 1000,
    scrollRestoration: true,
  },

  // HTTPヘッダー設定
  headers: async () => [
    {
      // 動的コンテンツ用のセキュリティヘッダー
      source: '/((?!_next/static|_next/image|favicon.ico).*)',
      headers: [
        {
          key: 'Strict-Transport-Security',
          value: 'max-age=31536000; includeSubDomains; preload',
        },
        {
          key: 'X-Content-Type-Options',
          value: 'nosniff',
        },
        {
          key: 'X-Frame-Options',
          value: 'DENY',
        },
        {
          key: 'Referrer-Policy',
          value: 'strict-origin-when-cross-origin',
        },
        {
          key: 'Permissions-Policy',
          value: 'camera=(), microphone=(), geolocation=(self)',
        },
        {
          key: 'Content-Security-Policy',
          value: [
            "default-src 'self'",
            `script-src 'self' 'unsafe-eval' 'unsafe-inline' https://static.tavinikkiy.com https://maps.googleapis.com https://www.gstatic.com https://apis.google.com https://accounts.google.com https://www.googletagmanager.com https://static.cloudflareinsights.com`,
            `style-src 'self' 'unsafe-inline' https://static.tavinikkiy.com https://fonts.googleapis.com`,
            "font-src 'self' https://fonts.gstatic.com",
            "img-src 'self' data: https: blob:",
            "connect-src 'self' https://api.tavinikkiy.com https://identitytoolkit.googleapis.com https://securetoken.googleapis.com https://firestore.googleapis.com https://maps.googleapis.com https://*.googleapis.com https://www.google-analytics.com https://analytics.google.com http://localhost:8080",
            "frame-src 'self' https://accounts.google.com https://tavinikkiy.firebaseapp.com https://*.firebaseapp.com",
            "worker-src 'self' blob:",
            "object-src 'none'",
            "base-uri 'self'",
            "form-action 'self'",
            "manifest-src 'self' https://static.tavinikkiy.com",
            'upgrade-insecure-requests',
          ].join('; '),
        },
        {
          key: 'Cross-Origin-Opener-Policy',
          value: 'same-origin-allow-popups',
        },
      ],
    },
    {
      source: '/manifest.json',
      headers: [
        {
          key: 'Content-Type',
          value: 'application/manifest+json',
        },
        {
          key: 'Cache-Control',
          value: 'public, max-age=86400',
        },
      ],
    },
    {
      // 静的アセット用の長期キャッシュヘッダー
      source: '/_next/static/(.*)',
      headers: [
        {
          key: 'Cache-Control',
          value: 'public, max-age=31536000, immutable',
        },
      ],
    },
  ],

  // Rewrites設定
  rewrites: async () => [
    {
      // manifest.jsonを同一オリジンから配信（assetPrefixの影響を受けないようにする）
      source: '/manifest.json',
      destination: '/manifest.json',
    },
  ],

  // Webpack設定
  webpack: (config, { isServer }) => {
    // エイリアス設定
    config.resolve.alias = {
      ...config.resolve.alias,
      '@': path.resolve('./src'),
    }

    // framer-motionの最適化
    if (!isServer) {
      config.resolve.fallback = {
        ...config.resolve.fallback,
        fs: false,
        path: false,
      }
    }

    return config
  },
}
