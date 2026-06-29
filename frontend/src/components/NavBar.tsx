import { NavLink } from 'react-router-dom'
import './NavBar.css'

export function NavBar() {
  return (
    <header className="nav-bar">
      <NavLink to="/" className="nav-bar__brand">
        Diário Oficial
      </NavLink>
      <nav>
        <NavLink to="/" end>
          Buscar
        </NavLink>
        <NavLink to="/upload">Enviar PDF</NavLink>
      </nav>
    </header>
  )
}
