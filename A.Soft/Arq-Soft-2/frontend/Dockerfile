# Dockerfile para Next.js Frontend
FROM node:20-alpine AS base

# Instalar dependencias solo cuando sea necesario
FROM base AS deps
RUN apk add --no-cache libc6-compat
WORKDIR /app

# Copiar archivos de dependencias
COPY package.json package-lock.json* pnpm-lock.yaml* ./
RUN \
  if [ -f package-lock.json ]; then \
    npm ci --legacy-peer-deps; \
  elif [ -f pnpm-lock.yaml ]; then \
    corepack enable pnpm && pnpm i --no-frozen-lockfile; \
  else \
    npm install --legacy-peer-deps; \
  fi

# Rebuild the source code only when needed
FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Variables de entorno para el build (se pueden sobrescribir en docker-compose)
ARG NEXT_PUBLIC_USERS_API_URL=http://localhost:8081
ARG NEXT_PUBLIC_ACTIVITIES_API_URL=http://localhost:8082
ARG NEXT_PUBLIC_SEARCH_API_URL=http://localhost:8083

ENV NEXT_PUBLIC_USERS_API_URL=$NEXT_PUBLIC_USERS_API_URL
ENV NEXT_PUBLIC_ACTIVITIES_API_URL=$NEXT_PUBLIC_ACTIVITIES_API_URL
ENV NEXT_PUBLIC_SEARCH_API_URL=$NEXT_PUBLIC_SEARCH_API_URL

# Build de Next.js
RUN npm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

# Copiar archivos necesarios desde el builder
COPY --from=builder --chown=nextjs:nodejs /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]

