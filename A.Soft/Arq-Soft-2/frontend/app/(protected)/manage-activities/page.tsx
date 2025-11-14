"use client"

import { useEffect, useState } from "react"
import { useAuth } from "@/context/auth-context"
import { useRouter } from "next/navigation"
import ActivityForm from "@/components/activity-form"
import ActivityList from "@/components/activity-list"
import SessionForm from "@/components/session-form"
import { activitiesAPI, type Activity } from "@/lib/api"

export default function ManageActivitiesPage() {
  const { user, token } = useAuth()
  const router = useRouter()
  const [activities, setActivities] = useState<Activity[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [showSessionForm, setShowSessionForm] = useState(false)
  const [selectedActivity, setSelectedActivity] = useState<Activity | null>(null)
  const [editingActivity, setEditingActivity] = useState<Activity | null>(null)

  // Redirect non-admin users
  useEffect(() => {
    if (user && user.role !== "admin") {
      router.push("/home")
    }
  }, [user, router])

  // Fetch activities
  const fetchActivities = async () => {
    try {
      setIsLoading(true)
      setError(null)
      const data = await activitiesAPI.list(0, 100, token)
      setActivities(data.activities || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred")
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    if (token) {
      fetchActivities()
    }
  }, [token])

  const handleActivityCreated = () => {
    setShowForm(false)
    setEditingActivity(null)
    fetchActivities()
  }

  const handleActivityEdit = (activity: Activity) => {
    setEditingActivity(activity)
    setShowForm(true)
    setShowSessionForm(false)
  }

  const handleCancelEdit = () => {
    setEditingActivity(null)
    setShowForm(false)
  }

  const handleActivityDeleted = () => {
    fetchActivities()
  }

  const handleSessionCreated = () => {
    setShowSessionForm(false)
    setSelectedActivity(null)
  }

  const handleCreateSessionClick = () => {
    if (activities.length === 0) {
      setError("Primero debes crear al menos una actividad")
      return
    }
    setShowSessionForm(!showSessionForm)
    if (!showSessionForm) {
      setShowForm(false) // Cerrar formulario de actividad si está abierto
    }
  }

  if (!user || user.role !== "admin") {
    return null
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
      <div className="flex justify-between items-center mb-12">
        <div>
          <h1 className="text-4xl font-bold text-foreground">Administrar Actividades</h1>
          <p className="text-text-light mt-2">Crea, edita y elimina actividades deportivas</p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => {
              if (showForm && !editingActivity) {
                setShowForm(false)
              } else {
                setEditingActivity(null)
                setShowForm(!showForm)
              }
              if (!showForm) {
                setShowSessionForm(false) // Cerrar formulario de sesión si está abierto
              }
            }}
            className="px-6 py-3 bg-success text-white rounded-lg hover:bg-green-600 transition font-semibold"
          >
            {showForm && !editingActivity ? "Cancelar" : "+ Nueva Actividad"}
          </button>
          <button
            onClick={handleCreateSessionClick}
            className="px-6 py-3 bg-success text-white rounded-lg hover:bg-green-600 transition font-semibold"
          >
            {showSessionForm ? "Cancelar" : "+ Nueva Sesión"}
          </button>
        </div>
      </div>

      {showForm && (
        <div className="mb-12 p-6 bg-white rounded-lg shadow">
          <h2 className="text-2xl font-bold text-foreground mb-6">
            {editingActivity ? "Editar Actividad" : "Crear Nueva Actividad"}
          </h2>
          <ActivityForm 
            onSuccess={handleActivityCreated} 
            token={token || ""} 
            activity={editingActivity}
            onCancel={editingActivity ? handleCancelEdit : undefined}
          />
        </div>
      )}

      {showSessionForm && (
        <div className="mb-12 p-6 bg-white rounded-lg shadow">
          <h2 className="text-2xl font-bold text-foreground mb-6">Crear Nueva Sesión</h2>
          {!selectedActivity ? (
            <div className="space-y-4">
              <p className="text-text-light mb-4">Selecciona una actividad para crear una sesión:</p>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {activities.map((activity) => (
                  <button
                    key={activity.id}
                    onClick={() => setSelectedActivity(activity)}
                    className="p-4 border-2 border-border rounded-lg hover:border-primary hover:bg-surface transition text-left"
                  >
                    <h3 className="font-semibold text-foreground mb-1">{activity.nombre}</h3>
                    <p className="text-sm text-text-light">{activity.categoria}</p>
                    <p className="text-sm text-text-light">📍 {activity.ubicacion}</p>
                  </button>
                ))}
              </div>
            </div>
          ) : (
            <div>
              <button
                onClick={() => setSelectedActivity(null)}
                className="mb-4 text-primary hover:underline text-sm"
              >
                ← Volver a seleccionar actividad
              </button>
              <SessionForm
                onSuccess={handleSessionCreated}
                token={token || ""}
                activityId={selectedActivity.id}
                activityName={selectedActivity.nombre}
              />
            </div>
          )}
        </div>
      )}

      {error && <div className="p-4 bg-red-100 text-red-700 rounded-lg mb-6">{error}</div>}

      {isLoading ? (
        <div className="flex justify-center items-center py-12">
          <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
        </div>
      ) : (
        <ActivityList 
          activities={activities} 
          token={token || ""} 
          onActivityDeleted={handleActivityDeleted}
          onActivityEdit={handleActivityEdit}
        />
      )}
    </div>
  )
}
