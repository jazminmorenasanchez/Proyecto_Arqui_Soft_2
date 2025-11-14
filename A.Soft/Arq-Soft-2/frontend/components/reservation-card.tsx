"use client"

import { useState } from "react"
import { Spinner } from "@/components/ui/spinner"
import { Calendar, Clock, MapPin } from "lucide-react"

interface Enrollment {
  id: number
  activityId: number
  sessionId: number
  userId: string
  precioFinal: number
  estado: string
  createdAt: string
  activity?: any
  session?: any
}

interface ReservationCardProps {
  enrollment: Enrollment
  onCancel: () => Promise<void>
  onToast: (toast: { message: string; type: "error" | "success" }) => void
}

export function ReservationCard({ enrollment, onCancel, onToast }: ReservationCardProps) {
  const [isCancelling, setIsCancelling] = useState(false)

  const handleCancel = async () => {
    if (!confirm("¿Estás seguro de que deseas cancelar esta reserva?")) return

    setIsCancelling(true)
    try {
      await onCancel()
    } catch (error) {
      // El error ya se maneja en el componente padre
      console.error("Error en handleCancel:", error)
    } finally {
      setIsCancelling(false)
    }
  }

  const getStatusColor = (estado: string) => {
    switch (estado?.toLowerCase()) {
      case "confirmada":
        return "bg-success text-white"
      case "pendiente":
        return "bg-accent text-white"
      case "cancelada":
        return "bg-red-500 text-white"
      default:
        return "bg-text-light text-white"
    }
  }

  const getStatusLabel = (estado: string) => {
    return estado?.charAt(0).toUpperCase() + estado?.slice(1).toLowerCase() || "Estado"
  }

  return (
    <div className="bg-white border border-border rounded-lg overflow-hidden shadow-sm hover:shadow-lg transition-shadow">
      <div className="p-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h3 className="text-xl font-bold text-foreground mb-1">{enrollment.activity?.nombre || "Actividad"}</h3>
            <p className="text-sm text-text-light flex items-center gap-1 mb-2">
              <MapPin className="w-4 h-4" />
              {enrollment.activity?.ubicacion || "Ubicación no disponible"}
            </p>
          </div>
          <span
            className={`${getStatusColor(enrollment.estado)} px-4 py-1 rounded-full text-xs font-semibold whitespace-nowrap ml-4`}
          >
            {getStatusLabel(enrollment.estado)}
          </span>
        </div>

        {enrollment.session && (
          <div className="bg-surface p-4 rounded-lg mb-4 space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <Calendar className="w-4 h-4 text-primary" />
              <p className="font-semibold text-foreground">
                {new Date(enrollment.session.fecha).toLocaleDateString("es-AR", {
                  weekday: "long",
                  year: "numeric",
                  month: "long",
                  day: "numeric",
                })}
              </p>
            </div>
            <div className="flex items-center gap-2 text-sm text-text-light ml-6">
              <Clock className="w-4 h-4" />
              <span>
                {enrollment.session.inicio} - {enrollment.session.fin}
              </span>
            </div>
          </div>
        )}

        <div className="flex items-center justify-between mb-4 pb-4 border-b border-border">
          <p className="text-lg font-bold text-success">
            ${enrollment.activity?.precioBase ?? enrollment.precioFinal}
          </p>
          <p className="text-xs text-text-light">
            Inscrito el {new Date(enrollment.createdAt).toLocaleDateString("es-AR")}
          </p>
        </div>

        <button
          onClick={handleCancel}
          disabled={isCancelling || enrollment.estado?.toLowerCase() === "cancelada"}
          className="w-full py-2.5 bg-red-500 text-white rounded-lg hover:bg-red-600 transition disabled:opacity-50 disabled:cursor-not-allowed font-semibold flex items-center justify-center gap-2"
        >
          {isCancelling ? (
            <>
              <Spinner className="w-4 h-4" />
              Cancelando...
            </>
          ) : enrollment.estado?.toLowerCase() === "cancelada" ? (
            "Reserva Cancelada"
          ) : (
            "Cancelar Reserva"
          )}
        </button>
      </div>
    </div>
  )
}
