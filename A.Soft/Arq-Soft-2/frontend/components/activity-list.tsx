"use client"

import { useState } from "react"
import { Trash2, Edit } from "lucide-react"
import { Toast } from "./toast"
import { activitiesAPI, type Activity } from "@/lib/api"

interface ActivityListProps {
  activities: Activity[]
  token: string
  onActivityDeleted: () => void
  onActivityEdit?: (activity: Activity) => void
}

export default function ActivityList({ activities, token, onActivityDeleted, onActivityEdit }: ActivityListProps) {
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [showToast, setShowToast] = useState(false)
  const [toastMessage, setToastMessage] = useState("")
  const [toastType, setToastType] = useState<"success" | "error">("success")

  const handleDelete = async (id: number) => {
    if (!window.confirm("¿Estás seguro de que deseas eliminar esta actividad?")) {
      return
    }

    try {
      setDeletingId(id)
      await activitiesAPI.delete(id, token)

      setToastMessage("Actividad eliminada exitosamente")
      setToastType("success")
      setShowToast(true)

      setTimeout(() => {
        onActivityDeleted()
      }, 500)
    } catch (error) {
      setToastMessage(error instanceof Error ? error.message : "An error occurred")
      setToastType("error")
      setShowToast(true)
    } finally {
      setDeletingId(null)
    }
  }

  if (activities.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-text-light text-lg">No hay actividades creadas aún</p>
      </div>
    )
  }

  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {activities.map((activity) => (
          <div key={activity.id} className="bg-white rounded-lg shadow hover:shadow-lg transition p-6">
            <div className="flex justify-between items-start mb-4">
              <h3 className="text-xl font-bold text-foreground flex-1">{activity.nombre}</h3>
              <div className="flex gap-2">
                {onActivityEdit && (
                  <button
                    onClick={() => onActivityEdit(activity)}
                    className="p-2 text-primary hover:bg-primary/10 rounded-lg transition"
                    title="Editar actividad"
                  >
                    <Edit className="w-5 h-5" />
                  </button>
                )}
                <button
                  onClick={() => handleDelete(activity.id)}
                  disabled={deletingId === activity.id}
                  className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition disabled:opacity-50"
                  title="Eliminar actividad"
                >
                  <Trash2 className="w-5 h-5" />
                </button>
              </div>
            </div>

            <div className="space-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-text-light">Categoría:</span>
                <span className="font-medium text-foreground">{activity.categoria}</span>
              </div>

              <div className="flex justify-between">
                <span className="text-text-light">Ubicación:</span>
                <span className="font-medium text-foreground">{activity.ubicacion}</span>
              </div>

              {activity.instructor && (
                <div className="flex justify-between">
                  <span className="text-text-light">Instructor:</span>
                  <span className="font-medium text-foreground">{activity.instructor}</span>
                </div>
              )}

              <div className="flex justify-between pt-2 border-t border-gray-200">
                <span className="text-text-light">Precio Base:</span>
                <span className="font-bold text-success">${activity.precioBase}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {showToast && <Toast message={toastMessage} type={toastType} onClose={() => setShowToast(false)} />}
    </>
  )
}
