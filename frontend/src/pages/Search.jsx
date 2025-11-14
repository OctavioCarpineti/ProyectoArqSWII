import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { searchSchedules, getAuthData, clearAuthData } from '../services/api'
import './Search.css'

function Search() {
  const [searchQuery, setSearchQuery] = useState('')
  const [category, setCategory] = useState('')
  const [dayOfWeek, setDayOfWeek] = useState('')
  const [instructor, setInstructor] = useState('')
  const [schedules, setSchedules] = useState([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [user, setUser] = useState(null)
  const navigate = useNavigate()

  useEffect(() => {
    const { user: userData } = getAuthData()
    setUser(userData)
    // Cargar schedules iniciales
    handleSearch()
  }, [])

  const handleSearch = async (e) => {
    if (e) e.preventDefault()

    setLoading(true)
    setError('')

    try {
      const params = {
        q: searchQuery || undefined,
        category: category || undefined,
        day_of_week: dayOfWeek || undefined,
        instructor: instructor || undefined,
        available: true, // Solo mostrar clases disponibles
        page: 1,
        size: 20
      }

      const data = await searchSchedules(params)
      setSchedules(data.results || [])
    } catch (err) {
      console.error('Search error:', err)
      setError('Error al buscar horarios. Intenta nuevamente.')
    } finally {
      setLoading(false)
    }
  }

  const handleLogout = () => {
    clearAuthData()
    navigate('/')
  }

  const handleScheduleClick = (scheduleId) => {
    navigate(`/schedule/${scheduleId}`)
  }

  const getDayName = (dayNum) => {
    const days = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado']
    return days[dayNum] || dayNum
  }

  return (
    <div className="container">
      <div className="navbar">
        <h2>💪 Gym Booking</h2>
        <div className="navbar-user">
          <span>Hola, {user?.username || 'Usuario'}!</span>
          <button onClick={handleLogout} className="btn btn-secondary">
            Cerrar Sesión
          </button>
        </div>
      </div>

      <div className="search-section card">
        <h1>Buscar Clases</h1>

        <form onSubmit={handleSearch} className="search-form">
          <div className="search-grid">
            <div className="form-group">
              <label htmlFor="search">Búsqueda general</label>
              <input
                id="search"
                type="text"
                placeholder="Buscar por nombre de actividad..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
              />
            </div>

            <div className="form-group">
              <label htmlFor="category">Categoría</label>
              <select
                id="category"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
              >
                <option value="">Todas</option>
                <option value="cardio">Cardio</option>
                <option value="strength">Fuerza</option>
                <option value="flexibility">Flexibilidad</option>
                <option value="dance">Baile</option>
                <option value="martial_arts">Artes Marciales</option>
                <option value="aquatic">Acuática</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="day">Día de la semana</label>
              <select
                id="day"
                value={dayOfWeek}
                onChange={(e) => setDayOfWeek(e.target.value)}
              >
                <option value="">Todos</option>
                <option value="1">Lunes</option>
                <option value="2">Martes</option>
                <option value="3">Miércoles</option>
                <option value="4">Jueves</option>
                <option value="5">Viernes</option>
                <option value="6">Sábado</option>
                <option value="0">Domingo</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="instructor">Instructor</label>
              <input
                id="instructor"
                type="text"
                placeholder="Nombre del instructor..."
                value={instructor}
                onChange={(e) => setInstructor(e.target.value)}
              />
            </div>
          </div>

          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? 'Buscando...' : '🔍 Buscar'}
          </button>
        </form>

        {error && <div className="error">{error}</div>}
      </div>

      <div className="results-section">
        {loading ? (
          <div className="loading">Cargando horarios...</div>
        ) : schedules.length === 0 ? (
          <div className="card no-results">
            <p>No se encontraron horarios disponibles con los filtros seleccionados.</p>
          </div>
        ) : (
          <div className="schedules-grid">
            {schedules.map((schedule) => (
              <div
                key={schedule.id}
                className="schedule-card card"
                onClick={() => handleScheduleClick(schedule.id)}
              >
                <div className="schedule-header">
                  <h3>{schedule.activity_name}</h3>
                  <span className="category-badge">{schedule.category}</span>
                </div>

                <div className="schedule-info">
                  <div className="info-row">
                    <span className="label">📅 Día:</span>
                    <span>{getDayName(schedule.day_of_week)}</span>
                  </div>
                  <div className="info-row">
                    <span className="label">🕐 Horario:</span>
                    <span>{schedule.start_time} - {schedule.end_time}</span>
                  </div>
                  <div className="info-row">
                    <span className="label">👨‍🏫 Instructor:</span>
                    <span>{schedule.instructor}</span>
                  </div>
                  <div className="info-row">
                    <span className="label">👥 Cupos:</span>
                    <span className={schedule.available_spots > 5 ? 'spots-available' : 'spots-limited'}>
                      {schedule.available_spots} disponibles de {schedule.max_capacity}
                    </span>
                  </div>
                </div>

                <button className="btn btn-primary btn-block">
                  Ver Detalles
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default Search
