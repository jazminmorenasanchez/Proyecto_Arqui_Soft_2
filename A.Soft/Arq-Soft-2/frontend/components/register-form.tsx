"use client"

import type React from "react"
import { useState } from "react"
import { useAuth } from "@/context/auth-context"
import { Spinner } from "@/components/ui/spinner"
import { Eye, EyeOff } from "lucide-react"

interface RegisterFormProps {
  onToast: (toast: { message: string; type: "error" | "success" }) => void
}

export function RegisterForm({ onToast }: RegisterFormProps) {
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const { register } = useAuth()

  const validateForm = () => {
    if (!username || !email || !password || !confirmPassword) {
      onToast({ message: "Por favor completa todos los campos", type: "error" })
      return false
    }
    if (username.length < 3) {
      onToast({ message: "El usuario debe tener al menos 3 caracteres", type: "error" })
      return false
    }
    if (username.length > 50) {
      onToast({ message: "El usuario no puede exceder 50 caracteres", type: "error" })
      return false
    }
    if (password.length < 6) {
      onToast({ message: "La contraseña debe tener al menos 6 caracteres", type: "error" })
      return false
    }
    if (password !== confirmPassword) {
      onToast({ message: "Las contraseñas no coinciden", type: "error" })
      return false
    }
    return true
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!validateForm()) return

    setIsLoading(true)

    try {
      await register(username, email, password)
      onToast({ message: "¡Cuenta creada exitosamente!", type: "success" })
    } catch (error) {
      onToast({
        message: error instanceof Error ? error.message : "Error al registrarse",
        type: "error",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-semibold mb-2 text-foreground">Usuario</label>
        <input
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="tu_usuario"
          required
          className="w-full px-4 py-2 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary bg-background transition"
        />
        <p className="text-xs text-text-light mt-1">3-50 caracteres</p>
      </div>

      <div>
        <label className="block text-sm font-semibold mb-2 text-foreground">Email</label>
        <input
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="tu@email.com"
          required
          className="w-full px-4 py-2 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary bg-background transition"
        />
      </div>

      <div>
        <label className="block text-sm font-semibold mb-2 text-foreground">Contraseña</label>
        <div className="relative">
          <input
            type={showPassword ? "text" : "password"}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            required
            className="w-full px-4 py-2 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary bg-background transition pr-10"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-light hover:text-foreground transition"
          >
            {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
          </button>
        </div>
        <p className="text-xs text-text-light mt-1">Mínimo 6 caracteres</p>
      </div>

      <div>
        <label className="block text-sm font-semibold mb-2 text-foreground">Confirmar Contraseña</label>
        <div className="relative">
          <input
            type={showConfirmPassword ? "text" : "password"}
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="••••••••"
            required
            className="w-full px-4 py-2 border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary bg-background transition pr-10"
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-text-light hover:text-foreground transition"
          >
            {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
          </button>
        </div>
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full py-2.5 bg-success text-white font-semibold rounded-lg hover:bg-green-600 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {isLoading ? (
          <>
            <Spinner className="w-4 h-4" />
            Registrando...
          </>
        ) : (
          "Crear Cuenta"
        )}
      </button>
    </form>
  )
}
