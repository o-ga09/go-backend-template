import Image from 'next/image'

export default function Home() {
  return (
    <>
      <main className="flex min-h-screen flex-col items-center justify-center bg-gray-100">
        <h1 className="mb-8 text-4xl font-bold text-gray-800">Welcome to TaviNikkiy!</h1>
        <p className="mb-4 text-lg text-gray-600">旅の思い出を記録し、共有しましょう。</p>
        <Image
          src="/next.svg"
          alt="Travel Memories"
          width={600}
          height={400}
          className="rounded-lg shadow-lg"
        />
      </main>
    </>
  )
}
