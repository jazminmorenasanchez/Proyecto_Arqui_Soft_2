"use client"

import type React from "react"

import { useState } from "react"
import { Toast } from "./toast"
import { sessionsAPI } from "@/lib/api"

interface SessionFormProps {
  onSuccess: () => void
  token: string
  activityId: number
  activityName: string
}

interface FormData {
  fecha: string // YYYY-MM-DD
  inicio: string // HH:mm
  fin: string // HH:mm
  capacidad: number | string
}

export default function SessionForm({ onSuccess, token, activityId, activityName }: SessionFormProps) {
  const [formData, setFormData] = useState<FormData>({
    fecha: "",
    inicio: "",
    fin: "",
    capacidad: "",
  })
  const [isLoading, setIsLoading] = useState(false)
  const [showToast, setShowToast] = useState(false)
  const [toastMessage, setToastMessage] = useState("")
  const [toastType, setToastType] = useState<"success" | "error">("success")

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFormData((prev) => ({
      ...prev,
      [name]: name === "capacidad" ? (value === "" ? "" : Number(value)) : value,
    }))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!formData.fecha || !formData.inicio || !formData.fin || !formData.capacidad) {
      setToastMessage("Por favor, completa todos los campos requeridos")
      setToastType("error")
      setShowToast(true)
      return
    }

    // Validar que la hora de fin sea mayor que la de inicio
    if (formData.inicio >= formData.fin) {
      setToastMessage("La hora de fin debe ser mayor que la hora de inicio")
      setToastType("error")
      setShowToast(true)
      return
    }

    try {
      setIsLoading(true)
      const payload = {
        fecha: formData.fecha,
        inicio: formData.inicio,
        fin: formData.fin,
        capacidad: Number(formData.capacidad),
      }

      await sessionsAPI.createForActivity(activityId, payload, token)

      setToastMessage("Sesión creada exitosamente")
      setToastType("success")
      setShowToast(true)

      setFormData({
        fecha: "",
        inicio: "",
        fin: "",
        capacidad: "",
      })

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

  // Obtener fecha mínima (hoy)
  const today = new Date().toISOString().split("T")[0]

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="mb-4 p-3 bg-surface rounded-lg">
          <p className="text-sm text-text-light">Actividad:</p>
          <p className="font-semibold text-foreground">{activityName}</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Fecha *</label>
            <input
              type="date"
              name="fecha"
              value={formData.fecha}
              onChange={handleChange}
              min={today}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Capacidad *</label>
            <input
              type="number"
              name="capacidad"
              value={formData.capacidad}
              onChange={handleChange}
              placeholder="Ej: 20"
              min="1"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Hora de Inicio *</label>
            <input
              type="time"
              name="inicio"
              value={formData.inicio}
              onChange={handleChange}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-2">Hora de Fin *</label>
            <input
              type="time"
              name="fin"
              value={formData.fin}
              onChange={handleChange}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent"
              required
            />
          </div>
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="w-full px-6 py-3 bg-success text-white rounded-lg hover:bg-green-600 transition font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? "Creando..." : "Crear Sesión"}
        </button>
      </form>

      {showToast && <Toast message={toastMessage} type={toastType} onClose={() => setShowToast(false)} />}
    </>
  )
}

