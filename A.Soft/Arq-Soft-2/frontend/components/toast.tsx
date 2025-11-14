"use client"

import { useEffect } from "react"

interface ToastProps {
  message: string
  type: "success" | "error"
  onClose: () => void
}

export function Toast({ message, type, onClose }: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, 3000)
    return () => clearTimeout(timer)
  }, [onClose])

  const bgColor = type === "success" ? "bg-success" : "bg-red-500"

  return (
    <div
      className={`fixed bottom-4 right-4 ${bgColor} text-white px-6 py-3 rounded-lg shadow-lg animate-in fade-in slide-in-from-bottom-4 duration-300`}
    >
      {message}
    </div>
  )
}
