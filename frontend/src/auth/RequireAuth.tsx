import { Navigate, useLocation } from 'react-router-dom'
import type { ReactNode } from 'react'
import { useAuth } from './AuthContext'

// RequireAuth protege rotas: sem usuário logado, redireciona para /login.
export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  const location = useLocation()

  if (loading) {
    return <p style={{ padding: 24 }}>Carregando…</p>
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return <>{children}</>
}
