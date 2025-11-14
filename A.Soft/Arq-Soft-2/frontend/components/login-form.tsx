"use client"

import type React from "react"
import { useState } from "react"
import { useAuth } from "@/context/auth-context"
import { Spinner } from "@/components/ui/spinner"
import { Eye, EyeOff } from "lucide-react"

interface LoginFormProps {
  onToast: (toast: { message: string; type: "error" | "success" }) => void
}

export function LoginForm({ onToast }: LoginFormProps) {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const { login } = useAuth()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!email || !password) {
      onToast({ message: "Por favor completa todos los campos", type: "error" })
      return
    }

    setIsLoading(true)

    try {
      await login(email, password)
      onToast({ message: "¡Bienvenido!", type: "success" })
    } catch (error) {
      onToast({
        message: error instanceof Error ? error.message : "Email o contraseña incorrectos",
        type: "error",
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
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
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full py-2.5 bg-primary text-white font-semibold rounded-lg hover:bg-blue-900 transition disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        {isLoading ? (
          <>
            <Spinner className="w-4 h-4" />
            Iniciando sesión...
          </>
        ) : (
          "Iniciar Sesión"
        )}
      </button>
    </form>
  )
}
