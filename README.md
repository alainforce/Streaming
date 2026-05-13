# Streaming App (Catus)

A comprehensive movie tracking and discovery platform featuring a modern React frontend and a high-performance Go backend.

## Project Structure

This repository is organized as a monorepo containing two main components:

- **Frontend (`/Catus`)**: A modern web application built with React, TypeScript, and Vite.
- **Backend (`/tmdb-api`)**: A robust RESTful API built with Go (Golang) and the Gin framework.

---

## 🚀 Features

### For Users
- **Discovery**: Browse trending movies and search with advanced filters (genre, year, sorting).
- **Movie Tracking**: Maintain a "Favorites" list and a "Watched" list with personal ratings and watch dates.
- **Account Management**: Secure Signup/Login, email/password updates, and account deletion.
- **Modern UI**: Fully responsive design with dark mode support and smooth transitions.

### For Administrators
- **User Management**: View all registered users, ban/unban accounts, or delete users.
- **Content Overview**: Monitor all movies saved by users across the platform.
- **Analytics Dashboard**: Real-time stats including total users, active vs. banned counts, and top-saved movies.

---

## 🛠️ Tech Stack

### Frontend (`/Catus`)
- **Framework**: React 19 (TypeScript)
- **Build Tool**: Vite
- **Styling**: Tailwind CSS 4 + Shadcn UI
- **State Management**: Zustand (with persistence)
- **Data Fetching**: TanStack Query (React Query) v5
- **Icons**: Lucide React
- **Notifications**: Sonner

### Backend (`/tmdb-api`)
- **Language**: Go 1.25+
- **Web Framework**: Gin Gonic
- **Database**: PostgreSQL (with connection pooling)
- **Caching**: Redis (Cache-aside pattern for TMDB data)
- **Security**: JWT Authentication + Password hashing (Bcrypt)
- **Rate Limiting**: Redis-backed rate limiting (General & Auth specific)
- **External API**: The Movie Database (TMDB) API integration

---

## 🛠️ Getting Started

### Prerequisites
- [Node.js](https://nodejs.org/) (v18+)
- [Go](https://golang.org/) (v1.21+)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- TMDB API Key (get one at [themoviedb.org](https://www.themoviedb.org/documentation/api))

### Backend Setup
1. Navigate to the backend directory:
   ```bash
   cd tmdb-api
   ```
2. Create a `.env` file based on `.env.example`:
   ```env
   TMDB_API_KEY=your_key_here
   SERVER_PORT=8080
   DATABASE_URL=postgres://user:pass@localhost:5432/db_name?sslmode=disable
   REDIS_ADDR=localhost:6379
   JWT_SECRET=your_32_char_secret_here
   ADMIN_EMAIL=admin@example.com
   ADMIN_PASSWORD=secure_password
   ```
3. Run the application (it will automatically handle migrations):
   ```bash
   go run cmd/api/main.go
   ```

### Frontend Setup
1. Navigate to the frontend directory:
   ```bash
   cd Catus
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Create a `.env` file:
   ```env
   VITE_API_BASE_URL=http://localhost:8080/api/v1
   VITE_TMDB_IMAGE_BASE_URL=https://image.tmdb.org/t/p
   ```
4. Start the development server:
   ```bash
   npm run dev
   ```

---

## 📁 Architecture Decisions (Backend)
- **Layered Architecture**: Strict unidirectional flow (Handler → Service → Repository).
- **Cache-aside Pattern**: Redis caches TMDB responses to reduce API latency and quota consumption.
- **Fail-Safe Design**: Security checks (Blacklist/Bans) "fail closed" (reject request if Redis is down), while performance optimizations "fail open" (serve live data if Redis is down).
- **UUIDs**: All user IDs are generated as UUID v4 in the database for enhanced security and non-predictability.
