"use client"

import { useState } from "react"
import { useAuth } from "@/context/auth-context"
import { Spinner } from "@/components/ui/spinner"
import { Calendar, Clock } from "lucide-react"
import { enrollmentsAPI, type Session } from "@/lib/api"

interface SessionListProps {
  sessions: Session[]
  activityId: number
  onToast: (toast: { message: string; type: "error" | "success" }) => void
}

export function SessionList({ sessions, activityId, onToast }: SessionListProps) {
  const [enrolling, setEnrolling] = useState<number | null>(null)
  const { token } = useAuth()

  const handleEnroll = async (sessionId: number) => {
    if (!token) {
      onToast({ message: "Debes iniciar sesión para inscribirte", type: "error" })
      return
    }

    setEnrolling(sessionId)
    try {
      await enrollmentsAPI.enroll(sessionId, token)
      onToast({ message: "¡Inscrito exitosamente!", type: "success" })
    } catch (error) {
      onToast({
        message: error instanceof Error ? error.message : "Error al inscribirse",
        type: "error",
      })
    } finally {
      setEnrolling(null)
    }
  }

  return (
    <div className="border-t border-border p-6 bg-surface">
      <h4 className="font-semibold mb-4 text-foreground text-lg">Sesiones Disponibles</h4>
      <div className="space-y-3">
        {sessions.length === 0 ? (
          <p className="text-text-light text-sm text-center py-4">Sin sesiones disponibles</p>
        ) : (
          sessions.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between bg-white p-4 rounded-lg border border-border hover:border-primary transition"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <Calendar className="w-4 h-4 text-primary" />
                  <p className="font-semibold text-foreground">
                    {new Date(session.fecha).toLocaleDateString("es-AR", {
                      weekday: "short",
                      month: "short",
                      day: "numeric",
                    })}
                  </p>
                </div>
                <div className="flex items-center gap-2 text-sm text-text-light ml-6">
                  <Clock className="w-4 h-4" />
                  <span>
                    {session.inicio} - {session.fin}
                  </span>
                </div>
                <p className="text-xs text-text-light mt-1 ml-6">Capacidad: {session.capacidad} personas</p>
              </div>
              <button
                onClick={() => handleEnroll(session.id)}
                disabled={enrolling === session.id}
                className="px-6 py-2 bg-success text-white rounded-lg hover:bg-green-600 transition disabled:opacity-50 ml-4 font-semibold flex items-center gap-2 whitespace-nowrap"
              >
                {enrolling === session.id ? (
                  <>
                    <Spinner className="w-4 h-4" />
                    Inscribiendo...
                  </>
                ) : (
                  "Inscribirse"
                )}
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
