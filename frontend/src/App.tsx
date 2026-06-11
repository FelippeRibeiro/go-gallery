import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'sonner'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import Login from '@/pages/Login'
import Register from '@/pages/Register'
import Dashboard from '@/pages/Dashboard'
import AlbumCreate from '@/pages/AlbumCreate'
import AlbumDetail from '@/pages/AlbumDetail'
import AcceptInvite from '@/pages/AcceptInvite'
import PublicGallery from '@/pages/PublicGallery'
import PublicAlbum from '@/pages/PublicAlbum'
import Order from '@/pages/Order'
import PublicPhotographer from '@/pages/PublicPhotographer'
import Discover from '@/pages/Discover'
import Profile from '@/pages/Profile'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth()
  return isAuthenticated ? <>{children}</> : <Navigate to="/login" replace />
}

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Toaster theme="dark" richColors position="top-right" closeButton />
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          <Route path="/accept-invite" element={<AcceptInvite />} />
          <Route path="/gallery" element={<PublicGallery />} />
          <Route path="/gallery/:id" element={<PublicAlbum />} />
          <Route path="/p/:id" element={<PublicPhotographer />} />
          <Route path="/discover" element={<Discover />} />

          {/* Protected routes */}
          <Route
            path="/dashboard"
            element={<PrivateRoute><Dashboard /></PrivateRoute>}
          />
          <Route
            path="/albums/new"
            element={<PrivateRoute><AlbumCreate /></PrivateRoute>}
          />
          <Route
            path="/albums/:id"
            element={<PrivateRoute><AlbumDetail /></PrivateRoute>}
          />
          <Route
            path="/orders/:id"
            element={<PrivateRoute><Order /></PrivateRoute>}
          />
          <Route
            path="/profile"
            element={<PrivateRoute><Profile /></PrivateRoute>}
          />

          <Route path="*" element={<Navigate to="/discover" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
