import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getScheduleById, createBooking, getAuthData } from '../services/api'
import './ScheduleDetails.css'

function ScheduleDetails() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [schedule, setSchedule] = useState(null)
  const [loading, setLoading] = useState(true)
  const [bookingLoading, setBookingLoading] = useState(false)
  const [error, setError] = useState('')
  const [user, setUser] = useState(null)

  useEffect(() => {
    const { user: userData } = getAuthData()
    setUser(userData)
    loadScheduleDetails()
  }, [id])

  const loadScheduleDetails = async () => {
    setLoading(true)
    setError('')

    try {
      const data = await getScheduleById(id)
      setSchedule(data)
    } catch (err) {
      console.error('Error loading schedule:', err)
      setError('Error al cargar los detalles del horario')
    } finally {
      setLoading(false)
    }
  }

  const handleBooking = async () => {
    if (!user || !user.id) {
      setError('Error: Usuario no autenticado')
      return
    }

    setBookingLoading(true)
    setError('')

    try {
      await createBooking(user.id, schedule.id)
      navigate('/congrats', {
        state: {
          scheduleName: schedule.activity_name,
          scheduleTime: `${schedule.start_time} - ${schedule.end_time}`,
          scheduleDay: getDayName(schedule.day_of_week)
        }
      })
    } catch (err) {
      console.error('Booking error:', err)
      if (err.response?.status === 409 || err.response?.status === 400) {
        setError(err.response?.data?.message || 'Ya tienes una reserva para este horario')
      } else if (err.response?.status === 404) {
        setError('Horario no encontrado o sin cupos disponibles')
      } else {
        setError('Error al crear la reserva. Intenta nuevamente.')
      }
    } finally {
      setBookingLoading(false)
    }
  }

  const getDayName = (day) => {
    // Mapeo de inglés a español
    const dayMapping = {
      'sunday': 'Domingo',
      'monday': 'Lunes',
      'tuesday': 'Martes',
      'wednesday': 'Miércoles',
      'thursday': 'Jueves',
      'friday': 'Viernes',
      'saturday': 'Sábado'
    }

    // Si es un número, convertir
    if (!isNaN(day)) {
      const days = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado']
      return days[parseInt(day)] || day
    }

    // Si es string en inglés, traducir
    return dayMapping[day.toLowerCase()] || day
  }

  if (loading) {
    return (
      <div className="container">
        <div className="loading">Cargando detalles...</div>
      </div>
    )
  }

  if (error && !schedule) {
    return (
      <div className="container">
        <div className="card">
          <div className="error">{error}</div>
          <button onClick={() => navigate('/search')} className="btn btn-secondary">
            Volver a la búsqueda
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="container">
      <div className="navbar">
        <h2>💪 Gym Booking</h2>
        <div className="navbar-user">
          <span>Hola, {user?.username || 'Usuario'}!</span>
        </div>
      </div>

      <button onClick={() => navigate('/search')} className="btn btn-secondary back-btn">
        ← Volver a la búsqueda
      </button>

      <div className="details-card card">
        <div className="details-header">
          <div>
            <h1>{schedule.activity_name}</h1>
            <span className="category-badge">{schedule.category}</span>
          </div>
          <div className="availability-badge">
            {schedule.available_spots > 0 ? (
              <span className="available">✓ Disponible</span>
            ) : (
              <span className="unavailable">✗ Sin cupos</span>
            )}
          </div>
        </div>

        <div className="details-grid">
          <div className="detail-item">
            <div className="detail-icon">📅</div>
            <div className="detail-content">
              <span className="detail-label">Día de la semana</span>
              <span className="detail-value">{getDayName(schedule.day_of_week)}</span>
            </div>
          </div>

          <div className="detail-item">
            <div className="detail-icon">🕐</div>
            <div className="detail-content">
              <span className="detail-label">Horario</span>
              <span className="detail-value">{schedule.start_time} - {schedule.end_time}</span>
            </div>
          </div>

          <div className="detail-item">
            <div className="detail-icon">👨‍🏫</div>
            <div className="detail-content">
              <span className="detail-label">Instructor</span>
              <span className="detail-value">{schedule.instructor}</span>
            </div>
          </div>

          <div className="detail-item">
            <div className="detail-icon">👥</div>
            <div className="detail-content">
              <span className="detail-label">Cupos disponibles</span>
              <span className={`detail-value ${schedule.available_spots > 5 ? 'spots-available' : 'spots-limited'}`}>
                {schedule.available_spots} de {schedule.max_capacity}
              </span>
            </div>
          </div>

          <div className="detail-item">
            <div className="detail-icon">📍</div>
            <div className="detail-content">
              <span className="detail-label">Ubicación</span>
              <span className="detail-value">{schedule.location || 'Sala principal'}</span>
            </div>
          </div>

          <div className="detail-item">
            <div className="detail-icon">ℹ️</div>
            <div className="detail-content">
              <span className="detail-label">Actividad ID</span>
              <span className="detail-value">{schedule.activity_id}</span>
            </div>
          </div>
        </div>

        {schedule.description && (
          <div className="description-section">
            <h3>Descripción</h3>
            <p>{schedule.description}</p>
          </div>
        )}

        {error && <div className="error">{error}</div>}

        <div className="action-section">
          {schedule.available_spots > 0 ? (
            <button
              onClick={handleBooking}
              className="btn btn-primary btn-large"
              disabled={bookingLoading}
            >
              {bookingLoading ? 'Procesando reserva...' : '✓ Reservar esta clase'}
            </button>
          ) : (
            <button className="btn btn-primary btn-large" disabled>
              Sin cupos disponibles
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

export default ScheduleDetails
