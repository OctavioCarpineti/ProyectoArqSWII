import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { isAuthenticated } from './services/api'
import Login from './pages/Login'
import Search from './pages/Search'
import ScheduleDetails from './pages/ScheduleDetails'
import Congrats from './pages/Congrats'

// Componente para proteger rutas que requieren autenticación
const ProtectedRoute = ({ children }) => {
  return isAuthenticated() ? children : <Navigate to="/" replace />
}

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Login />} />
        <Route
          path="/search"
          element={
            <ProtectedRoute>
              <Search />
            </ProtectedRoute>
          }
        />
        <Route
          path="/schedule/:id"
          element={
            <ProtectedRoute>
              <ScheduleDetails />
            </ProtectedRoute>
          }
        />
        <Route
          path="/congrats"
          element={
            <ProtectedRoute>
              <Congrats />
            </ProtectedRoute>
          }
        />
      </Routes>
    </Router>
  )
}

export default App
