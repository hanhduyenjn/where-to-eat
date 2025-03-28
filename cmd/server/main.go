package main

import (
	"log"
	"os"

	"wheretoeat/internal/adapter/handler/get"
	"wheretoeat/internal/adapter/repository/postgres"
	"wheretoeat/internal/adapter/util"
	"wheretoeat/internal/core/service"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	util.LoadEnv()

	// PostgreSQL connection
	pgDB, err := sqlx.Connect("postgres", os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	// Repositories
	placesRepo := postgres.NewPlacesRepo(pgDB)

	// Services
	getPlacesService := service.NewGetPlacesService(placesRepo)

	// Handlers
	getPlacesHandler := get.NewGetPlacesHandler(getPlacesService)

	// Router
	r := gin.Default()
	r.GET("/nearby-places", getPlacesHandler.Handle)


	// Start server
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
