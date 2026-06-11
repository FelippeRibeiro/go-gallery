import Layout from '@/components/Layout'
import PublicHeader from '@/components/PublicHeader'
import { useAuth } from '@/context/AuthContext'

export default function PublicOrAppLayout({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()

  if (isAuthenticated) {
    return <Layout>{children}</Layout>
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <PublicHeader />
      <main className="max-w-6xl mx-auto px-4 py-10">{children}</main>
    </div>
  )
}
