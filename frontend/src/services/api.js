import axios from 'axios';

// Configuración de URLs base (desde docker-compose o localhost)
const USERS_API_URL = import.meta.env.VITE_USERS_API_URL || 'http://localhost:8080';
const SEARCH_API_URL = import.meta.env.VITE_SEARCH_API_URL || 'http://localhost:8083';
const BOOKINGS_API_URL = import.meta.env.VITE_BOOKINGS_API_URL || 'http://localhost:8082';

// Crear instancia de axios
const api = axios.create({
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptor para agregar token en cada request
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// ============================================
// AUTH & USERS
// ============================================

export const login = async (usernameOrEmail, password) => {
  const response = await axios.post(`${USERS_API_URL}/login`, {
    username_or_email: usernameOrEmail,
    password: password,
  });
  return response.data;
};

export const createUser = async (userData) => {
  const response = await axios.post(`${USERS_API_URL}/users`, userData);
  return response.data;
};

// ============================================
// SEARCH
// ============================================

export const searchSchedules = async (params = {}) => {
  const queryParams = new URLSearchParams();

  // Agregar wildcards para búsqueda parcial
  if (params.q) {
    const query = params.q.includes('*') ? params.q : `*${params.q}*`;
    queryParams.append('q', query);
  }
  if (params.category) queryParams.append('category', params.category);
  if (params.day_of_week) queryParams.append('day_of_week', params.day_of_week);
  if (params.instructor) queryParams.append('instructor', params.instructor);
  if (params.available !== undefined) queryParams.append('available', params.available);
  if (params.page) queryParams.append('page', params.page);
  if (params.size) queryParams.append('size', params.size);

  const response = await axios.get(
    `${SEARCH_API_URL}/search?${queryParams.toString()}`
  );
  return response.data;
};

export const getScheduleById = async (scheduleId) => {
  const response = await axios.get(`${SEARCH_API_URL}/search/${scheduleId}`);
  return response.data;
};

// ============================================
// BOOKINGS
// ============================================

export const createBooking = async (userId, scheduleId) => {
  const token = localStorage.getItem('token');
  const response = await axios.post(
    `${BOOKINGS_API_URL}/bookings`,
    {
      user_id: parseInt(userId),
      schedule_id: scheduleId,
    },
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  return response.data;
};

export const getMyBookings = async (userId) => {
  const token = localStorage.getItem('token');
  const response = await axios.get(
    `${BOOKINGS_API_URL}/bookings?user_id=${userId}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  return response.data;
};

export const cancelBooking = async (bookingId) => {
  const token = localStorage.getItem('token');
  const response = await axios.delete(
    `${BOOKINGS_API_URL}/bookings/${bookingId}`,
    {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  return response.data;
};

// ============================================
// HELPERS
// ============================================

export const saveAuthData = (token, user) => {
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(user));
};

export const getAuthData = () => {
  const token = localStorage.getItem('token');
  const userStr = localStorage.getItem('user');
  const user = userStr ? JSON.parse(userStr) : null;
  return { token, user };
};

export const clearAuthData = () => {
  localStorage.removeItem('token');
  localStorage.removeItem('user');
};

export const isAuthenticated = () => {
  return !!localStorage.getItem('token');
};
