"use client"

import { useState } from "react"
import { Search, X } from "lucide-react"
import { useAuth } from "@/context/auth-context"
import { searchAPI, type SearchResult } from "@/lib/api"

interface SearchBarProps {
  onSearchResults: (results: SearchResult | null) => void
  onClear: () => void
}

export function SearchBar({ onSearchResults, onClear }: SearchBarProps) {
  const [query, setQuery] = useState("")
  const [isSearching, setIsSearching] = useState(false)
  const { token } = useAuth()

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!query.trim()) {
      onClear()
      return
    }

    setIsSearching(true)
    try {
      console.log("Buscando:", query.trim())
      const results = await searchAPI.search({ query: query.trim(), size: 20 }, token)
      console.log("Respuesta de búsqueda:", results)
      // Siempre pasar los resultados, incluso si están vacíos
      onSearchResults(results)
    } catch (error) {
      console.error("Error al buscar:", error)
      // En caso de error, pasar null para que se limpie la búsqueda
      onSearchResults(null)
    } finally {
      setIsSearching(false)
    }
  }

  const handleClear = () => {
    setQuery("")
    onClear()
  }

  return (
    <form onSubmit={handleSearch} className="mb-8">
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
          <Search className="h-5 w-5 text-text-light" />
        </div>
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Buscar actividades por nombre..."
          className="w-full pl-12 pr-12 py-3 border border-border rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent text-foreground bg-white"
        />
        {query && (
          <button
            type="button"
            onClick={handleClear}
            className="absolute inset-y-0 right-0 pr-4 flex items-center text-text-light hover:text-foreground transition"
          >
            <X className="h-5 w-5" />
          </button>
        )}
      </div>
      {isSearching && (
        <div className="mt-2 text-sm text-text-light text-center">Buscando...</div>
      )}
    </form>
  )
}

