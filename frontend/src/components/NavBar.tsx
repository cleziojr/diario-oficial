import { NavLink } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import './NavBar.css'

export function NavBar() {
  const { user, logout } = useAuth()

  return (
    <header className="nav-bar">
      <NavLink to="/" className="nav-bar__brand">
        Diário Oficial
      </NavLink>
      <nav>
        <NavLink to="/" end>
          Buscar
        </NavLink>
        {user && <NavLink to="/upload">Enviar PDF</NavLink>}
        {user ? (
          <span className="nav-bar__user">
            <span className="nav-bar__email" title={user.email}>
              {user.email}
            </span>
            <button type="button" className="nav-bar__logout" onClick={logout}>
              Sair
            </button>
          </span>
        ) : (
          <>
            <NavLink to="/login">Entrar</NavLink>
            <NavLink to="/register">Cadastrar</NavLink>
          </>
        )}
      </nav>
    </header>
  )
}
