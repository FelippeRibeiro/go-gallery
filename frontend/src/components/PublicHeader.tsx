import { Link } from 'react-router-dom'
import { Camera } from 'lucide-react'
import { useAuth } from '@/context/AuthContext'
import { Button } from '@/components/ui/button'

export default function PublicHeader() {
  const { isAuthenticated } = useAuth()

  return (
    <header className="border-b border-border px-6 py-4 flex items-center justify-between">
      <Link to="/discover" className="flex items-center gap-2">
        <Camera className="h-5 w-5 text-primary" />
        <span className="text-lg font-bold tracking-tight">Go Gallery</span>
      </Link>
      <div className="flex items-center gap-2">
        {isAuthenticated ? (
          <Button variant="ghost" size="sm" asChild>
            <Link to="/dashboard">Meus álbuns</Link>
          </Button>
        ) : (
          <>
            <Button variant="ghost" size="sm" asChild>
              <Link to="/login">Entrar</Link>
            </Button>
            <Button size="sm" asChild>
              <Link to="/register">Criar conta</Link>
            </Button>
          </>
        )}
      </div>
    </header>
  )
}
