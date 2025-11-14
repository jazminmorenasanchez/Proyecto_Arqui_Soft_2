// API Client centralizado para todas las llamadas a las APIs

const USERS_API_BASE = process.env.NEXT_PUBLIC_USERS_API_URL || "http://localhost:8081"
const ACTIVITIES_API_BASE = process.env.NEXT_PUBLIC_ACTIVITIES_API_URL || "http://localhost:8082"
const SEARCH_API_BASE = process.env.NEXT_PUBLIC_SEARCH_API_URL || "http://localhost:8083"

// Tipos de datos
export interface Activity {
  id: number
  ownerUserId: string
  categoria: string
  nombre: string
  ubicacion: string
  instructor?: string
  precioBase: number
  rating: number
  updatedAt: string
}

export interface Session {
  id: number
  activityId: number
  fecha: string // YYYY-MM-DD
  inicio: string // HH:mm
  fin: string // HH:mm
  capacidad: number
  createdAt: string
  updatedAt: string
}

export interface Enrollment {
  id: number
  activityId: number
  sessionId: number
  userId: string
  precioFinal: number
  estado: string
  createdAt: string
}

export interface CreateActivityRequest {
  categoria: string
  nombre: string
  ubicacion: string
  instructor?: string
  precioBase: number
}

export interface CreateSessionRequest {
  activityId: string
  startTime: string
  endTime: string
  capacity: number
}

export interface ActivitiesListResponse {
  activities: Activity[]
  total: number
  skip: number
  limit: number
}

// Helper para hacer requests con autenticación
async function fetchWithAuth(
  url: string,
  options: RequestInit = {},
  token?: string | null
): Promise<Response> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...options.headers,
  }

  if (token) {
    headers["Authorization"] = `Bearer ${token}`
  }

  try {
    const response = await fetch(url, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ error: "Unknown error" }))
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
    }

    return response
  } catch (error) {
    // Si es un error de red (fail to fetch), proporcionar un mensaje más claro
    if (error instanceof TypeError && error.message.includes("fetch")) {
      throw new Error("Error de conexión. Verifica que el servidor esté en ejecución y que la URL sea correcta.")
    }
    throw error
  }
}

// ==================== USERS API ====================

export const usersAPI = {
  async login(login: string, password: string) {
    const response = await fetchWithAuth(`${USERS_API_BASE}/auth/login`, {
      method: "POST",
      body: JSON.stringify({ login, password }),
    })
    return response.json()
  },

  async register(username: string, email: string, password: string) {
    const response = await fetchWithAuth(`${USERS_API_BASE}/users`, {
      method: "POST",
      body: JSON.stringify({
        username,
        email,
        password,
        rol: "user",
      }),
    })
    return response.json()
  },

  async getUserById(id: string, token: string) {
    const response = await fetchWithAuth(`${USERS_API_BASE}/users/${id}`, {}, token)
    return response.json()
  },
}

// ==================== ACTIVITIES API ====================

export const activitiesAPI = {
  // Listar actividades con paginación
  async list(skip: number = 0, limit: number = 20, token?: string | null): Promise<ActivitiesListResponse> {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities?skip=${skip}&limit=${limit}`,
      {},
      token
    )
    return response.json()
  },

  // Obtener actividad por ID
  async getById(id: number, token?: string | null): Promise<Activity> {
    const response = await fetchWithAuth(`${ACTIVITIES_API_BASE}/activities/${id}`, {}, token)
    return response.json()
  },

  // Crear actividad (admin only)
  async create(data: CreateActivityRequest, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    )
    return response.json()
  },

  // Actualizar actividad (admin only)
  async update(id: number, data: Partial<CreateActivityRequest>, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities/${id}`,
      {
        method: "PUT",
        body: JSON.stringify(data),
      },
      token
    )
    return response.json()
  },

  // Eliminar actividad (admin only)
  async delete(id: number, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities/${id}`,
      {
        method: "DELETE",
      },
      token
    )
    return response.json()
  },
}

// ==================== SESSIONS API ====================

export interface CreateSessionPayload {
  fecha: string // YYYY-MM-DD
  inicio: string // HH:mm
  fin: string // HH:mm
  capacidad: number
}

export const sessionsAPI = {
  // Obtener sesiones de una actividad
  async getByActivity(activityId: number, token?: string | null): Promise<Session[]> {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities/${activityId}/sessions`,
      {},
      token
    )
    return response.json()
  },

  // Obtener sesión por ID
  async getById(sessionId: number, token?: string | null): Promise<Session> {
    const response = await fetchWithAuth(`${ACTIVITIES_API_BASE}/sessions/${sessionId}`, {}, token)
    return response.json()
  },

  // Crear sesión para una actividad específica (admin only)
  async createForActivity(activityId: number, data: CreateSessionPayload, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/activities/${activityId}/sessions`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    )
    return response.json()
  },

  // Crear sesión (admin only) - endpoint alternativo
  async create(data: CreateSessionRequest, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/sessions`,
      {
        method: "POST",
        body: JSON.stringify(data),
      },
      token
    )
    return response.json()
  },
}

// ==================== ENROLLMENTS API ====================

export const enrollmentsAPI = {
  // Inscribirse en una sesión
  async enroll(sessionId: number, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/enrollments`,
      {
        method: "POST",
        body: JSON.stringify({ sessionId: sessionId.toString() }),
      },
      token
    )
    return response.json()
  },

  // Obtener inscripciones de un usuario
  async getByUser(userId: string, token?: string | null): Promise<Enrollment[]> {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/enrollments/by-user/${userId}`,
      {},
      token
    )
    return response.json()
  },

  // Cancelar inscripción
  async cancel(enrollmentId: number, token: string) {
    const response = await fetchWithAuth(
      `${ACTIVITIES_API_BASE}/enrollments/${enrollmentId}/cancel`,
      {
        method: "PATCH",
      },
      token
    )
    return response.json()
  },
}

// ==================== SEARCH API ====================

export interface SearchParams {
  query?: string
  sport?: string
  site?: string
  date?: string // yyyy-mm-dd
  sort?: string
  page?: number
  size?: number
}

export interface SearchDoc {
  id: string
  activity_id: string
  session_id: string
  name: string
  sport: string
  site: string
  instructor: string
  start_dt: string
  end_dt: string
  difficulty: number
  price: number
  tags: string[]
  updated_dt: string
}

export interface SearchResult {
  total: number
  page: number
  size: number
  docs: SearchDoc[]
}

export const searchAPI = {
  async search(params: SearchParams, token?: string | null): Promise<SearchResult> {
    const queryParams = new URLSearchParams()
    if (params.query) queryParams.set("query", params.query)
    if (params.sport) queryParams.set("sport", params.sport)
    if (params.site) queryParams.set("site", params.site)
    if (params.date) queryParams.set("date", params.date)
    if (params.sort) queryParams.set("sort", params.sort)
    if (params.page) queryParams.set("page", params.page.toString())
    if (params.size) queryParams.set("size", params.size.toString())

    const response = await fetchWithAuth(
      `${SEARCH_API_BASE}/search?${queryParams.toString()}`,
      {},
      token
    )
    return response.json()
  },
}
