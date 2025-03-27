package job

import (
	"context"
	"log"

	"wheretoeat/internal/adapter/util"
	"wheretoeat/internal/core/service/fetch"
	"wheretoeat/internal/adapter/repository/mongodb"
	"wheretoeat/internal/adapter/api"
)

func RunFetchAreasJob(query string) {
	util.LoadEnv()

	client, err := mongodb.NewMongoAdapter()
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer client.Disconnect(context.TODO())

	areasRepo := mongodb.NewAreasRepo(client)
	apiAdapter := api.NewTextSearchAPI()

	areaService := fetch.NewFetchAreasService(areasRepo, apiAdapter)
	err = areaService.FetchAreas(context.Background(), query)
	if err != nil {
		log.Printf("Error fetching areas: %v", err)
	} else {
		log.Printf("Finished fetching areas for query '%s'", query)
	}
}