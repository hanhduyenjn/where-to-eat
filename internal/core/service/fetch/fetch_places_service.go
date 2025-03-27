package fetch

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"wheretoeat/internal/core/domain"
	"wheretoeat/internal/core/port"
)

type FetchPlacesService struct {
	placesRepo     port.PlacesRepository
	categoriesRepo port.CategoriesRepository
	apiAdapter     port.PlacesAPIAdapter

	// Global variables for concurrency control
	wg        sync.WaitGroup
	errChan   chan error
	semaphore chan struct{}
}

func NewFetchPlacesService(placesRepo port.PlacesRepository, categoriesRepo port.CategoriesRepository, apiAdapter port.PlacesAPIAdapter) *FetchPlacesService {
	return &FetchPlacesService{
		placesRepo:     placesRepo,
		categoriesRepo: categoriesRepo,
		apiAdapter:     apiAdapter,
		errChan:        make(chan error), // Initialize channel
		semaphore:      make(chan struct{}, 10), // Max 10 concurrent requests
	}
}

func (s *FetchPlacesService) FetchPlaces(ctx context.Context, minLat, maxLat, minLng, maxLng float64, category string) error {
	// Query config collection for types based on category
	types, err := s.categoriesRepo.GetCategoryTypes(ctx, category)
	if err != nil {
		return err
	}

	// Generate initial grid of circles
	circles := generateGrid(minLat, maxLat, minLng, maxLng)

	// Resize error channel to match the number of circles
	s.errChan = make(chan error, len(circles))

	// Process each circle in a goroutine
	for _, circle := range circles {
		s.wg.Add(1)
		go func(c domain.Circle) {
			defer s.wg.Done()
			// Acquire semaphore slot
			s.semaphore <- struct{}{}
			defer func() { <-s.semaphore }() // Release slot

			if err := s.fetchPlacesForCircle(ctx, category, types, c); err != nil {
				s.errChan <- err
			}
		}(circle)
	}

	// Wait for all goroutines to complete
	s.wg.Wait()
	close(s.errChan)

	// Check for any errors
	for err := range s.errChan {
		if err != nil {
			log.Printf("Error in goroutine: %v", err)
			// Continue despite errors; return nil unless critical
		}
	}

	return nil
}

// fetchPlacesForCircle handles fetching and subdividing for a single circle
func (s *FetchPlacesService) fetchPlacesForCircle(ctx context.Context, category string, types []string, circle domain.Circle) error {
	// Check if the circle already exists in the places collection
	if circle.Radius < domain.MinRadius {
		return nil
	}

	exists, err := s.placesRepo.CircleExists(ctx, category, circle)
	if err != nil {
		log.Printf("Failed to check if has fetched places for %s at (%f, %f, %fm): %v", category, circle.Lat, circle.Lng, circle.Radius, err)
		return err
	}
	
	// If circle exists, skip fetching but continue to check for subdivision
	if !exists {
		params := domain.RequestParams{
			Types:  types,
			Circle: circle,
		}
		time.Sleep(100 * time.Millisecond) // Rate limit API requests
		places, err := s.apiAdapter.FetchPlaces(ctx, params)
		if err != nil {
			log.Printf("Failed to fetch places for %s at (%f, %f, %fm): %v", category, circle.Lat, circle.Lng, circle.Radius, err)
			return err
		}
		log.Printf("Fetched %d places for %s at (%f, %f, %fm)", len(places), category, circle.Lat, circle.Lng, circle.Radius)

		// Save fetched places
		err = s.placesRepo.SavePlaces(ctx, category, circle, places)
		if err != nil {
			log.Printf("Failed to save places for %s at (%f, %f, %fm): %v", category, circle.Lat, circle.Lng, circle.Radius, err)
			return err
		}
		// Log all place names
		for _, place := range places {
			log.Printf("Place name: %s", place.(map[string]interface{})["displayName"].(map[string]interface{})["text"])
		}

		// If max results returned, subdivide and fetch again
		if len(places) < domain.MaxResultsPerReq {
			log.Printf("Fetched less than %d places for %s at (%f, %f, %fm), skipping subdivision", domain.MaxResultsPerReq, category, circle.Lat, circle.Lng, circle.Radius)
			return nil
		}
	} else {
		// check if the number of places < domain.MaxResultsPerReq, then skip
		numPlaces, err := s.placesRepo.GetNumPlaces(ctx, category, circle)
		if err != nil {
			log.Printf("Failed to get number of places for %s at (%f, %f, %fm): %v", category, circle.Lat, circle.Lng, circle.Radius, err)
		}
		if numPlaces < domain.MaxResultsPerReq {
			log.Printf("Fetched less than %d places for %s at (%f, %f, %fm), skipping fetch", domain.MaxResultsPerReq, category, circle.Lat, circle.Lng, circle.Radius)
			return nil
		}
	}
	newRadius := circle.Radius / 2

	subCircles := subdivideCircle(circle.Lat, circle.Lng, circle.Radius, newRadius)
	for _, subCircle := range subCircles {
		err := s.fetchPlacesForCircle(ctx, category, types, subCircle)
		if err != nil {
			log.Printf("Error in sub-circle for %s at (%f, %f, %fm): %v", category, subCircle.Lat, subCircle.Lng, subCircle.Radius, err)
			continue
		}
	}

	return nil
}

// generateGrid creates a grid of circles for the specified area with dynamic radius
func generateGrid(minLat, maxLat, minLng, maxLng float64) []domain.Circle {
	latDistance := (maxLat - minLat) * 111320                    // meters N-S
	lngDistance := (maxLng - minLng) * 111320 * domain.LngCosineAdjust // meters E-W (adjusted for latitude)

	initialRadius := math.Min(latDistance, lngDistance)
	if initialRadius < domain.MinRadius {
		initialRadius = domain.MinRadius
	}

	latStep := initialRadius * 2 * domain.OverlapFactor * domain.LatMeterToDegree
	lngStep := initialRadius * 2 * domain.OverlapFactor * domain.LatMeterToDegree / domain.LngCosineAdjust

	var circles []domain.Circle
	for lat := minLat; lat < maxLat + latStep/2; lat += latStep {
		for lng := minLng; lng < maxLng + lngStep/2; lng += lngStep {
			circles = append(circles, domain.Circle{Lat: lat, Lng: lng, Radius: initialRadius})
		}
	}	
	return circles
}

// subdivideCircle splits a circle into smaller overlapping circles
func subdivideCircle(lat, lng, radius, newRadius float64) []domain.Circle {
	latStep := newRadius * domain.LatMeterToDegree
	lngStep := newRadius * domain.LatMeterToDegree / domain.LngCosineAdjust

	return []domain.Circle{
		{Lat: lat - latStep, Lng: lng - lngStep, Radius: newRadius},
		{Lat: lat - latStep, Lng: lng + lngStep, Radius: newRadius},
		{Lat: lat + latStep, Lng: lng - lngStep, Radius: newRadius},
		{Lat: lat + latStep, Lng: lng + lngStep, Radius: newRadius},
	}
}