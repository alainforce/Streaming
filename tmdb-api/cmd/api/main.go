// cmd/api/main.go
package main

import (
	"context"
	"log"

	"github.com/alainforce/streaming/tmdb-api/internal/cache"
	"github.com/alainforce/streaming/tmdb-api/internal/config"
	"github.com/alainforce/streaming/tmdb-api/internal/database"
	"github.com/alainforce/streaming/tmdb-api/internal/handlers"
	"github.com/alainforce/streaming/tmdb-api/internal/middleware"
	"github.com/alainforce/streaming/tmdb-api/internal/repository"
	"github.com/alainforce/streaming/tmdb-api/internal/services"
	jwtpkg "github.com/alainforce/streaming/tmdb-api/pkg/jwt"
	"github.com/alainforce/streaming/tmdb-api/pkg/tmdb"
	"github.com/gin-gonic/gin"
)

func main() {
	// Configuration
	cfg := config.Load()

	// Database
	db, err := database.NewPostgresPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✅ Database connected and migrations applied")

	// Redis
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Redis connected")

	//Blacklist
	blacklist := cache.NewTokenBlacklist(redisClient)

	// JWT
	jwtManager := jwtpkg.NewManager(cfg.JWTSecret, cfg.JWTExpiryHours)

	// Rate limiters
	rateCfg := middleware.DefaultRateLimiterConfig()
	generalLimiter, _ := middleware.NewRateLimiter(redisClient, rateCfg.GeneralRate)
	authLimiter, _ := middleware.NewRateLimiter(redisClient, rateCfg.AuthRate)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	favoriteRepo := repository.NewFavoriteRepository(db)
	watchedRepo := repository.NewWatchedRepository(db)
	adminRepo := repository.NewAdminRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	// Services
	movieCache := cache.NewMovieCache(redisClient)
	tmdbClient := tmdb.NewClient(cfg.TMDBApiKey)
	movieService := services.NewMovieService(tmdbClient, movieCache)
	authService := services.NewAuthService(userRepo, jwtManager, blacklist)
	favoriteService := services.NewFavoriteService(favoriteRepo)
	watchedService := services.NewWatchedService(watchedRepo)
	adminService := services.NewAdminService(userRepo, adminRepo, blacklist)
	settingsService := services.NewSettingsService(settingsRepo)

	// Handlers
	movieHandler := handlers.NewMovieHandler(movieService)
	authHandler := handlers.NewAuthHandler(authService)
	favoriteHandler := handlers.NewFavoriteHandler(favoriteService)
	watchedHandler := handlers.NewWatchedHandler(watchedService)
	adminHandler := handlers.NewAdminHandler(adminService)
	settingsHandler := handlers.NewSettingsHandler(settingsService)

	// Seed admin
	if err := authService.SeedAdmin(context.Background(), cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}
	// Router setup
	router := gin.Default()
	router.Use(generalLimiter)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public
		movies := v1.Group("/movies")
		{
			movies.GET("/trending", movieHandler.GetTrending)
			movies.GET("/search", movieHandler.SearchMovies)
			movies.GET("/genres", movieHandler.GetGenres)
		}

		// Auth
		auth := v1.Group("/auth")
		auth.Use(authLimiter)
		{
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout",
				middleware.RequireAuth(jwtManager, blacklist),
				authHandler.Logout,
			)
			auth.DELETE("/account",
				middleware.RequireAuth(jwtManager, blacklist),
				authHandler.DeleteAccount,
			)
		}

		// Protected — favorites
		favorites := v1.Group("/favorites")
		favorites.Use(middleware.RequireAuth(jwtManager, blacklist))
		{
			favorites.POST("", favoriteHandler.AddFavorite)
			favorites.GET("", favoriteHandler.GetFavoriteByUser)
			favorites.DELETE("/:movie_id", favoriteHandler.DeleteFavorite)
		}

		// Protected — watched list
		watched := v1.Group("/watched")
		watched.Use(middleware.RequireAuth(jwtManager, blacklist))
		{
			watched.POST("", watchedHandler.AddWatched)
			watched.GET("", watchedHandler.GetWatched)
			watched.DELETE("/:movie_id", watchedHandler.DeleteWatched)
		}
		// Protected — user settings
		settings := v1.Group("/settings")
		settings.Use(middleware.RequireAuth(jwtManager, blacklist))
		{
			settings.GET("/profile", settingsHandler.GetProfile)
			settings.PATCH("/email", settingsHandler.UpdateEmail)
			settings.PATCH("/password", settingsHandler.UpdatePassword)
		}

		// Admin
		admin := v1.Group("/admin")
		admin.Use(middleware.RequireAuth(jwtManager, blacklist), middleware.RequireAdmin())
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.PATCH("/users/:id/ban", adminHandler.BanUser)
			admin.PATCH("/users/:id/unban", adminHandler.UnbanUser)
			admin.GET("/movies", adminHandler.GetAllSavedMovies)
			admin.GET("/stats", adminHandler.GetStats)
		}
	}

	// Start server
	log.Printf("🚀 Server running on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
