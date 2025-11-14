"use client"

import { useAuth } from "@/context/auth-context"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { LogOut, Home, TicketX as Tickets, Settings } from "lucide-react"

export function Navbar() {
  const { logout, user } = useAuth()
  const pathname = usePathname()

  const isActive = (path: string) => pathname === path

  return (
    <nav className="fixed top-0 left-0 right-0 bg-primary text-white shadow-lg z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-20">
          <Link href="/home" className="flex items-center gap-2 hover:opacity-90 transition">
            <div className="w-10 h-10 bg-black rounded-lg flex items-center justify-center font-bold text-white text-lg">
              ⚽
            </div>
            <span className="text-xl font-bold hidden sm:inline">CanchaLibre</span>
          </Link>

          <div className="flex items-center gap-8">
            <Link
              href="/home"
              className={`flex items-center gap-2 transition ${
                isActive("/home") ? "text-white font-semibold" : "text-gray-200 hover:text-white"
              }`}
            >
              <Home className="w-5 h-5" />
              <span className="hidden sm:inline">Inicio</span>
            </Link>

            <Link
              href="/my-reservations"
              className={`flex items-center gap-2 transition ${
                isActive("/my-reservations") ? "text-white font-semibold" : "text-gray-200 hover:text-white"
              }`}
            >
              <Tickets className="w-5 h-5" />
              <span className="hidden sm:inline">Mis Reservas</span>
            </Link>

            {user?.role === "admin" && (
              <Link
                href="/manage-activities"
                className={`flex items-center gap-2 transition ${
                  isActive("/manage-activities") ? "text-white font-semibold" : "text-gray-200 hover:text-white"
                }`}
              >
                <Settings className="w-5 h-5" />
                <span className="hidden sm:inline">Administrar</span>
              </Link>
            )}

            <div className="flex items-center gap-4 pl-8 border-l border-blue-400">
              <span className="text-sm text-gray-200 hidden sm:inline">{user?.email}</span>
              <button
                onClick={logout}
                className="px-4 py-2 bg-orange-600 text-white rounded-lg hover:bg-orange-700 transition font-semibold flex items-center gap-2"
              >
                <LogOut className="w-4 h-4" />
                <span className="hidden sm:inline">Salir</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}
