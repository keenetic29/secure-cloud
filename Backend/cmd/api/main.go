package main

import (
	"log"
	
	"github.com/gin-gonic/gin"
	
	"server/config"
	"server/internal/controller/http"
	"server/internal/controller/middleware"
	"server/pkg/auth"
	"server/pkg/database"
	"server/pkg/yandex_disk"
	"server/internal/repository/postgres"
	"server/internal/usecase"
)

func main() {
	// Загрузка конфигурации
	cfg := config.Load()
	
	// Подключение к базе данных
	db, err := database.NewPostgresDB(
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Инициализация зависимостей
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	yandexDiskClient := yandex_disk.NewClient(
		cfg.YandexDisk.ClientID,
		cfg.YandexDisk.ClientSecret,
		cfg.YandexDisk.RedirectURI,
	)
	
	// Репозитории
	userRepo := postgres.NewUserRepository(db)
	fileRepo := postgres.NewFileRepository(db)
	
	// Use cases
	authUC := usecase.NewAuthUseCase(userRepo, jwtManager)
	storageUC := usecase.NewStorageUseCase(fileRepo, userRepo, yandexDiskClient)
	userUC := usecase.NewUserUseCase(userRepo)
	
	// Handlers
	authHandler := http.NewAuthHandler(authUC)
	storageHandler := http.NewStorageHandler(storageUC)
	userHandler := http.NewUserHandler(userUC)
	
	// Настройка роутера
	router := gin.Default()
	
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Public routes
	public := router.Group("/api/v1")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
	}
	
	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(authUC))
	{
		// Storage routes
		storageGroup := protected.Group("/storage")
		{
			storageGroup.GET("/yandex/auth-url", storageHandler.GetYandexAuthURL)
			storageGroup.POST("/yandex/callback", storageHandler.HandleYandexCallback)
			storageGroup.GET("/yandex/token", storageHandler.GetYandexToken) 
			storageGroup.GET("/files", storageHandler.GetFiles)
			storageGroup.GET("/files/:id", storageHandler.GetFileInfo)
			storageGroup.POST("/files/:id/decrypt-name", storageHandler.GetDecryptedFilename)
			storageGroup.POST("/upload", storageHandler.UploadFile)
			storageGroup.POST("/files/:id/download", storageHandler.DownloadFile)
			storageGroup.DELETE("/files/:id", storageHandler.DeleteFile)
		}
		
		// User routes
		userGroup := protected.Group("/user")
		{
			userGroup.GET("/profile", userHandler.GetProfile)
		}
	}
	
	// Запуск сервера
	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(router.Run(":" + cfg.ServerPort))
}