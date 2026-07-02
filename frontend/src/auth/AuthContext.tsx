import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import * as authApi from '../api/auth'
import type { User } from '../api/auth'
import { clearTokens, getAccessToken } from '../api/client'

type AuthContextValue = {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!getAccessToken()) {
      setLoading(false)
      return
    }
    authApi
      .getMe()
      .then(setUser)
      .catch(() => clearTokens())
      .finally(() => setLoading(false))
  }, [])

  async function login(email: string, password: string) {
    setUser(await authApi.login(email, password))
  }

  async function register(email: string, password: string) {
    setUser(await authApi.register(email, password))
  }

  function logout() {
    authApi.logout()
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth deve ser usado dentro de <AuthProvider>')
  return ctx
}
