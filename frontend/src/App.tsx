import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { NavBar } from './components/NavBar'
import { SearchPage } from './pages/SearchPage'
import { MateriaDetailPage } from './pages/MateriaDetailPage'
import { UploadPage } from './pages/UploadPage'
import { AuthPage } from './pages/AuthPage'
import { AuthProvider } from './auth/AuthContext'
import { RequireAuth } from './auth/RequireAuth'
import './App.css'

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <main className="app-shell">
          <NavBar />
          <Routes>
            <Route path="/" element={<SearchPage />} />
            <Route path="/materia/:id" element={<MateriaDetailPage />} />
            <Route
              path="/upload"
              element={
                <RequireAuth>
                  <UploadPage />
                </RequireAuth>
              }
            />
            <Route path="/login" element={<AuthPage mode="login" />} />
            <Route path="/register" element={<AuthPage mode="register" />} />
          </Routes>
        </main>
      </BrowserRouter>
    </AuthProvider>
  )
}

export default App
