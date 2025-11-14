import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { login, saveAuthData } from '../services/api'
import './Login.css'

function Login() {
  const [usernameOrEmail, setUsernameOrEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    try {
      const data = await login(usernameOrEmail, password)

      if (data.token && data.user) {
        saveAuthData(data.token, data.user)
        navigate('/search')
      } else {
        setError('Respuesta inválida del servidor')
      }
    } catch (err) {
      console.error('Login error:', err)
      if (err.response?.status === 401) {
        setError('Usuario o contraseña incorrectos')
      } else if (err.response?.status === 400) {
        setError('Por favor complete todos los campos')
      } else {
        setError('Error al conectar con el servidor')
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-container">
      <div className="login-card card">
        <div className="login-header">
          <h1>💪 Gym Booking System</h1>
          <p>Reserva tus clases de gimnasio</p>
        </div>

        <form onSubmit={handleSubmit} className="login-form">
          <div className="form-group">
            <label htmlFor="username">Usuario o Email</label>
            <input
              id="username"
              type="text"
              placeholder="Ingresa tu usuario o email"
              value={usernameOrEmail}
              onChange={(e) => setUsernameOrEmail(e.target.value)}
              disabled={loading}
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Contraseña</label>
            <input
              id="password"
              type="password"
              placeholder="Ingresa tu contraseña"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading}
              required
            />
          </div>

          {error && <div className="error">{error}</div>}

          <button
            type="submit"
            className="btn btn-primary btn-block"
            disabled={loading}
          >
            {loading ? 'Iniciando sesión...' : 'Iniciar Sesión'}
          </button>
        </form>

        <div className="login-footer">
          <p className="test-credentials">
            <strong>Credenciales de prueba:</strong><br />
            Usuario: testuser | Contraseña: user123<br />
            Admin: admin | Contraseña: admin123
          </p>
        </div>
      </div>
    </div>
  )
}

export default Login
