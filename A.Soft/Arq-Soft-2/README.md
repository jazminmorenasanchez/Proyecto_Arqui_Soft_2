# Arq-soft-2 – Sistema de Gestión de Actividades Deportivas

## 📖 Descripción
Sistema de microservicios para gestión de actividades deportivas con arquitectura distribuida que incluye:
- **users-api**: Gestión de usuarios y autenticación (MySQL)
- **activities-api**: Gestión de actividades y sesiones (MongoDB)
- **search-api**: Búsqueda avanzada con Solr y caché distribuido

## 🎯 Objetivo
Permitir a los usuarios buscar, inscribirse y gestionar actividades deportivas de manera eficiente, con un sistema de búsqueda avanzada y gestión de sesiones.

## 🏗️ Arquitectura
- **Base de datos**: MySQL (usuarios) + MongoDB (actividades)
- **Búsqueda**: Apache Solr
- **Caché**: Memcached + caché local
- **Mensajería**: RabbitMQ
- **APIs**: Go con Gin framework

## ⚙️ Configuración

### Variables de Entorno
Crea un archivo `.env` en la raíz del proyecto con las siguientes variables:

```bash
# Database Configuration
DB_PASSWORD=secret123
DB_NAME=sporthub_users
DB_PORT=3306

# MongoDB Configuration
MONGO_USER=admin
MONGO_PASSWORD=secret123
MONGO_PORT=27017

# Solr Configuration
SOLR_CORE=sporthub_core
SOLR_PORT=8983

# RabbitMQ Configuration
RABBITMQ_USER=admin
RABBITMQ_PASS=secret123
RABBITMQ_VHOST=/
RABBIT_PORT=5672
RABBIT_MGMT_PORT=15672

# Memcached Configuration
MEMCACHED_MEMORY=64
MEMCACHED_PORT=11211

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# API Ports
USERS_API_PORT=8081
ACTIVITIES_API_PORT=8082
SEARCH_API_PORT=8083
```

## 🚀 Levantar la infraestructura

```bash
# Levantar todos los servicios
docker compose up -d --build

# Ver logs
docker compose logs -f

# Parar servicios
docker compose down
```

## 🌐 Servicios Disponibles

### APIs
- **Users API**: http://localhost:8081
- **Activities API**: http://localhost:8082  
- **Search API**: http://localhost:8083

### Herramientas de Administración
- 🐬 **MySQL Workbench** → Base de datos de usuarios
- 🍃 **MongoDB Compass** → Base de datos de actividades
- 🔍 **Solr Admin UI** → http://localhost:8983/solr
- 🐇 **RabbitMQ Management** → http://localhost:15672

## 📁 Estructura del Proyecto

```
├── users-api/          # API de usuarios (MySQL + JWT)
├── activities-api/     # API de actividades (MongoDB + RabbitMQ)
├── search-api/         # API de búsqueda (Solr + Memcached)
├── frontend/           # Frontend React (pendiente)
├── deploy/            # Configuraciones Docker
└── docker-compose.yml # Orquestación de servicios
```

## 🔗 Endpoints Principales

### Users API (8081)
- `POST /auth/login` - Autenticación
- `POST /users` - Crear usuario
- `GET /users/:id` - Obtener usuario

### Activities API (8082)
- `GET /activities` - Listar actividades
- `POST /activities` - Crear actividad (admin)
- `GET /activities/:id/sessions` - Sesiones de actividad
- `POST /activities/:id/sessions` - Crear sesión (admin)

### Search API (8083)
- `GET /search?query=...` - Búsqueda avanzada
- `GET /health` - Health check

## 💻 Desarrollo

### Prerequisitos
- Docker & Docker Compose
- Go 1.22+
- Node.js (para frontend)

### Comandos Útiles

```bash
# Reconstruir un servicio específico
docker compose up -d --build users-api

# Ver logs de un servicio
docker compose logs -f activities-api

# Ejecutar tests
cd users-api && go test ./...
cd activities-api && go test ./...
cd search-api && go test ./...

# Limpiar volúmenes
docker compose down -v
```

## 📝 Notas de Desarrollo
- Las APIs están configuradas para desarrollo local
- Los servicios se comunican vía HTTP y RabbitMQ
- El sistema incluye health checks y graceful shutdown
- Configuración de CORS habilitada para desarrollo

---

## 📋 Resumen Técnico por API

### 🔐 Users API (Puerto 8081)

#### Variables de Entorno
| Variable | Descripción | Valor por Defecto |
|----------|-------------|-------------------|
| `APP_PORT` | Puerto del servidor HTTP | `8081` |
| `MYSQL_HOST` | Host de la base de datos MySQL | `mysql` |
| `MYSQL_PORT` | Puerto de MySQL | `3306` |
| `MYSQL_USER` | Usuario de MySQL | `root` |
| `MYSQL_PASSWORD` | Contraseña de MySQL | `secret` |
| `MYSQL_DB` | Nombre de la base de datos | `sporthub_users` |
| `JWT_SECRET` | Clave secreta para JWT | `change_me` |
| `JWT_EXP_MINUTES` | Tiempo de expiración del token (minutos) | `60` |

#### Arquitectura por Capas

**Controllers** (`internal/controllers/`)
- `auth.go`: Manejo de autenticación y login
- `users.go`: CRUD de usuarios
- `routes.go`: Registro de rutas HTTP
- `health.go`: Health check endpoint

**Services** (`internal/services/`)
- `users.go`: Lógica de negocio para usuarios
  - `Create()`: Crear nuevo usuario con hash de contraseña
  - `GetByID()`: Obtener usuario por ID
  - `Login()`: Autenticación y generación de JWT

**Repository** (`internal/repository/`)
- `users_mysql.go`: Acceso a datos MySQL
  - `Create()`: Insertar usuario en BD
  - `FindByID()`: Buscar usuario por ID
  - `FindByUsernameOrEmail()`: Buscar por username o email

**Utils** (`internal/utils/`)
- `bcrypt.go`: Hash y verificación de contraseñas
- `jwt.go`: Generación y validación de tokens JWT

---

### 🏃 Activities API (Puerto 8082)

#### Variables de Entorno
| Variable | Descripción | Valor por Defecto |
|----------|-------------|-------------------|
| `ACTIVITIES_PORT` | Puerto del servidor HTTP | `8082` |
| `MONGO_URI` | URI de conexión a MongoDB | `mongodb://localhost:27017` |
| `MONGO_DB` | Nombre de la base de datos | `sporthub` |
| `RABBITMQ_URL` | URL de conexión a RabbitMQ | `amqp://guest:guest@localhost:5672/` |
| `RABBITMQ_EXCHANGE` | Exchange de RabbitMQ | `activities.events` |
| `RABBITMQ_EXCHANGE_TYPE` | Tipo de exchange | `topic` |
| `USERS_API_BASE_URL` | URL base del Users API | `http://localhost:8081` |
| `JWT_SECRET` | Clave secreta para JWT | `change_me` |

#### Arquitectura por Capas

**Controllers** (`internal/controllers/`)
- `activities.go`: CRUD de actividades
- `sessions.go`: CRUD de sesiones de actividades
- `enrollments.go`: Gestión de inscripciones
- `cors.go`: Configuración CORS

**Services** (`internal/services/`)
- `activities.go`: Lógica de negocio para actividades
  - `Create()`: Crear actividad y publicar evento
  - `GetByID()`: Obtener actividad por ID
  - `Update()`: Actualizar actividad y publicar evento
  - `Delete()`: Eliminar actividad y publicar evento
  - `List()`: Listar actividades con paginación
- `sessions.go`: Lógica de negocio para sesiones
  - `CreateSession()`: Crear sesión de actividad
  - `GetSessionsByActivity()`: Obtener sesiones de una actividad
  - `GetSessionByID()`: Obtener sesión por ID
  - `UpdateSession()`: Actualizar sesión
  - `DeleteSession()`: Eliminar sesión
- `enrollments.go`: Lógica de negocio para inscripciones
  - `Enroll()`: Inscribir usuario en sesión
  - `Unenroll()`: Desinscribir usuario
  - `GetEnrollmentsByUser()`: Obtener inscripciones del usuario

**Repository** (`internal/repository/`)
- `activities_mongo.go`: Acceso a datos de actividades en MongoDB
- `sessions_mongo.go`: Acceso a datos de sesiones en MongoDB
- `enrollments_mongo.go`: Acceso a datos de inscripciones en MongoDB
- `helpers.go`: Funciones auxiliares para MongoDB

**Clients** (`internal/clients/`)
- `user_client.go`: Cliente HTTP para comunicarse con Users API
- `rabbitmq_client.go`: Cliente para publicar eventos en RabbitMQ

**Domain** (`internal/domain/`)
- `activity.go`: Estructura de datos para actividades
- `session.go`: Estructura de datos para sesiones
- `enrollment.go`: Estructura de datos para inscripciones
- `search_doc.go`: Estructura para documentos de búsqueda

---

### 🔍 Search API (Puerto 8083)

#### Variables de Entorno
| Variable | Descripción | Valor por Defecto |
|----------|-------------|-------------------|
| `SEARCH_API_PORT` | Puerto del servidor HTTP | `8083` |
| `SOLR_URL` | URL de Apache Solr | `http://solr:8983/solr/sporthub_core` |
| `MEMCACHED_ADDR` | Dirección de Memcached | `memcached:11211` |
| `CACHE_TTL_SECONDS` | TTL del caché en segundos | `60` |
| `RABBIT_URL` | URL de conexión a RabbitMQ | `amqp://guest:guest@rabbitmq:5672/` |
| `RABBIT_EXCHANGE` | Exchange de RabbitMQ | `activities.events` |
| `RABBIT_QUEUE` | Cola de RabbitMQ | `search_sync` |
| `RABBIT_ROUTING_KEY` | Routing key para RabbitMQ | `#` |
| `ACTIVITIES_API_BASE` | URL base del Activities API | `http://activities-api:8082` |

#### Arquitectura por Capas

**Controllers** (`internal/controllers/`)
- `search.go`: Endpoint de búsqueda
  - `Search()`: Búsqueda con parámetros (query, sport, site, date, sort, page, size)
- `routes.go`: Registro de rutas HTTP

**Services** (`internal/services/`)
- `search.go`: Lógica de búsqueda con caché
  - `Search()`: Búsqueda con caché local y distribuido
  - `Bust()`: Invalidar caché
  - `key()`: Generar clave de caché basada en parámetros

**Repository** (`internal/repository/`)
- `solr_repository.go`: Acceso a Apache Solr
  - `Search()`: Ejecutar consulta en Solr
- `cache_local.go`: Caché local en memoria
- `cache_memcached.go`: Caché distribuido con Memcached

**Consumers** (`internal/consumers/`)
- `rabbitmq_consumer.go`: Consumidor de eventos RabbitMQ
  - `Start()`: Iniciar consumidor de eventos
  - `handleEvent()`: Procesar eventos de sincronización

**Domain** (`internal/domain/`)
- `search_doc.go`: Estructura de documentos de búsqueda
- `search_doc.go`: Estructura de resultados de búsqueda

---

## 🔄 Flujo de Comunicación

1. **Users API** ↔ **Activities API**: Validación de usuarios vía HTTP
2. **Activities API** → **RabbitMQ**: Publicación de eventos (create/update/delete)
3. **RabbitMQ** → **Search API**: Consumo de eventos para sincronización
4. **Search API** ↔ **Activities API**: Obtención de datos completos vía HTTP
5. **Search API** ↔ **Solr**: Indexación y búsqueda
6. **Search API** ↔ **Memcached**: Caché distribuido