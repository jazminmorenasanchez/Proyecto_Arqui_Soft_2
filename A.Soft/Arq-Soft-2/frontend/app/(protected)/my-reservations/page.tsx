"use client"

import { useEffect, useState } from "react"
import { useAuth } from "@/context/auth-context"
import { ReservationCard } from "@/components/reservation-card"
import { Toast } from "@/components/toast"
import { Spinner } from "@/components/ui/spinner"
import { enrollmentsAPI, activitiesAPI, sessionsAPI, type Enrollment, type Activity, type Session } from "@/lib/api"

interface EnrichedEnrollment extends Enrollment {
  activity?: Activity
  session?: Session
}

export default function MyReservationsPage() {
  const [enrollments, setEnrollments] = useState<EnrichedEnrollment[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [toast, setToast] = useState<{ message: string; type: "error" | "success" } | null>(null)
  const { token, user } = useAuth()

  useEffect(() => {
    if (user?.id) {
      fetchEnrollments()
    }
  }, [token, user?.id])

  const fetchEnrollments = async () => {
    if (!user?.id || !token) return

    try {
      const data = await enrollmentsAPI.getByUser(user.id, token)

      const enriched = await Promise.all(
        (data || []).map(async (enrollment: Enrollment) => {
          try {
            const [activity, sessions] = await Promise.all([
              activitiesAPI.getById(enrollment.activityId, token).catch(() => null),
              sessionsAPI.getByActivity(enrollment.activityId, token).catch(() => []),
            ])

            const session = sessions?.find((s) => s.id === enrollment.sessionId)

            return {
              ...enrollment,
              activity: activity || undefined,
              session: session || undefined,
            }
          } catch (e) {
            return enrollment
          }
        }),
      )

      setEnrollments(enriched)
    } catch (error) {
      setToast({ message: "Error al cargar reservas", type: "error" })
    } finally {
      setIsLoading(false)
    }
  }

  const handleCancel = async (enrollmentId: number) => {
    if (!token) {
      setToast({ message: "Debes iniciar sesión para cancelar reservas", type: "error" })
      return
    }

    try {
      await enrollmentsAPI.cancel(enrollmentId, token)
      setToast({ message: "Reserva cancelada exitosamente", type: "success" })
      // Recargar las reservas después de cancelar
      await fetchEnrollments()
    } catch (error) {
      console.error("Error al cancelar reserva:", error)
      setToast({
        message: error instanceof Error ? error.message : "Error al cancelar reserva. Intenta nuevamente.",
        type: "error",
      })
    }
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Spinner className="w-12 h-12 mx-auto mb-4 text-primary" />
          <p className="text-text-light">Cargando reservas...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-foreground mb-2">Mis Reservas</h1>
        <p className="text-text-light text-lg">Gestiona tus inscripciones y actividades</p>
      </div>

      {enrollments.length === 0 ? (
        <div className="text-center py-16 bg-surface rounded-lg border border-border">
          <p className="text-text-light text-lg mb-4">No tienes reservas aún</p>
          <a
            href="/home"
            className="inline-block px-6 py-2 bg-primary text-white rounded-lg hover:bg-blue-900 transition font-semibold"
          >
            Explorar actividades →
          </a>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {enrollments.map((enrollment) => (
            <ReservationCard
              key={enrollment.id}
              enrollment={enrollment}
              onCancel={async () => await handleCancel(enrollment.id)}
              onToast={setToast}
            />
          ))}
        </div>
      )}

      {toast && <Toast message={toast.message} type={toast.type} onClose={() => setToast(null)} />}
    </div>
  )
}
