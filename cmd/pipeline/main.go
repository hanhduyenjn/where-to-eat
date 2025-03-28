package main

import (
	"context"
	"log"
	"os"
	
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"wheretoeat/internal/core/pipeline/mongo2postgres"
	"wheretoeat/internal/adapter/repository/mongodb"
	"wheretoeat/internal/adapter/repository/postgres"
	"wheretoeat/internal/adapter/util"
)

func main() {
	util.LoadEnv()

	// MongoDB connection
	mongoClient, err := mongodb.NewMongoAdapter()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer mongoClient.Disconnect(context.TODO())

	// PostgreSQL connection
	pgDB, err := sqlx.Connect("postgres", os.Getenv("POSTGRES_URI"))
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	// Initialize repositories
	mongoRepo := mongodb.NewPlacesRepo(mongoClient)
	pgRepo := postgres.NewPlacesRepo(pgDB)

	// Run ETL pipeline
	etlService := pipeline.NewPlacesETLService(mongoRepo, pgRepo)
	err = etlService.Run(context.Background())
	if err != nil {
		log.Fatalf("ETL pipeline failed: %v", err)
	}
	log.Println("ETL pipeline completed successfully")
}