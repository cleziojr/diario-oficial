import { BrowserRouter, Route, Routes } from 'react-router-dom'
import { NavBar } from './components/NavBar'
import { SearchPage } from './pages/SearchPage'
import { MateriaDetailPage } from './pages/MateriaDetailPage'
import { UploadPage } from './pages/UploadPage'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <main className="app-shell">
        <NavBar />
        <Routes>
          <Route path="/" element={<SearchPage />} />
          <Route path="/materia/:id" element={<MateriaDetailPage />} />
          <Route path="/upload" element={<UploadPage />} />
        </Routes>
      </main>
    </BrowserRouter>
  )
}

export default App
