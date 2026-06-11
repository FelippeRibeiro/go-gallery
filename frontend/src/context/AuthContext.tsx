import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { User } from '@/types'
import * as authApi from '@/api/auth'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  login: (email: string, senha: string) => Promise<void>
  register: (nome: string, email: string, senha: string) => Promise<void>
  loginWith: (user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

// O JWT vive num cookie httpOnly (inacessível ao JS). Apenas os dados do usuário
// ficam no sessionStorage, como cache para render imediato; a fonte da verdade
// é a rota /me, consultada no carregamento.
function readStoredUser(): User | null {
  try {
    const saved = sessionStorage.getItem('user')
    return saved ? (JSON.parse(saved) as User) : null
  } catch {
    return null
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(readStoredUser)

  const persist = (u: User) => {
    sessionStorage.setItem('user', JSON.stringify(u))
    setUser(u)
  }

  const clear = () => {
    sessionStorage.removeItem('user')
    setUser(null)
  }

  // Valida/restaura a sessão pelo cookie ao montar.
  useEffect(() => {
    authApi
      .me()
      .then((u) => persist(u))
      .catch(() => clear())
  }, [])

  useEffect(() => {
    const onLogout = () => clear()
    window.addEventListener('auth:logout', onLogout)
    return () => window.removeEventListener('auth:logout', onLogout)
  }, [])

  const login = async (email: string, senha: string) => {
    const res = await authApi.login(email, senha)
    persist(res.user)
  }

  const register = async (nome: string, email: string, senha: string) => {
    const res = await authApi.register(nome, email, senha)
    persist(res.user)
  }

  const loginWith = (u: User) => persist(u)

  const logout = () => {
    authApi.logout().catch(() => { /* limpa local de qualquer forma */ })
    clear()
  }

  return (
    <AuthContext.Provider value={{ user, isAuthenticated: !!user, login, register, loginWith, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
