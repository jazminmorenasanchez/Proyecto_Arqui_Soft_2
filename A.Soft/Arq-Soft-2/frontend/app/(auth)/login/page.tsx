"use client"

import { useState } from "react"
import Link from "next/link"
import { LoginForm } from "@/components/login-form"
import { Toast } from "@/components/toast"

export default function LoginPage() {
  const [toast, setToast] = useState<{ message: string; type: "error" | "success" } | null>(null)

  return (
    <div className="w-full max-w-md">
      <div className="text-center mb-8">
        <div className="text-5xl mb-3">⚽</div>
        <h1 className="text-4xl font-bold text-primary mb-2">CanchaLibre</h1>
        <p className="text-text-light">Reserva actividades deportivas fácilmente</p>
      </div>

      <LoginForm onToast={setToast} />

      <p className="text-center text-sm text-text-light mt-6">
        ¿No tienes cuenta?{" "}
        <Link href="/register" className="text-primary font-semibold hover:underline">
          Registrate aquí
        </Link>
      </p>

      {toast && <Toast message={toast.message} type={toast.type} onClose={() => setToast(null)} />}
    </div>
  )
}
