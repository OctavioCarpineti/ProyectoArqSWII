import { useEffect } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import './Congrats.css'

function Congrats() {
  const navigate = useNavigate()
  const location = useLocation()
  const { scheduleName, scheduleTime, scheduleDay } = location.state || {}

  useEffect(() => {
    // Si no hay datos, redirigir a search
    if (!scheduleName) {
      navigate('/search')
    }
  }, [scheduleName, navigate])

  return (
    <div className="congrats-container">
      <div className="congrats-card card">
        <div className="congrats-icon">
          🎉
        </div>

        <h1>¡Reserva Confirmada!</h1>

        <p className="congrats-message">
          Tu clase ha sido reservada exitosamente
        </p>

        {scheduleName && (
          <div className="booking-summary">
            <h3>Detalles de tu reserva:</h3>
            <div className="summary-items">
              <div className="summary-item">
                <span className="summary-icon">📚</span>
                <div>
                  <span className="summary-label">Actividad</span>
                  <span className="summary-value">{scheduleName}</span>
                </div>
              </div>

              <div className="summary-item">
                <span className="summary-icon">📅</span>
                <div>
                  <span className="summary-label">Día</span>
                  <span className="summary-value">{scheduleDay}</span>
                </div>
              </div>

              <div className="summary-item">
                <span className="summary-icon">🕐</span>
                <div>
                  <span className="summary-label">Horario</span>
                  <span className="summary-value">{scheduleTime}</span>
                </div>
              </div>
            </div>
          </div>
        )}

        <div className="congrats-actions">
          <button
            onClick={() => navigate('/search')}
            className="btn btn-primary btn-large"
          >
            Buscar más clases
          </button>
        </div>

        <div className="congrats-footer">
          <p>
            💪 ¡Nos vemos en tu clase! Recuerda llegar 10 minutos antes.
          </p>
        </div>
      </div>
    </div>
  )
}

export default Congrats
