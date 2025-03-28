package fetch

import (
	"context"
	"io"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"wheretoeat/internal/core/domain"
	"wheretoeat/internal/core/port"
)

const (
	numWorkers    = 5  // Number of concurrent workers
	minDelay      = 100 * time.Millisecond
	maxDelay      = 1 * time.Second
)

// Shared Colly collector template (cloned per worker for thread safety)
var baseCollector = colly.NewCollector(
	colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	colly.AllowURLRevisit(),
)

func init() {
	baseCollector.SetRequestTimeout(30 * time.Second)
	baseCollector.WithTransport(&http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	})
}

type FetchImagesService struct {
	photoStorage  port.Storage
	placesRepo    port.PlacesRepository
}

func NewFetchImagesService(photoStorage port.Storage, placesRepo port.PlacesRepository) *FetchImagesService {
	return &FetchImagesService{
		photoStorage: photoStorage,
		placesRepo:    placesRepo,
	}
}

// FetchImages runs the pipeline to fetch and process images
func (s *FetchImagesService) FetchImages(ctx context.Context, limit, offset int) error {
	// Step 1: Producer - Fetch photos from Postgres
	photos, err := s.placesRepo.GetPhotos(ctx, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch photos: %v", err)
		return err
	}
	log.Printf("Fetched %d photos from Postgres", len(photos))

	// Step 2: Pipeline setup
	photoChan := make(chan domain.Photo, len(photos))
	resultChan := make(chan error, len(photos))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, photoChan, resultChan, &wg)
	}

	// Feed photos into the pipeline
	go func() {
		for _, photo := range photos {
			photoChan <- photo
		}
		close(photoChan)
	}()

	// Wait for workers to finish and close result channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Step 3: Consume results
	for err := range resultChan {
		if err != nil {
			return err // Return first error
		}
	}

	return nil
}

// worker processes photos in the pipeline
func (s *FetchImagesService) worker(ctx context.Context, photoChan <-chan domain.Photo, resultChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	// Clone the base collector for thread safety
	collector := baseCollector.Clone()

	for photo := range photoChan {
		log.Printf("Worker processing photo for place %s, photo %s", photo.PlaceID, photo.PhotoId)
		
		if photo.ImageUrl.Valid {
			log.Printf("Photo already uploaded to storage: %s", photo.ImageUrl.String)
			resultChan <- nil
			continue
		}

		// Fetch image URL with random delay
		imageURL, err := fetchImageURL(photo.FlagContentUri, collector)
		if err != nil {
			log.Printf("Failed to fetch image URL for %s: %v", photo.FlagContentUri, err)
			resultChan <- err
			continue
		}

		// Download image with random delay
		imagePath, err := downloadImage(imageURL)
		if err != nil {
			log.Printf("Failed to download image from %s: %v", imageURL, err)
			resultChan <- err
			continue
		}
		defer os.Remove(imagePath)

		// Upload image
		imgURL, err := s.photoStorage.Upload(ctx, imagePath, "images/"+photo.PlaceID+"/"+photo.PhotoId+".jpg")
		if err != nil {
			log.Printf("Failed to upload image for %s/%s: %v", photo.PlaceID, photo.PhotoId, err)
			resultChan <- err
			continue
		}

		// Update photo URL in Postgres
		err = s.placesRepo.UpdatePhotoURL(ctx, imgURL, photo.PlaceID, photo.PhotoId)
		if err != nil {
			log.Printf("Failed to update photo URL for %s/%s: %v", photo.PlaceID, photo.PhotoId, err)
			resultChan <- err
			continue
		}

		log.Printf("Successfully processed image for %s/%s, uploaded to %s", photo.PlaceID, photo.PhotoId, imgURL)
		resultChan <- nil
	}
}

// fetchImageURL fetches the image URL from the provided URI with a random delay
func fetchImageURL(uri string, c *colly.Collector) (string, error) {
	// Random delay to bypass rate limiting
	time.Sleep(minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay))))

	log.Printf("Fetching image URL for %s", uri)
	var imageURL string

	c.OnHTML("img#preview-image", func(e *colly.HTMLElement) {
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

	return imageURL, nil
}

// downloadImage downloads the image from the URL with a random delay
func downloadImage(imageURL string) (string, error) {
	// Random delay to bypass rate limiting
	time.Sleep(minDelay + time.Duration(rand.Int63n(int64(maxDelay-minDelay))))

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