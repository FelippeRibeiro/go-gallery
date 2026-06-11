import api from './client'
import type { AuthResponse, User } from '../types'

export const register = (nome: string, email: string, senha: string) =>
  api.post<AuthResponse>('/auth/register', { nome, email, senha }).then((r) => r.data)

export const login = (email: string, senha: string) =>
  api.post<AuthResponse>('/auth/login', { email, senha }).then((r) => r.data)

export const acceptInvite = (token: string, nome: string, senha: string) =>
  api
    .post<AuthResponse & { album_id: number }>('/auth/accept-invite', { token, nome, senha })
    .then((r) => r.data)

// Decodifica o cookie JWT no servidor e devolve o usuário atual.
export const me = () => api.get<{ user: User }>('/me').then((r) => r.data.user)

// Limpa o cookie de autenticação no servidor.
export const logout = () => api.post('/auth/logout').then((r) => r.data)
