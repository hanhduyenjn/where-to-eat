package fetch

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gocolly/colly/v2"

	"wheretoeat/internal/core/domain"
	"wheretoeat/internal/core/port"
	
)

// Initialize a shared Colly collector
var sharedCollector = colly.NewCollector(
	colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	colly.AllowURLRevisit(),
)

func init() {
	// Set request timeout and transport for the shared collector
	sharedCollector.SetRequestTimeout(30 * time.Second)
	sharedCollector.WithTransport(&http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	})
}

type FetchImagesService struct {
	photoUploader port.Uploader
	placesRepo    port.PlacesRepository
}

func NewFetchImagesService(photoUploader port.Uploader, placesRepo port.PlacesRepository) *FetchImagesService {
	return &FetchImagesService{
		photoUploader: photoUploader,
		placesRepo:    placesRepo,
	}
}

func (s *FetchImagesService) FetchImages(ctx context.Context, limit int, offset int) error {
	// Fetch places from the database
	places, err := s.placesRepo.GetPlaces(ctx, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch places: %v", err)
	}
	log.Printf("Fetched %d places", len(places))

	var wg sync.WaitGroup
	errChan := make(chan error, len(places))

	for _, place := range places {
		wg.Add(1)
		go func(place domain.Place) {
			log.Printf("Fetching images for %s", place.ID)
			defer wg.Done()
			for _, photo := range place.Photos {
				flagContentUri := photo.FlagContentUri
				imageURL, err := fetchImageURL(flagContentUri, sharedCollector)
				if err != nil {
					log.Printf("Failed to fetch image URL for %s: %v", flagContentUri, err)
					errChan <- err
					return
				}

				imagePath, err := downloadImage(imageURL)
				if err != nil {
					log.Printf("Failed to download image from %s: %v", imageURL, err)
					errChan <- err
					return
				}
				defer os.Remove(imagePath)
				// save to local\
				uuid := uuid.New().String()
				err = s.photoUploader.Upload(ctx, imagePath, "images/"+place.ID+"/"+uuid+".jpg")
				if err != nil {
					log.Printf("Failed to upload image to local for %s/%s: %v", place.ID, uuid, err)
					errChan <- err
					return
				}
				log.Printf("Uploaded image for %s/%s", place.ID, uuid)
			}
		}(place)
	}

	wg.Wait()
	close(errChan)

	// Return the first error if any occurred
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchImageURL(uri string, c *colly.Collector) (string, error) {
	log.Printf("Fetching image URL for %s", uri)

	var imageURL string
	c.OnHTML("img#preview-image", func(e *colly.HTMLElement) {
		// Extract the src attribute of the image tag with id="preview-image"
		imageURL = e.Attr("src")
		log.Printf("Found image URL: %s", imageURL)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error fetching %s: %v", uri, err)
	})

	err := c.Visit(uri)
	if err != nil {
		return "", err
	}

	if imageURL == "" {
		return "", fmt.Errorf("no image URL found in %s", uri)
	}
	log.Printf("Fetched image URL: %s", imageURL)

	return imageURL, nil
}

func downloadImage(imageURL string) (string, error) {
	log.Printf("Downloading image from %s", imageURL)
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image from %s: status %d", imageURL, resp.StatusCode)
	}

	tempFile, err := os.CreateTemp("", "image-*.jpg")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return "", err
	}
	log.Printf("Downloaded image to %s", tempFile.Name())
	return tempFile.Name(), nil
}