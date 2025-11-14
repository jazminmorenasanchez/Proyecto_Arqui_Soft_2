"use client"

import type React from "react"

import { useState, useEffect } from "react"
import { Toast } from "./toast"
import { activitiesAPI, type Activity as ActivityType } from "@/lib/api"

interface ActivityFormProps {
  onSuccess: () => void
  token: string
  activity?: ActivityType | null
  onCancel?: () => void
}

interface FormData {
  nombre: string
  categoria: string
  instructor: string
  ubicacion: string
  precioBase: number | string
}

export default function ActivityForm({ onSuccess, token, activity, onCancel }: ActivityFormProps) {
  const isEditMode = !!activity
  
  const [formData, setFormData] = useState<FormData>({
    nombre: activity?.nombre || "",
    categoria: activity?.categoria || "",
    instructor: activity?.instructor || "",
    ubicacion: activity?.ubicacion || "",
    precioBase: activity?.precioBase || "",
  })
  const [isLoading, setIsLoading] = useState(false)
  const [showToast, setShowToast] = useState(false)
  const [toastMessage, setToastMessage] = useState("")
  const [toastType, setToastType] = useState<"success" | "error">("success")

  // Actualizar formulario cuando cambie la actividad
  useEffect(() => {
    if (activity) {
      setFormData({
        nombre: activity.nombre || "",
        categoria: activity.categoria || "",
        instructor: activity.instructor || "",
        ubicacion: activity.ubicacion || "",
        precioBase: activity.precioBase || "",
      })
    } else {
      setFormData({
        nombre: "",
        categoria: "",
        instructor: "",
        ubicacion: "",
        precioBase: "",
      })
    }
  }, [activity])

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: name === "precioBase" ? (value === "" ? "" : Number(value)) : value,
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.nombre || !formData.categoria || !formData.ubicacion || !formData.precioBase) {
      setToastMessage("Por favor, completa los campos requeridos")
      setToastType("error")
      setShowToast(true)
      return
    }

    try {
      setIsLoading(true)
      const payload = {
        nombre: formData.nombre,
        categoria: formData.categoria,
        ubicacion: formData.ubicacion,
        instructor: formData.instructor || undefined,
        precioBase: Number(formData.precioBase),
      }

      if (isEditMode && activity) {
        await activitiesAPI.update(activity.id, payload, token)
        setToastMessage("Actividad actualizada exitosamente")
      } else {
        await activitiesAPI.create(payload, token)
        setToastMessage("Actividad creada exitosamente")
        setFormData({
          nombre: "",
          categoria: "",
          instructor: "",
          ubicacion: "",
          precioBase: "",
        })
      }

      setToastType("success")
      setShowToast(true)

      setTimeout(() => {
        onSuccess()
      }, 1000)
    } catch (error) {
      setToastMessage(error instanceof Error ? error.message : "An error occurred")
      setToastType("error")
      setShowToast(true)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Nombre de la Actividad</label>
            <input
              type="text"
              name="nombre"
              value={formData.nombre}
              onChange={handleChange}
              placeholder="Ej: Fútbol 5"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Categoría</label>
            <input
              type="text"
              name="categoria"
              value={formData.categoria}
              onChange={handleChange}
              placeholder="Ej: Fútbol, Básquet, Tenis..."
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Ubicación</label>
            <input
              type="text"
              name="ubicacion"
              value={formData.ubicacion}
              onChange={handleChange}
              placeholder="Ej: Cancha Centro"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Instructor (Opcional)</label>
            <input
              type="text"
              name="instructor"
              value={formData.instructor}
              onChange={handleChange}
              placeholder="Ej: Juan Pérez"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Precio Base por Sesión</label>
            <input
              type="number"
              name="precioBase"
              value={formData.precioBase}
              onChange={handleChange}
              placeholder="Ej: 100"
              step="1"
              min="0"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>
        </div>

        <div className="flex gap-3">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={isLoading}
              className="flex-1 px-6 py-3 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Cancelar
            </button>
          )}
          <button
            type="submit"
            disabled={isLoading}
            className={`${onCancel ? 'flex-1' : 'w-full'} px-6 py-3 bg-success text-white rounded-lg hover:bg-green-600 transition font-semibold disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isLoading 
              ? (isEditMode ? "Actualizando..." : "Creando...") 
              : (isEditMode ? "Actualizar Actividad" : "Crear Actividad")
            }
          </button>
        </div>
      </form>

      {showToast && <Toast message={toastMessage} type={toastType} onClose={() => setShowToast(false)} />}
    </>
  )
}
