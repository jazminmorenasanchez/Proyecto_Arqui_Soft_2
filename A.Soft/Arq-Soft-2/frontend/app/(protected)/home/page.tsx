"use client"

import { useEffect, useState } from "react"
import { useAuth } from "@/context/auth-context"
import { ActivityCard } from "@/components/activity-card"
import { SearchBar } from "@/components/search-bar"
import { Toast } from "@/components/toast"
import { Spinner } from "@/components/ui/spinner"
import { activitiesAPI, searchAPI, type Activity, type SearchResult } from "@/lib/api"

export default function HomePage() {
  const [activities, setActivities] = useState<Activity[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSearching, setIsSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<SearchResult | null>(null)
  const [toast, setToast] = useState<{ message: string; type: "error" | "success" } | null>(null)
  const { token } = useAuth()

  useEffect(() => {
    if (token && !searchResults) {
      fetchActivities()
    }
  }, [token])

  const fetchActivities = async () => {
    try {
      setIsLoading(true)
      const data = await activitiesAPI.list(0, 20, token)
      setActivities(data.activities || [])
    } catch (error) {
      setToast({ message: "Error al cargar actividades", type: "error" })
    } finally {
      setIsLoading(false)
    }
  }

  const handleSearchResults = async (results: SearchResult | null) => {
    if (!results) {
      // Si hay error, limpiar búsqueda y volver a la lista normal
      setSearchResults(null)
      fetchActivities()
      return
    }

    console.log("Resultados de búsqueda:", results)
    setSearchResults(results)
    
    // Si no hay resultados, limpiar actividades inmediatamente
    if (results.total === 0 || results.docs.length === 0) {
      console.log("No hay resultados, limpiando actividades")
      setActivities([])
      setIsSearching(false)
      return
    }

    setIsSearching(true)
    try {
      // Obtener actividades únicas de los resultados de búsqueda
      const uniqueActivityIds = [...new Set(results.docs.map((doc) => doc.activity_id))]
      console.log("Activity IDs únicos:", uniqueActivityIds)
      
      const activityPromises = uniqueActivityIds.map((id) => {
        const numId = Number(id)
        console.log(`Buscando actividad con ID: ${id} (convertido a: ${numId})`)
        return activitiesAPI.getById(numId, token).catch((err) => {
          console.error(`Error al obtener actividad ${id}:`, err)
          return null
        })
      })
      const fetchedActivities = await Promise.all(activityPromises)
      const validActivities = fetchedActivities.filter((a): a is Activity => a !== null)
      console.log("Actividades obtenidas:", validActivities.length)
      setActivities(validActivities)
    } catch (error) {
      console.error("Error al cargar actividades de búsqueda:", error)
      setToast({ message: "Error al cargar actividades de búsqueda", type: "error" })
      setActivities([])
    } finally {
      setIsSearching(false)
    }
  }

  const handleClearSearch = () => {
    setSearchResults(null)
    fetchActivities()
  }

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Spinner className="w-12 h-12 mx-auto mb-4 text-primary" />
          <p className="text-text-light">Cargando actividades...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="mb-8">
        <h1 className="text-4xl font-bold text-foreground mb-2">Actividades Deportivas</h1>
        <p className="text-text-light text-lg">Descubre y reserva tus actividades favoritas</p>
      </div>

      <SearchBar onSearchResults={handleSearchResults} onClear={handleClearSearch} />

      {(isLoading || isSearching) ? (
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <Spinner className="w-12 h-12 mx-auto mb-4 text-primary" />
            <p className="text-text-light">
              {isSearching ? "Buscando actividades..." : "Cargando actividades..."}
            </p>
          </div>
        </div>
      ) : searchResults ? (
        <>
          <div className="mb-4 text-text-light">
            {searchResults.total > 0 ? (
              <p>
                Se encontraron {searchResults.total} resultado{searchResults.total !== 1 ? "s" : ""}
              </p>
            ) : (
              <p>No se encontraron resultados</p>
            )}
          </div>
          {searchResults.total === 0 || activities.length === 0 ? (
            <div className="text-center py-16 bg-surface rounded-lg border border-border">
              <p className="text-text-light text-lg mb-4">No se encontraron actividades</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {activities.map((activity) => (
                <ActivityCard key={activity.id} activity={activity} onToast={setToast} />
              ))}
            </div>
          )}
        </>
      ) : activities.length === 0 ? (
        <div className="text-center py-16 bg-surface rounded-lg border border-border">
          <p className="text-text-light text-lg mb-4">No hay actividades disponibles en este momento</p>
          <button
            onClick={fetchActivities}
            className="px-6 py-2 bg-primary text-white rounded-lg hover:bg-blue-900 transition font-semibold"
          >
            Reintentar
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {activities.map((activity) => (
            <ActivityCard key={activity.id} activity={activity} onToast={setToast} />
          ))}
        </div>
      )}

      {toast && <Toast message={toast.message} type={toast.type} onClose={() => setToast(null)} />}
    </div>
  )
}
