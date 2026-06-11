import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  // Envia o cookie httpOnly de autenticação em toda requisição.
  withCredentials: true,
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    // Só força logout se acreditávamos estar logados (há usuário em sessão);
    // evita redirecionar em 401 de rotas públicas (ex.: probe inicial do /me).
    if (err.response?.status === 401 && sessionStorage.getItem('user')) {
      sessionStorage.removeItem('user')
      window.dispatchEvent(new Event('auth:logout'))
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export default api
