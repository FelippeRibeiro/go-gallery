import api from './client'
import type { Perfil } from '../types'

export const getProfile = () =>
  api.get<Perfil>('/profile').then((r) => r.data)

export const updateProfile = (bio: string) =>
  api.patch('/profile', { bio })

export const uploadProfilePhoto = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return api.post<{ foto_perfil: string }>('/profile/photo', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  }).then((r) => r.data)
}
