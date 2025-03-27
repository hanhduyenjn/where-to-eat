package main

import (
	"log"
	"os"

	"wheretoeat/internal/adapter/job"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: job <command> [args...]")
	}

	command := os.Args[1]
	log.Printf("Running command: %s", os.Args)

	switch command {
		case "run-fetch-places":
			if len(os.Args) < 7 { // Command + 5 args
				log.Fatal("Usage: job run-fetch-places <minLat> <maxLat> <minLng> <maxLng> <category>")
			}
			job.RunFetchPlacesJob(os.Args[2:]) // Pass slice of args

		case "run-fetch-areas":
			if len(os.Args) < 3 { // Command + 1 arg
				log.Fatal("Usage: job run-fetch-areas <search>. E.g. job run-fetch-areas 'Quận Tân Phú'")
			}
			job.RunFetchAreasJob(os.Args[2]) // Pass single string arg
		case "run-fetch-images":
			if len(os.Args) < 4	{
				log.Fatal("Usage: job run-fetch-images <limit> <offset>.")
			}
			job.RunFetchImagesJob(os.Args[2:]) // Pass slice of args
		default:
			log.Fatalf("Unknown command: %s", command)
	}
}