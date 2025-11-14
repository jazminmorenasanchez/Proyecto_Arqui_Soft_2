import type React from "react"
import { Inter, Merriweather } from "next/font/google"
import { AuthProvider } from "@/context/auth-context"
import "./globals.css"

const inter = Inter({ subsets: ["latin"] })
const merriweather = Merriweather({
  subsets: ["latin"],
  weight: ["400", "700"],
})

export const metadata = {
  title: "CanchaLibre - Reserva Actividades Deportivas",
  description: "Plataforma para gestión y reserva de actividades deportivas",
    generator: 'v0.app'
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="es">
      <body className={inter.className}>
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  )
}
