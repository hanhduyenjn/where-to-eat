package job

import (
	"context"
	"log"
	"strconv"

	"wheretoeat/internal/adapter/util"
	"wheretoeat/internal/core/service/fetch"
	"wheretoeat/internal/adapter/storage"
	"wheretoeat/internal/adapter/repository/mongodb"
)

func RunFetchImagesJob(args []string) {
	util.LoadEnv()
	localUploader := storage.NewLocalUploader()

	client, err := mongodb.NewMongoAdapter()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	placesRepo := mongodb.NewPlacesRepo(client)

	service := fetch.NewFetchImagesService(localUploader, placesRepo)

	limit, offset := args[0], args[1]
	// convert to int
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		log.Fatalf("Invalid limit: %v", err)
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		log.Fatalf("Invalid offset: %v", err)
	}
	service.FetchImages(context.TODO(), limitInt, offsetInt)
	log.Printf("Fetched images with limit %s and offset %s", limit, offset)
}