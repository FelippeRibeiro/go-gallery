import { createContext, useContext, useState, type ReactNode } from 'react'
import type { User } from '@/types'
import * as authApi from '@/api/auth'

interface AuthContextType {
  user: User | null
  isAuthenticated: boolean
  login: (email: string, senha: string) => Promise<void>
  register: (nome: string, email: string, senha: string) => Promise<void>
  loginWith: (token: string, user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(() => {
    try {
      const saved = localStorage.getItem('user')
      return saved ? JSON.parse(saved) : null
    } catch {
      return null
    }
  })

  const persist = (token: string, u: User) => {
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(u))
    setUser(u)
  }

  const login = async (email: string, senha: string) => {
    const res = await authApi.login(email, senha)
    persist(res.token, res.user)
  }

  const register = async (nome: string, email: string, senha: string) => {
    const res = await authApi.register(nome, email, senha)
    persist(res.token, res.user)
  }

  const loginWith = (token: string, u: User) => persist(token, u)

  const logout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setUser(null)
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
