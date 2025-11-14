"use client"

import { useState } from "react"
import { SessionList } from "./session-list"
import { useAuth } from "@/context/auth-context"
import { ChevronDown } from "lucide-react"
import { sessionsAPI, type Activity, type Session } from "@/lib/api"

interface ActivityCardProps {
  activity: Activity
  onToast: (toast: { message: string; type: "error" | "success" }) => void
}

export function ActivityCard({ activity, onToast }: ActivityCardProps) {
  const [expanded, setExpanded] = useState(false)
  const [sessions, setSessions] = useState<Session[]>([])
  const [isLoading, setIsLoading] = useState(false)
  const { token } = useAuth()

  const handleExpandClick = async () => {
    if (expanded) {
      setExpanded(false)
      return
    }

    setIsLoading(true)
    try {
      const data = await sessionsAPI.getByActivity(activity.id, token)
      setSessions(data || [])
      setExpanded(true)
    } catch (error) {
      onToast({ message: "Error al cargar sesiones", type: "error" })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="bg-white border border-border rounded-lg overflow-hidden shadow-sm hover:shadow-lg transition-shadow">
      <div className="p-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <h3 className="text-xl font-bold text-foreground mb-1">{activity.nombre}</h3>
            <p className="text-sm text-text-light mb-3 flex items-center gap-1">📍 {activity.ubicacion}</p>
            <div className="flex gap-2 flex-wrap">
              <span className="inline-block px-3 py-1 bg-primary text-white text-xs rounded-full font-medium">
                {activity.categoria}
              </span>
              {activity.instructor && (
                <span className="inline-block px-3 py-1 bg-surface text-foreground text-xs rounded-full">
                  👨‍🏫 {activity.instructor}
                </span>
              )}
            </div>
          </div>
          <div className="text-right ml-4">
            <p className="text-2xl font-bold text-success">${activity.precioBase}</p>
            <p className="text-sm text-text-light">★ {activity.rating.toFixed(1)}</p>
          </div>
        </div>

        <button
          onClick={handleExpandClick}
          disabled={isLoading}
          className="w-full py-2.5 bg-primary text-white font-semibold rounded-lg hover:bg-blue-900 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
        >
          {isLoading ? (
            "Cargando..."
          ) : expanded ? (
            <>
              Ocultar Sesiones <ChevronDown className="w-4 h-4 rotate-180" />
            </>
          ) : (
            <>
              Ver Sesiones <ChevronDown className="w-4 h-4" />
            </>
          )}
        </button>
      </div>

      {expanded && <SessionList sessions={sessions} activityId={activity.id} onToast={onToast} />}
    </div>
  )
}
