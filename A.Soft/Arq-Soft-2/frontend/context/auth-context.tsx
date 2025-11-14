"use client"

import { createContext, useContext, useEffect, useState, type ReactNode } from "react"
import { useRouter } from "next/navigation"
import { usersAPI } from "@/lib/api"

interface User {
  id: string | number
  username: string
  email: string
  role: "user" | "admin"
}

interface AuthContextType {
  user: User | null
  token: string | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  register: (username: string, email: string, password: string) => Promise<void>
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const router = useRouter()

  // Initialize from localStorage
  useEffect(() => {
    if (typeof window === "undefined") {
      setIsLoading(false)
      return
    }

    const storedToken = localStorage.getItem("auth_token")
    const storedUser = localStorage.getItem("auth_user")

    if (storedToken && storedUser) {
      try {
        setToken(storedToken)
        setUser(JSON.parse(storedUser))
      } catch (error) {
        localStorage.removeItem("auth_token")
        localStorage.removeItem("auth_user")
      }
    }
    setIsLoading(false)
  }, [])

  const login = async (email: string, password: string) => {
    try {
      const data = await usersAPI.login(email, password)

      const userId = String(data.userId)
      const userData: User = {
        id: userId,
        username: email.split("@")[0],
        email,
        role: (data.role || "user") as "user" | "admin",
      }

      if (typeof window !== "undefined") {
      localStorage.setItem("auth_token", data.token)
      localStorage.setItem("auth_user", JSON.stringify(userData))
      }

      setToken(data.token)
      setUser(userData)

      router.push("/home")
    } catch (error) {
      throw error
    }
  }

  const register = async (username: string, email: string, password: string) => {
    try {
      await usersAPI.register(username, email, password)

      // Auto-login after registration
      await login(email, password)
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    if (typeof window !== "undefined") {
    localStorage.removeItem("auth_token")
    localStorage.removeItem("auth_user")
    }
    setToken(null)
    setUser(null)
    router.push("/login")
  }

  return (
    <AuthContext.Provider
      value={{
        user,
        token,
        isLoading,
        login,
        register,
        logout,
        isAuthenticated: !!token,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error("useAuth must be used within AuthProvider")
  }
  return context
}
