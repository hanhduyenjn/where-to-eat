package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"wheretoeat/internal/adapter/api"
	"wheretoeat/internal/adapter/repository/mongodb"
	"wheretoeat/internal/adapter/repository/postgres"
	"wheretoeat/internal/adapter/storage"
	"wheretoeat/internal/adapter/util"
	"wheretoeat/internal/core/fetch"
)

func main() {
	util.LoadEnv()

	if len(os.Args) < 2 {
		log.Fatal("Usage: main <job_name> [<args>...]")
	}

	jobName := os.Args[1]
	args := os.Args[2:]

	switch strings.ToLower(jobName) {
	case "run-fetch-places":
		if len(args) < 5 {
			log.Fatal("Usage: fetch_places <minLat> <maxLat> <minLng> <maxLng> <category>")
		}

		minLat, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			log.Fatalf("Invalid minLat: %v", err)
		}

		maxLat, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			log.Fatalf("Invalid maxLat: %v", err)
		}

		minLng, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			log.Fatalf("Invalid minLng: %v", err)
		}

		maxLng, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			log.Fatalf("Invalid maxLng: %v", err)
		}

		category := args[4]

		log.Printf("Fetching places for %s in area (%f-%f, %f-%f)", category, minLat, maxLat, minLng, maxLng)
		
		client, err := mongodb.NewMongoAdapter()
		if err != nil {
			log.Fatalf("Failed to initialize MongoDB client: %v", err)
		}
		defer client.Disconnect(context.TODO())

		placesRepo := mongodb.NewPlacesRepo(client)
		categoriesRepo := mongodb.NewCategoriesRepo(client)
		apiAdapter := api.NewNearbySearchAPI()

		service := fetch.NewFetchPlacesService(placesRepo, categoriesRepo, apiAdapter)
		err = service.FetchPlaces(context.TODO(), minLat, maxLat, minLng, maxLng, category)
		if err != nil {
			log.Fatalf("Failed to fetch places: %v", err)
		}

		log.Println("Fetch places job completed successfully.")

	case "run-fetch-images":
		if len(args) < 2 {
			log.Fatal("Usage: fetch_images <limit> <offset> (limit and offset of raw places responses, each response contains multiple places)")
		}

		limit, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid limit: %v", err)
		}

		offset, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("Invalid offset: %v", err)
		}

		log.Printf("Fetching images with limit %d and offset %d", limit, offset)

		// PostgreSQL connection
		pgDB, err := sqlx.Connect("postgres", os.Getenv("POSTGRES_URI"))
		if err != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", err)
		}
		defer pgDB.Close()

		localUploader := storage.NewLocalUploader()
		placesRepo := postgres.NewPlacesRepo(pgDB)

		service := fetch.NewFetchImagesService(localUploader, placesRepo)
		err = service.FetchImages(context.TODO(), limit, offset)
		if err != nil {
			log.Fatalf("Failed to fetch images: %v", err)
		}

		log.Println("Fetch images job completed successfully.")

	case "run-fetch-areas":
		if len(args) < 1 {
			log.Fatal("Usage: fetch_areas <query>")
		}

		query := args[0]
		log.Printf("Fetching areas for query '%s'", query)

		client, err := mongodb.NewMongoAdapter()
		if err != nil {
			log.Fatalf("Failed to initialize MongoDB client: %v", err)
		}
		defer client.Disconnect(context.TODO())

		areasRepo := mongodb.NewAreasRepo(client)
		apiAdapter := api.NewTextSearchAPI()

		service := fetch.NewFetchAreasService(areasRepo, apiAdapter)
		err = service.FetchAreas(context.TODO(), query)
		if err != nil {
			log.Fatalf("Failed to fetch areas: %v", err)
		}

		log.Println("Fetch areas job completed successfully.")

	default:
		log.Fatalf("Unknown job: %s", jobName)
	}
}