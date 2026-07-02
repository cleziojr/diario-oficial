import { useState } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import './AuthPage.css'

type Props = { mode: 'login' | 'register' }

export function AuthPage({ mode }: Props) {
  const { login, register } = useAuth()
  const navigate = useNavigate()
  const location = useLocation() as { state?: { from?: string } }

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const isRegister = mode === 'register'
  const title = isRegister ? 'Criar conta' : 'Entrar'

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      if (isRegister) {
        await register(email, password)
      } else {
        await login(email, password)
      }
      navigate(location.state?.from ?? '/upload', { replace: true })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Falha na autenticação')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="auth-page">
      <h1>{title}</h1>
      <p className="auth-page__lede">
        {isRegister
          ? 'Cadastre-se para enviar Diários Oficiais. A busca é pública e não exige conta.'
          : 'Entre para enviar Diários Oficiais. A busca continua pública.'}
      </p>

      <form className="auth-page__form" onSubmit={handleSubmit}>
        <label>
          <span>E-mail</span>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
            required
          />
        </label>
        <label>
          <span>Senha</span>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete={isRegister ? 'new-password' : 'current-password'}
            minLength={6}
            required
          />
        </label>

        {error && (
          <p className="auth-page__error" role="alert">
            {error}
          </p>
        )}

        <button type="submit" disabled={busy}>
          {busy ? 'Enviando…' : title}
        </button>
      </form>

      <p className="auth-page__switch">
        {isRegister ? (
          <>
            Já tem conta? <Link to="/login">Entrar</Link>
          </>
        ) : (
          <>
            Não tem conta? <Link to="/register">Cadastre-se</Link>
          </>
        )}
      </p>
    </div>
  )
}
