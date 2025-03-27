package job

import (
	"context"
	"log"
	"strconv"

	"wheretoeat/internal/adapter/util"
	"wheretoeat/internal/core/service/fetch"
	"wheretoeat/internal/adapter/repository/mongodb"
	"wheretoeat/internal/adapter/api"
)

func RunFetchPlacesJob(args []string) {
	util.LoadEnv()

	if len(args) < 5 {
		log.Fatal("Usage: fetch_places_job <minLat> <maxLat> <minLng> <maxLng> <category>")
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
}
